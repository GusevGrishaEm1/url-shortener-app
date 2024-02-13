package postgres

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	customerrors "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/errors"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StoragePostgres struct {
	databaseURL string
	pool        *pgxpool.Pool
	userIDSeq   atomic.Int64
	config      *config.Config
}

func NewPostgresStorage(config *config.Config) (*StoragePostgres, error) {
	pool, err := pgxpool.New(context.TODO(), config.DatabaseURL)
	if err != nil {
		return nil, err
	}
	storage := &StoragePostgres{
		databaseURL: config.DatabaseURL,
		pool:        pool,
		config:      config,
	}
	if err := storage.createTables(config.DatabaseURL); err != nil {
		return nil, err
	}
	if err := storage.setUserIDSeq(config.DatabaseURL); err != nil {
		return nil, err
	}
	return storage, err
}

func (storage *StoragePostgres) createTables(databaseURL string) error {
	query := `
		create table if not exists urls (
			id serial primary key,
			created_by int,
			created_ts timestamp default now(),
			short_url varchar unique not null,
			original_url varchar unique not null,
			is_deleted bool default false
		);
	`
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30*time.Second))
	defer cancel()
	_, err := storage.pool.Exec(ctx, query)
	if err != nil {
		return err
	}
	return nil
}

func (storage *StoragePostgres) setUserIDSeq(databaseURL string) error {
	query := `
		select coalesce(max(created_by), 0) + 1 from urls
	`
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30*time.Second))
	defer cancel()
	var userID int
	err := storage.pool.QueryRow(ctx, query).Scan(&userID)
	storage.userIDSeq.Store(int64(userID))
	return err
}

func (storage *StoragePostgres) Save(ctx context.Context, url models.URL) error {
	query := getInsertQuery()
	tr, err := storage.pool.Begin(ctx)
	if err != nil {
		return customerrors.NewCustomErrorInternal(err)
	}
	var shortURL string
	err = tr.QueryRow(ctx, query, url.ShortURL, url.OriginalURL, url.CreatedBy).Scan(&shortURL)
	if shortURL != "" {
		tr.Rollback(ctx)
		err := customerrors.NewCustomError(errors.New("original url already exists"))
		err.Status = http.StatusConflict
		err.ShortURL = shortURL
		return err
	}
	if err != nil {
		tr.Rollback(ctx)
		return customerrors.NewCustomErrorInternal(err)
	}
	tr.Commit(ctx)
	return nil
}

func (storage *StoragePostgres) SaveBatch(ctx context.Context, urls []models.URL) error {
	query := getInsertQuery()
	batch := &pgx.Batch{}
	var queueQuery *pgx.QueuedQuery
	for _, el := range urls {
		queueQuery = batch.Queue(query, el.ShortURL, el.OriginalURL, el.CreatedBy)
	}
	tr, err := storage.pool.Begin(ctx)
	if err != nil {
		tr.Rollback(ctx)
		return customerrors.NewCustomErrorInternal(err)
	}
	queueQuery.QueryRow(func(row pgx.Row) error {
		var shortURL string
		row.Scan(&shortURL)
		if shortURL != "" {
			tr.Rollback(ctx)
			err := customerrors.NewCustomError(errors.New("original url already exists"))
			err.Status = http.StatusConflict
			err.ShortURL = shortURL
			return err
		}
		return nil
	})
	res := tr.SendBatch(ctx, batch)
	err = res.Close()
	if err != nil {
		tr.Rollback(ctx)
		return customerrors.NewCustomErrorInternal(err)
	}
	tr.Commit(ctx)
	return nil
}

func getInsertQuery() string {
	return `
	with new_id as (
		insert into urls(short_url, original_url, created_by) values($1, $2, $3)
		on conflict(original_url) do nothing
		returning id
	) select
		case when (select id from new_id) is null
			then (select short_url from urls where original_url = $2)
			else ''
		end as shortURL
	`
}

func (storage *StoragePostgres) FindByShortURL(ctx context.Context, shortURL string) (*models.URL, error) {
	query := "select id, short_url, original_url, is_deleted from urls where short_url = $1"
	var url models.URL
	err := storage.pool.QueryRow(ctx, query, shortURL).Scan(&url.ID, &url.ShortURL, &url.OriginalURL, &url.IsDeleted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customerrors.NewCustomErrorBadRequest(errors.New("original url isn't found"))
		}
		return nil, customerrors.NewCustomErrorInternal(err)
	}
	return &url, nil
}

func (storage *StoragePostgres) Ping(ctx context.Context) bool {
	return storage.pool.Ping(ctx) == nil
}

func (storage *StoragePostgres) GetUserID(context.Context) int {
	userID := storage.userIDSeq.Load()
	storage.userIDSeq.Add(1)
	return int(userID)
}

func (storage *StoragePostgres) FindByUser(ctx context.Context, userID int) ([]*models.URL, error) {
	query := "select id, short_url, original_url from urls where created_by = $1"
	rows, err := storage.pool.Query(ctx, query, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customerrors.NewCustomErrorBadRequest(errors.New("original url isn't found"))
		}
		return nil, customerrors.NewCustomErrorInternal(err)
	}
	urls := make([]*models.URL, 0)
	for rows.Next() {
		url := models.URL{}
		err := rows.Scan(&url.ID, &url.ShortURL, &url.OriginalURL)
		if err != nil {
			return nil, customerrors.NewCustomErrorInternal(err)
		}
		urls = append(urls, &url)
	}
	return urls, nil
}

func (storage *StoragePostgres) DeleteUrls(ctx context.Context, urls []models.URLToDelete) error {
	query := "update urls set is_deleted = true where short_url = $1 and created_by = $2"
	batch := &pgx.Batch{}
	for _, url := range urls {
		batch.Queue(query, url.ShortURL, url.UserID)
	}
	res := storage.pool.SendBatch(ctx, batch)
	err := res.Close()
	if err != nil {
		return customerrors.NewCustomErrorInternal(err)
	}
	return nil
}

func (storage *StoragePostgres) IsShortURLExists(ctx context.Context, shortURL string) (bool, error) {
	query := "select exists(select * from urls where short_url = $1) "
	var ok bool
	err := storage.pool.QueryRow(ctx, query, shortURL).Scan(&ok)
	if err != nil {
		return false, customerrors.NewCustomErrorInternal(err)
	}
	return ok, nil
}
