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

type URLRepository interface {
	FindByShortURL(shortURL string) (*models.URLInfo, error)
	Save(models.URLInfo) error
	PingDB() bool
}

var ErrOriginalURLNotFound = errors.New("original url isn't found")

func New(config *config.Config) (URLRepository, error) {
	if config.DatabaseURL != "" {
		logger.Logger.Info(config.DatabaseURL)
		// err := createTables(config.DatabaseURL)
		// return &URLRepositoryPostgres{
		// 	databaseURL: config.DatabaseURL,
		// }, err
		return &URLRepositoryInMemory{
			urls: make(map[string]string),
		}, nil
	} else if config.FileStoragePath != "" {
		file, err := os.OpenFile(config.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		repo := &URLRepositoryFile{
			file:    file,
			encoder: json.NewEncoder(file),
			decoder: json.NewDecoder(file),
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

func (r *URLRepositoryInMemory) FindByShortURL(shortURL string) (*models.URLInfo, error) {
	originalURL, ok := r.urls[shortURL]
	if !ok {
		return nil, ErrOriginalURLNotFound
	}
	return &models.URLInfo{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}, nil
}

func (r *URLRepositoryInMemory) Save(url models.URLInfo) error {
	r.urls[url.ShortURL] = url.OriginalURL
	return nil
}

func (r *URLRepositoryInMemory) PingDB() bool {
	return false
}

type URLRepositoryFile struct {
	file    *os.File
	encoder *json.Encoder
	decoder *json.Decoder
	uuidSeq int
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
	err := repo.decoder.Decode(&storageInfo)
	if err != nil {
		logger.Logger.Warn(err.Error())
	}
	for err == nil {
		array = append(array, storageInfo)
		err = repo.decoder.Decode(&storageInfo)
		if err != nil {
			logger.Logger.Warn(err.Error())
		}
	}
	return array
}

func (r *URLRepositoryFile) Save(url models.URLInfo) error {
	url.UUID = r.uuidSeq
	r.uuidSeq++
	return r.encoder.Encode(url)
}

func (r *URLRepositoryFile) FindByShortURL(shortURL string) (*models.URLInfo, error) {
	var url models.URLInfo
	err := r.decoder.Decode(&url)
	if err != nil {
		logger.Logger.Warn(err.Error())
		if errors.Is(err, io.EOF) {
			return nil, ErrOriginalURLNotFound
		}
		return nil, err
	}
	if url.ShortURL == shortURL {
		return &url, nil
	}
	for err == nil {
		err = r.decoder.Decode(&url)
		if err != nil {
			logger.Logger.Warn(err.Error())
			if errors.Is(err, io.EOF) {
				return nil, ErrOriginalURLNotFound
			}
			return nil, err
		}
		if url.ShortURL == shortURL {
			return &url, nil
		}
	}
	return nil, ErrOriginalURLNotFound
}

func (r *URLRepositoryFile) PingDB() bool {
	return false
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
			original_url varchar not null
		);
	`
	conn, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		return err
	}
	_, err = conn.Query(context.Background(), query)
	if err != nil {
		return err
	}
	return nil
}

func (r *URLRepositoryPostgres) Save(url models.URLInfo) error {
	query := `
		insert into urls(short_url, original_url) values($1,$2)
	`
	conn, err := pgx.Connect(context.Background(), r.databaseURL)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())
	_, err = conn.Query(context.Background(), query, url.ShortURL, url.OriginalURL)
	if err != nil {
		return err
	}
	return nil
}

func (r *URLRepositoryPostgres) FindByShortURL(shortURL string) (*models.URLInfo, error) {
	query := "select id, short_url, original_url from urls where short_url = $1"
	conn, err := pgx.Connect(context.Background(), r.databaseURL)
	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())
	var url models.URLInfo
	err = conn.QueryRow(context.Background(), query, shortURL).Scan(&url.UUID, &url.ShortURL, &url.OriginalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOriginalURLNotFound
		}
		return nil, err
	}
	return &url, nil
}

func (r *URLRepositoryPostgres) PingDB() bool {
	conn, err := pgx.Connect(context.Background(), r.databaseURL)
	if err != nil {
		return false
	}
	defer conn.Close(context.Background())
	return conn.Ping(context.Background()) == nil
}
