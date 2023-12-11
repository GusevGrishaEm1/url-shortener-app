package service

import "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/util"

type ShortenerService interface {
	CreateShortURL(originalURL string) (string, bool)
	GetByShortURL(shortURL string) (string, bool)
}

type ShortenerServiceImpl struct {
	Urls map[string]string
}

func (service *ShortenerServiceImpl) CreateShortURL(originalURL string) (string, bool) {
	if originalURL == "" {
		return "", false
	} else {
		shortURL := util.GetShortURL()
		for _, ok := service.Urls[shortURL]; ok; {
			shortURL = util.GetShortURL()
		}
		service.Urls[shortURL] = originalURL
		return shortURL, true
	}
}

func (service *ShortenerServiceImpl) GetByShortURL(shortURL string) (string, bool) {
	originalURL, ok := service.Urls[shortURL]
	return originalURL, ok
}
