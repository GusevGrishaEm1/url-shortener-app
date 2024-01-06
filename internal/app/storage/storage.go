package storage

import (
	"encoding/json"
	"os"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
)

type URLStorage interface {
	LoadFromStorage() []models.StorageURLInfo
	SaveToStorage(models.StorageURLInfo) error
}

func NewFileStorage(fileStoragePath string) (URLStorage, error) {
	file, err := os.OpenFile(fileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	enc := json.NewEncoder(file)
	dec := json.NewDecoder(file)
	return &URLStorageFileImpl{
		file:    file,
		encoder: enc,
		decoder: dec,
	}, err
}

type URLStorageFileImpl struct {
	file    *os.File
	encoder *json.Encoder
	decoder *json.Decoder
}

func (storage *URLStorageFileImpl) LoadFromStorage() []models.StorageURLInfo {
	var storageInfo models.StorageURLInfo
	array := make([]models.StorageURLInfo, 0)
	for err := storage.decoder.Decode(&storageInfo); err == nil; err = storage.decoder.Decode(&storageInfo) {
		array = append(array, storageInfo)
	}
	return array
}

func (storage *URLStorageFileImpl) SaveToStorage(info models.StorageURLInfo) error {
	err := storage.encoder.Encode(info)
	return err
}
