package storage

import (
	"context"
	"errors"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/storage/file"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/storage/inmemory"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/storage/postgres"
)

type StorageType string

const (
	StorageTypeInMemory StorageType = "inmemory"
	StorageTypeFile     StorageType = "file"
	StorageTypePostgres StorageType = "postgres"
)

type Storage interface {
	FindByShortURL(ctx context.Context, shortURL string) (*models.URL, error)
	Save(ctx context.Context, url models.URL) error
	SaveBatch(ctx context.Context, urls []models.URL) error
	Ping(ctx context.Context) bool
	FindByUser(ctx context.Context, userID int) ([]models.URL, error)
	GetUserID(ctx context.Context) int
	DeleteUrls(ctx context.Context, urls []models.URLToDelete) error
	IsShortURLExists(ctx context.Context, shortURL string) (bool, error)
}

func GetStorageTypeByConfig(config config.Config) StorageType {
	if config.DatabaseURL != "" {
		return StorageTypePostgres
	} else if config.FileStoragePath != "" {
		return StorageTypeFile
	} else {
		return StorageTypeInMemory
	}
}

func New(storageType StorageType, config config.Config) (Storage, error) {
	switch storageType {
	case StorageTypeInMemory:
		return inmemory.NewInMemoryStorage(config), nil
	case StorageTypeFile:
		return file.NewFileStorage(config)
	case StorageTypePostgres:
		return postgres.NewPostgresStorage(config)
	default:
		return nil, errors.New("unknown storage type")
	}
}
