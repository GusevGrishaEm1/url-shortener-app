package postgres

import (
	"context"
	"errors"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	customerrors "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/errors"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresStorage(config *config.Config) (*StoragePostgres, error) {
	pool, err := pgxpool.New(context.TODO(), config.DatabaseURL)
	if err != nil {
		return nil, err
	}
	storage := &StoragePostgres{
		databaseURL: config.DatabaseURL,
		pool:        pool,
	}
	err = storage.createTables(config.DatabaseURL)
	return storage, err
}

type StoragePostgres struct {
	databaseURL string
	pool        *pgxpool.Pool
}

func (storage *StoragePostgres) createTables(databaseURL string) error {
	query := `
		create table if not exists urls (
			id serial primary key,
			create_ts timestamp default now(),
			short_url varchar unique not null,
			original_url varchar unique not null
		);
	`
	_, err := storage.pool.Exec(context.TODO(), query)
	if err != nil {
		return err
	}
	return nil
}

func (storage *StoragePostgres) Save(ctx context.Context, url models.URLInfo) error {
	query := `
		with new_id as (
			insert into urls(short_url, original_url) values($1, $2)
			on conflict(original_url) do nothing
			returning id
		) select
			case when (select id from new_id) is null
				then (select short_url from urls where original_url = $2)
				else ''
			end as shortURL
	`
	tr, err := storage.pool.Begin(ctx)
	if err != nil {
		return err
	}
	var shortURL string
	err = tr.QueryRow(ctx, query, url.ShortURL, url.OriginalURL).Scan(&shortURL)
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

func (storage *StoragePostgres) SaveBatch(ctx context.Context, urls []models.URLInfo) error {
	query := `
	with new_id as (
		insert into urls(short_url, original_url) values($1, $2)
		on conflict(original_url) do nothing
		returning id
	) select
		case when (select id from new_id) is null
			then (select short_url from urls where original_url = $2)
			else ''
		end as shortURL
	`
	batch := &pgx.Batch{}
	var queueQuery *pgx.QueuedQuery
	for _, el := range urls {
		queueQuery = batch.Queue(query, el.ShortURL, el.OriginalURL)
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

func (storage *StoragePostgres) FindByShortURL(ctx context.Context, shortURL string) (*models.URLInfo, error) {
	query := "select id, short_url, original_url from urls where short_url = $1"
	var url models.URLInfo
	err := storage.pool.QueryRow(ctx, query, shortURL).Scan(&url.UUID, &url.ShortURL, &url.OriginalURL)
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
