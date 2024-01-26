package postgres

import (
	"context"
	"errors"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	customerrors "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/errors"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/jackc/pgx/v5"
)

func NewPostgresStorage(config *config.Config) (*StoragePostgres, error) {
	storage := &StoragePostgres{
		databaseURL: config.DatabaseURL,
	}
	err := storage.createTables(config.DatabaseURL)
	return storage, err
}

type StoragePostgres struct {
	databaseURL string
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
	conn, err := pgx.Connect(context.TODO(), databaseURL)
	if err != nil {
		return err
	}
	defer conn.Close(context.TODO())
	_, err = conn.Query(context.TODO(), query)
	if err != nil {
		return err
	}
	return nil
}

func (r *StoragePostgres) Save(ctx context.Context, url models.URLInfo) error {
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
	var tr pgx.Tx
	conn, err := pgx.Connect(ctx, r.databaseURL)
	if err != nil {
		return err
	}
	tr, err = conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	var shortURL string
	err = conn.QueryRow(ctx, query, url.ShortURL, url.OriginalURL).Scan(&shortURL)
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

func (r *StoragePostgres) SaveBatch(ctx context.Context, urls []models.URLInfo) error {
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
	conn, err := pgx.Connect(ctx, r.databaseURL)
	if err != nil {
		return err
	}
	batch := &pgx.Batch{}
	for _, el := range urls {
		batch.Queue(query, el.ShortURL, el.OriginalURL)
	}
	var tr pgx.Tx
	tr, err = conn.Begin(ctx)
	if err != nil {
		tr.Rollback(ctx)
		return err
	}
	res := conn.SendBatch(ctx, batch)
	defer res.Close()
	defer conn.Close(ctx)
	len := batch.Len()
	counter := 0
	for counter < len {
		var shortURL string
		res.QueryRow().Scan(&shortURL)
		if shortURL != "" {
			tr.Rollback(ctx)
			return customerrors.NewErrOriginalURLAlreadyExists(shortURL)
		}
		counter++
	}
	tr.Commit(ctx)
	return nil
}

func (r *StoragePostgres) FindByShortURL(ctx context.Context, shortURL string) (*models.URLInfo, error) {
	query := "select id, short_url, original_url from urls where short_url = $1"
	conn, err := pgx.Connect(ctx, r.databaseURL)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)
	var url models.URLInfo
	err = conn.QueryRow(ctx, query, shortURL).Scan(&url.UUID, &url.ShortURL, &url.OriginalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customerrors.ErrOriginalURLNotFound
		}
		return nil, err
	}
	return &url, nil
}

func (r *StoragePostgres) Ping(ctx context.Context) bool {
	conn, err := pgx.Connect(ctx, r.databaseURL)
	if err != nil {
		return false
	}
	defer conn.Close(ctx)
	return conn.Ping(ctx) == nil
}
