package storage

import (
	"encoding/json"
	"os"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/logger"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
)

type URLStorage interface {
	LoadFromStorage() []models.URLInfo
}

func New(fileStoragePath string) (URLStorage, error) {
	file, err := os.OpenFile(fileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &URLStorageFileImpl{
		file:    file,
		encoder: json.NewEncoder(file),
		decoder: json.NewDecoder(file),
	}, err
}

type URLStorageFileImpl struct {
	file    *os.File
	encoder *json.Encoder
	decoder *json.Decoder
}

func (storage *URLStorageFileImpl) LoadFromStorage() []models.URLInfo {
	var storageInfo *models.URLInfo
	array := make([]models.URLInfo, 0)
	err := storage.decoder.Decode(&storageInfo)
	if err != nil {
		logger.Logger.Warn(err.Error())
	}
	for err == nil {
		array = append(array, *storageInfo)
		err = storage.decoder.Decode(&storageInfo)
		if err != nil {
			logger.Logger.Warn(err.Error())
		}
	}
	defer storage.file.Close()
	return array
}
