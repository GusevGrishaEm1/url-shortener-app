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
	FindByShortURL(context.Context, string) (*models.URL, error)
	Save(context.Context, models.URL) error
	SaveBatch(context.Context, []models.URL) error
	Ping(context.Context) bool
	FindByUser(context.Context, int) ([]*models.URL, error)
	GetUserID(context.Context) int
	DeleteUrls(context.Context, []models.URLToDelete, int) error
}

func GetStorageTypeByConfig(config *config.Config) StorageType {
	if config.DatabaseURL != "" {
		return StorageTypePostgres
	} else if config.FileStoragePath != "" {
		return StorageTypeFile
	} else {
		return StorageTypeInMemory
	}
}

func New(storageType StorageType, config *config.Config) (Storage, error) {
	switch storageType {
	case StorageTypeInMemory:
		return inmemory.NewInMemoryStorage(), nil
	case StorageTypeFile:
		return file.NewFileStorage(config)
	case StorageTypePostgres:
		return postgres.NewPostgresStorage(config)
	default:
		return nil, errors.New("unknown storage type")
	}
}
