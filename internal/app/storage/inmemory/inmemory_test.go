package inmemory

import (
	"context"
	"testing"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/stretchr/testify/assert"
)

func TestStorageInMemory_FindByShortURL(t *testing.T) {
	// Create a new instance of StorageInMemory
	storage := NewInMemoryStorage(config.Config{})

	// Create a test URL
	url := models.URL{
		ShortURL:    "abc",
		OriginalURL: "https://example.com",
	}

	// Save the URL to the storage
	err := storage.Save(context.Background(), url)
	assert.NoError(t, err)

	// Find the URL by short URL
	foundURL, err := storage.FindByShortURL(context.Background(), url.ShortURL)
	assert.NoError(t, err)
	assert.Equal(t, url, *foundURL)

	// Try to find a non-existent URL
	_, err = storage.FindByShortURL(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestStorageInMemory_Save(t *testing.T) {
	// Create a new instance of StorageInMemory
	storage := NewInMemoryStorage(config.Config{})

	// Create a test URL
	url := models.URL{
		ShortURL:    "abc",
		OriginalURL: "https://example.com",
	}

	// Save the URL to the storage
	err := storage.Save(context.Background(), url)
	assert.NoError(t, err)

	// Check if the URL was saved correctly
	foundURL, err := storage.FindByShortURL(context.Background(), url.ShortURL)
	assert.NoError(t, err)
	assert.Equal(t, url, *foundURL)
}

// Add more test functions for other methods in the StorageInMemory struct
