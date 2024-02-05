package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	customerrors "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/errors"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StoragePostgres struct {
	databaseURL string
	pool        *pgxpool.Pool
	userIDSeq   int
}

func NewPostgresStorage(config *config.Config) (*StoragePostgres, error) {
	pool, err := pgxpool.New(context.TODO(), config.DatabaseURL)
	if err != nil {
		return nil, err
	}
	storage := &StoragePostgres{
		databaseURL: config.DatabaseURL,
		pool:        pool,
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
			is_deleted bool
		);
	`
	_, err := storage.pool.Exec(context.TODO(), query)
	if err != nil {
		return err
	}
	return nil
}

func (storage *StoragePostgres) setUserIDSeq(databaseURL string) error {
	query := `
		select coalesce(max(created_by), 0) + 1 from urls
	`
	var userID int
	err := storage.pool.QueryRow(context.TODO(), query).Scan(&userID)
	storage.userIDSeq = userID
	return err
}

func (storage *StoragePostgres) Save(ctx context.Context, url models.URL) error {
	query := getInsertQuery()
	tr, err := storage.pool.Begin(ctx)
	if err != nil {
		return err
	}
	var shortURL string
	err = tr.QueryRow(ctx, query, url.ShortURL, url.OriginalURL, url.CreatedBy).Scan(&shortURL)
	if shortURL != "" {
		tr.Rollback(ctx)
		return customerrors.NewErrOriginalURLAlreadyExists(shortURL)
	}
	if err != nil {
		tr.Rollback(ctx)
		return err
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
		return err
	}
	queueQuery.QueryRow(func(row pgx.Row) error {
		var shortURL string
		row.Scan(&shortURL)
		if shortURL != "" {
			tr.Rollback(ctx)
			return customerrors.NewErrOriginalURLAlreadyExists(shortURL)
		}
		return nil
	})
	res := tr.SendBatch(ctx, batch)
	err = res.Close()
	if err != nil {
		tr.Rollback(ctx)
		return err
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
	query := "select id, short_url, original_url from urls where short_url = $1"
	var url models.URL
	err := storage.pool.QueryRow(ctx, query, shortURL).Scan(&url.ID, &url.ShortURL, &url.OriginalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customerrors.ErrOriginalURLNotFound
		}
		return nil, err
	}
	return &url, nil
}

func (storage *StoragePostgres) Ping(ctx context.Context) bool {
	return storage.pool.Ping(ctx) == nil
}

func (storage *StoragePostgres) GetUserID(context.Context) int {
	userID := storage.userIDSeq
	storage.userIDSeq++
	return userID
}

func (storage *StoragePostgres) FindByUser(ctx context.Context, userID int) ([]*models.URL, error) {
	query := "select id, short_url, original_url from urls where created_by = $1"
	rows, err := storage.pool.Query(ctx, query, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customerrors.ErrOriginalURLNotFound
		}
		return nil, err
	}
	urls := make([]*models.URL, 0)
	for rows.Next() {
		url := models.URL{}
		err := rows.Scan(&url.ID, &url.ShortURL, &url.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("unable to scan row: %w", err)
		}
		urls = append(urls, &url)
	}
	return urls, nil
}

func (storage *StoragePostgres) DeleteUrls(ctx context.Context, urls []models.URLToDelete, userID int) error {
	query := "update urls set is_deleted = true where short_url = $1"
	batch := &pgx.Batch{}
	var queueQuery *pgx.QueuedQuery
	for _, url := range urls {
		queueQuery = batch.Queue(query, string(url))
	}
	tr, err := storage.pool.Begin(ctx)
	if err != nil {
		tr.Rollback(ctx)
		return err
	}
	queueQuery.Exec(func(ct pgconn.CommandTag) error {
		return nil
	})
	res := tr.SendBatch(ctx, batch)
	err = res.Close()
	if err != nil {
		tr.Rollback(ctx)
		return err
	}
	tr.Commit(ctx)
	return nil
}
