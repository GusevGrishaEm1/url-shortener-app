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

// StorageType определяет тип хранилища.
type StorageType string

// StorageTypeInMemory определяет тип хранилища в памяти.
// StorageTypePostgres определяет тип хранилища в базе данных
// StorageTypeFile определяет тип хранилища в файле.
const (
	StorageTypeInMemory StorageType = "inmemory"
	StorageTypeFile     StorageType = "file"
	StorageTypePostgres StorageType = "postgres"
)

// ShortenerStorage определяет методы для взаимодействия с хранилищем URL-ов.
type ShortenerStorage interface {
	// FindByShortURL находит оригинальный URL по сокращенному URL.
	FindByShortURL(ctx context.Context, shortURL string) (*models.URL, error)
	// Save сохраняет URL в хранилище.
	Save(ctx context.Context, url models.URL) error
	// SaveBatch сохраняет список URL в хранилище.
	SaveBatch(ctx context.Context, urls []models.URL) error
	// Ping проверяет доступность хранилища.
	Ping(ctx context.Context) bool
	// FindByUser находит URL, созданные конкретным пользователем.
	FindByUser(ctx context.Context, userID int) ([]models.URL, error)
	// GetUserID возвращает идентификатор пользователя из контекста.
	GetUserID(ctx context.Context) int
	// DeleteUrls удаляет список URL из хранилища.
	DeleteUrls(ctx context.Context, urls []models.URLToDelete) error
	// IsShortURLExists проверяет, существует ли указанный сокращенный URL в хранилище.
	IsShortURLExists(ctx context.Context, shortURL string) (bool, error)
	// GetStats возвращает статистику по хранилищу.
	GetStats(ctx context.Context) (models.Stats, error)
}

// GetStorageTypeByConfig возвращает тип хранилища на основе конфигурации.
func GetStorageTypeByConfig(config config.Config) StorageType {
	if config.DatabaseURL != "" {
		return StorageTypePostgres
	} else if config.FileStoragePath != "" {
		return StorageTypeFile
	} else {
		return StorageTypeInMemory
	}
}

// NewShortenerStorage создает новый экземпляр хранилища URL-ов в зависимости от указанного типа хранилища.
func NewShortenerStorage(storageType StorageType, config config.Config) (ShortenerStorage, error) {
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
