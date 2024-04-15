package file

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/logger"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/stretchr/testify/assert"
)

func TestSave(t *testing.T) {
	logger.Init(slog.LevelInfo)
	// Create a test configuration
	config := config.Config{
		FileStoragePath: "test_data",
	}

	// Create a new instance of StorageFile
	storage, err := NewFileStorage(config)
	assert.NoError(t, err)

	// Create a test URL
	url := models.URL{
		ShortURL:    "abc",
		OriginalURL: "https://example.com",
	}

	// Save the URL to the storage
	err = storage.Save(context.Background(), url)
	assert.NoError(t, err)

	// Check if the URL was saved correctly
	foundURL, err := storage.FindByShortURL(context.Background(), url.ShortURL)
	url.ID = foundURL.ID
	assert.NoError(t, err)
	assert.Equal(t, url, *foundURL)
	err = os.Remove(config.FileStoragePath)
	assert.NoError(t, err)
}

func TestFindByShortURL(t *testing.T) {
	logger.Init(slog.LevelInfo)
	// Create a test configuration
	config := config.Config{
		FileStoragePath: "test_data",
	}

	// Create a new instance of StorageFile
	storage, err := NewFileStorage(config)
	assert.NoError(t, err)

	// Create a test URL
	url := models.URL{
		ShortURL:    "abc",
		OriginalURL: "https://example.com",
	}

	// Save the URL to the storage
	err = storage.Save(context.Background(), url)
	assert.NoError(t, err)

	// Find the URL by short URL
	foundURL, err := storage.FindByShortURL(context.Background(), url.ShortURL)
	assert.NoError(t, err)
	url.ID = foundURL.ID
	assert.Equal(t, url, *foundURL)

	// Try to find a non-existent URL
	_, err = storage.FindByShortURL(context.Background(), "nonexistent")
	assert.Error(t, err)
	err = os.Remove(config.FileStoragePath)
	assert.NoError(t, err)
}
