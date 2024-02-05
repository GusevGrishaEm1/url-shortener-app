package file

import (
	"context"
	"encoding/json"
	"os"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	customerrors "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/errors"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/logger"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
)

type StorageFile struct {
	filePath  string
	uuidSeq   int
	userIDSeq int
}

type URLInFile struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	CreatedBy   int    `json:"created_by"`
	IsDeleted   bool   `json:"is_deleted"`
}

func NewFileStorage(config *config.Config) (*StorageFile, error) {
	storage := &StorageFile{
		filePath: config.FileStoragePath,
	}
	storage.setSeqFromFile()
	return storage, nil
}

func (storage *StorageFile) setSeqFromFile() {
	uuidSeq := 1
	userIDSeq := 1
	urlsFromFile := storage.loadFromFile()
	for _, el := range urlsFromFile {
		if uuidSeq <= el.UUID {
			uuidSeq = el.UUID + 1
		}
		if userIDSeq <= el.CreatedBy {
			userIDSeq = el.CreatedBy + 1
		}
	}
	storage.uuidSeq = uuidSeq
	storage.userIDSeq = userIDSeq
}

func (storage *StorageFile) loadFromFile() []URLInFile {
	array := make([]URLInFile, 0)
	file, err := os.OpenFile(storage.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return array
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	urlInFile := URLInFile{}
	err = decoder.Decode(&urlInFile)
	if err != nil {
		logger.Logger.Warn(err.Error())
	}
	for err == nil {
		array = append(array, urlInFile)
		err = decoder.Decode(&urlInFile)
		if err != nil {
			logger.Logger.Warn(err.Error())
		}
	}
	return array
}

func (storage *StorageFile) Save(ctx context.Context, url models.URL) error {
	file, err := os.OpenFile(storage.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	urlInFile := &URLInFile{
		UUID:        storage.uuidSeq,
		ShortURL:    url.ShortURL,
		OriginalURL: url.OriginalURL,
		CreatedBy:   url.CreatedBy,
	}
	storage.uuidSeq++
	return encoder.Encode(urlInFile)
}

func (storage *StorageFile) FindByShortURL(_ context.Context, shortURL string) (*models.URL, error) {
	urlsInFile := storage.loadFromFile()
	for _, el := range urlsInFile {
		if el.ShortURL == shortURL {
			url := &models.URL{
				ID:          el.UUID,
				ShortURL:    el.ShortURL,
				OriginalURL: el.OriginalURL,
				CreatedBy:   el.CreatedBy,
				IsDeleted:   el.IsDeleted,
			}
			return url, nil
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

func (storage *StorageFile) SaveBatch(ctx context.Context, urls []models.URL) error {
	for _, url := range urls {
		err := storage.Save(ctx, url)
		if err != nil {
			return err
		}
	}
	return nil
}

func (storage *StorageFile) GetUserID(context.Context) int {
	userID := storage.userIDSeq
	storage.userIDSeq++
	return userID
}

func (storage *StorageFile) FindByUser(ctx context.Context, userID int) ([]*models.URL, error) {
	urlsInFile := storage.loadFromFile()
	urls := make([]*models.URL, 0)
	for _, el := range urlsInFile {
		if el.CreatedBy == userID {
			urls = append(urls, &models.URL{
				ID:          el.UUID,
				ShortURL:    el.ShortURL,
				OriginalURL: el.OriginalURL,
				CreatedBy:   el.CreatedBy,
			})
		}
	}
	if len(urls) > 0 {
		return urls, nil
	}
	return nil, customerrors.ErrOriginalURLNotFound
}

func (storage *StorageFile) DeleteUrls(_ context.Context, urls []models.URLToDelete, userID int) error {
	urlsFromFile := storage.loadFromFile()
	for _, url := range urls {
		for i, urlFromFile := range urlsFromFile {
			if string(url) == urlFromFile.ShortURL && urlFromFile.CreatedBy == userID {
				urlsFromFile[i].IsDeleted = true
			}
		}
	}
	file, err := os.OpenFile(storage.filePath, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, url := range urlsFromFile {
		encoder := json.NewEncoder(file)
		encoder.Encode(url)
	}
	return nil
}
