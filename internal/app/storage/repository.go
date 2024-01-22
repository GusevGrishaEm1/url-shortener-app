package repository

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/logger"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/jackc/pgx/v5"
)

var ErrOriginalURLNotFound = errors.New("original url isn't found")

type URLRepository interface {
	FindByShortURL(context.Context, string) (*models.URLInfo, error)
	Save(context.Context, models.URLInfo) error
	SaveBatch(context.Context, []models.URLInfo) error
	PingDB(context.Context) bool
}

type OriginalURLAlreadyExists struct {
	ShortURL string
}

func (err *OriginalURLAlreadyExists) Error() string {
	return "original url already exists"
}

func NewErrOriginalURLAlreadyExists(shortURL string) *OriginalURLAlreadyExists {
	return &OriginalURLAlreadyExists{
		ShortURL: shortURL,
	}
}

func New(config *config.Config) (URLRepository, error) {
	if config.DatabaseURL != "" {
		err := createTables(config.DatabaseURL)
		return &URLRepositoryPostgres{
			databaseURL: config.DatabaseURL,
		}, err
	} else if config.FileStoragePath != "" {
		repo := &URLRepositoryFile{
			filePath: config.FileStoragePath,
		}
		repo.uuidSeq = getUUIDSeqFromFile(repo)
		return repo, nil
	} else {
		return &URLRepositoryInMemory{
			urls: make(map[string]string),
		}, nil
	}
}

type URLRepositoryInMemory struct {
	urls map[string]string
}

func (r *URLRepositoryInMemory) FindByShortURL(_ context.Context, shortURL string) (*models.URLInfo, error) {
	originalURL, ok := r.urls[shortURL]
	if !ok {
		return nil, ErrOriginalURLNotFound
	}
	return &models.URLInfo{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}, nil
}

func (r *URLRepositoryInMemory) Save(_ context.Context, url models.URLInfo) error {
	r.urls[url.ShortURL] = url.OriginalURL
	return nil
}

func (r *URLRepositoryInMemory) PingDB(_ context.Context) bool {
	return false
}

func (r *URLRepositoryInMemory) SaveBatch(ctx context.Context, urls []models.URLInfo) error {
	for _, url := range urls {
		err := r.Save(ctx, url)
		if err != nil {
			return err
		}
	}
	return nil
}

type URLRepositoryFile struct {
	filePath string
	uuidSeq  int
}

func getUUIDSeqFromFile(repo *URLRepositoryFile) int {
	uuidSeq := 1
	urlsFromFile := loadFromFile(repo)
	for _, el := range urlsFromFile {
		if uuidSeq <= el.UUID {
			uuidSeq = el.UUID + 1
		}
	}
	return uuidSeq
}

func loadFromFile(repo *URLRepositoryFile) []models.URLInfo {
	var storageInfo models.URLInfo
	array := make([]models.URLInfo, 0)
	file, err := os.OpenFile(repo.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return array
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&storageInfo)
	if err != nil {
		logger.Logger.Warn(err.Error())
	}
	for err == nil {
		array = append(array, storageInfo)
		err = decoder.Decode(&storageInfo)
		if err != nil {
			logger.Logger.Warn(err.Error())
		}
	}
	return array
}

func (r *URLRepositoryFile) Save(_ context.Context, url models.URLInfo) error {
	file, err := os.OpenFile(r.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	url.UUID = r.uuidSeq
	r.uuidSeq++
	return encoder.Encode(url)
}

func (r *URLRepositoryFile) FindByShortURL(_ context.Context, shortURL string) (*models.URLInfo, error) {
	file, err := os.OpenFile(r.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(file)
	var url models.URLInfo
	err = decoder.Decode(&url)
	if url.ShortURL == shortURL {
		return &url, nil
	}
	if err != nil {
		logger.Logger.Warn(err.Error())
		if errors.Is(err, io.EOF) {
			return nil, ErrOriginalURLNotFound
		}
		return nil, err
	}
	for err == nil {
		err = decoder.Decode(&url)
		if url.ShortURL == shortURL {
			return &url, nil
		}
		if err != nil {
			logger.Logger.Warn(err.Error())
			if errors.Is(err, io.EOF) {
				return nil, ErrOriginalURLNotFound
			}
			return nil, err
		}
	}
	return nil, ErrOriginalURLNotFound
}

func (r *URLRepositoryFile) PingDB(_ context.Context) bool {
	return false
}

func (r *URLRepositoryFile) SaveBatch(ctx context.Context, urls []models.URLInfo) error {
	for _, url := range urls {
		err := r.Save(ctx, url)
		if err != nil {
			return err
		}
	}
	return nil
}

type URLRepositoryPostgres struct {
	databaseURL string
}

func createTables(databaseURL string) error {
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
	_, err = conn.Query(context.TODO(), query)
	if err != nil {
		return err
	}
	return nil
}

func (r *URLRepositoryPostgres) Save(ctx context.Context, url models.URLInfo) error {
	query := `
		with new_id as (
			insert into urls(short_url, original_url) values($1, $2)
			on conflict(original_url) do nothing
			returning id
		) select
			case when (select id from new_id) is null
				then (select short_url from urls where original_url = $2)
				else null
			end as short_url
	`
	conn, err := pgx.Connect(ctx, r.databaseURL)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	var shortURL string
	err = conn.QueryRow(ctx, query, url.ShortURL, url.OriginalURL).Scan(&shortURL)
	if shortURL != "" {
		return NewErrOriginalURLAlreadyExists(shortURL)
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *URLRepositoryPostgres) SaveBatch(ctx context.Context, urls []models.URLInfo) error {
	conn, err := pgx.Connect(ctx, r.databaseURL)
	if err != nil {
		return err
	}
	var tr pgx.Tx
	tr, err = conn.Begin(ctx)
	if err != nil {
		tr.Rollback(ctx)
		return err
	}
	for _, url := range urls {
		err := r.Save(ctx, url)
		if err != nil {
			tr.Rollback(ctx)
			return err
		}
	}
	tr.Commit(ctx)
	return nil
}

func (r *URLRepositoryPostgres) FindByShortURL(ctx context.Context, shortURL string) (*models.URLInfo, error) {
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
			return nil, ErrOriginalURLNotFound
		}
		return nil, err
	}
	return &url, nil
}

func (r *URLRepositoryPostgres) PingDB(ctx context.Context) bool {
	conn, err := pgx.Connect(ctx, r.databaseURL)
	if err != nil {
		return false
	}
	defer conn.Close(ctx)
	return conn.Ping(ctx) == nil
}
