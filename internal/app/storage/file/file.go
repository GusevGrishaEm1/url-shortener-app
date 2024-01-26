package file

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	customerrors "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/errors"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/logger"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
)

func NewFileStorage(config *config.Config) (*StorageFile, error) {
	storage := &StorageFile{
		filePath: config.FileStoragePath,
	}
	storage.setUUIDSeqFromFile()
	return storage, nil
}

type StorageFile struct {
	filePath string
	uuidSeq  int
}

func (storage *StorageFile) setUUIDSeqFromFile() {
	uuidSeq := 1
	urlsFromFile := storage.loadFromFile()
	for _, el := range urlsFromFile {
		if uuidSeq <= el.UUID {
			uuidSeq = el.UUID + 1
		}
	}
	storage.uuidSeq = uuidSeq
}

func (storage *StorageFile) loadFromFile() []models.URLInfo {
	var urlInfo models.URLInfo
	array := make([]models.URLInfo, 0)
	file, err := os.OpenFile(storage.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return array
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&urlInfo)
	if err != nil {
		logger.Logger.Warn(err.Error())
	}
	for err == nil {
		array = append(array, urlInfo)
		err = decoder.Decode(&urlInfo)
		if err != nil {
			logger.Logger.Warn(err.Error())
		}
	}
	return array
}

func (storage *StorageFile) Save(_ context.Context, url models.URLInfo) error {
	file, err := os.OpenFile(storage.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	url.UUID = storage.uuidSeq
	storage.uuidSeq++
	return encoder.Encode(url)
}

func (storage *StorageFile) FindByShortURL(_ context.Context, shortURL string) (*models.URLInfo, error) {
	file, err := os.OpenFile(storage.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var url models.URLInfo
	err = decoder.Decode(&url)
	if url.ShortURL == shortURL {
		return &url, nil
	}
	if err != nil {
		logger.Logger.Warn(err.Error())
		if errors.Is(err, io.EOF) {
			return nil, customerrors.ErrOriginalURLNotFound
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
				return nil, customerrors.ErrOriginalURLNotFound
			}
			return nil, err
		}
	}
	return nil, customerrors.ErrOriginalURLNotFound
}

func (storage *StorageFile) Ping(_ context.Context) bool {
	file, err := os.OpenFile(storage.filePath, os.O_RDWR, 0666)
	if err != nil {
		return false
	}
	defer file.Close()
	return true
}

func (storage *StorageFile) SaveBatch(ctx context.Context, urls []models.URLInfo) error {
	for _, url := range urls {
		err := storage.Save(ctx, url)
		if err != nil {
			return err
		}
	}
	return nil
}
