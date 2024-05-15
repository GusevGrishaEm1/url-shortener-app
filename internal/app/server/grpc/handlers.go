package grpc

import (
	"context"
	"errors"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/middlewares/security"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
)

// ShortenerService определяет методы для взаимодействия с сервисом сокращения URL.
type ShortenerService interface {
	// CreateShortURL создает сокращенный URL на основе исходного URL.
	CreateShortURL(ctx context.Context, userInfo models.UserInfo, shortURL string) (string, error)
	// CreateBatchShortURL создает несколько сокращенных URL на основе списка исходных URL.
	CreateBatchShortURL(ctx context.Context, userInfo models.UserInfo, arr []models.OriginalURLInfoBatch) ([]models.ShortURLInfoBatch, error)
	// GetByShortURL возвращает исходный URL по сокращенному URL.
	GetByShortURL(ctx context.Context, shortURL string) (string, error)
	// PingStorage проверяет доступность хранилища данных.
	PingStorage(ctx context.Context) bool
	// GetUrlsByUser возвращает список URL, созданных пользователем.
	GetUrlsByUser(ctx context.Context, userInfo models.UserInfo) ([]models.URLByUser, error)
	// DeleteUrlsByUser удаляет список URL, созданных пользователем.
	DeleteUrlsByUser(ctx context.Context, userInfo models.UserInfo, urls []string)
	// GetStats возвращающий в ответ объект статистики.
	GetStats(ctx context.Context) (models.Stats, error)
}

type shortenerHandler struct {
	service      ShortenerService
	serverConfig config.Config
	UnsafeShortenerServiceServer
}

func NewShortenerHandler(config config.Config, service ShortenerService) *shortenerHandler {
	return &shortenerHandler{
		service:      service,
		serverConfig: config,
	}
}

// CreateShortURL создает сокращенный URL на основе исходного URL.
func (s *shortenerHandler) CreateShortURL(ctx context.Context, in *CreateShortURLRequest) (*CreateShortURLResponse, error) {
	user, ok := ctx.Value(security.UserID).(models.UserInfo)
	if !ok {
		return nil, errors.New("invalid user id")
	}
	shortURL, err := s.service.CreateShortURL(ctx, user, in.URL)
	if err != nil {
		return nil, err
	}
	return &CreateShortURLResponse{
		URL: s.serverConfig.BaseReturnURL + "/" + shortURL,
	}, nil
}

// CreateBatchShortURL создает несколько сокращенных URL на основе списка исходных URL.
func (s *shortenerHandler) CreateBatchShortURL(ctx context.Context, in *CreateBatchShortURLRequest) (*CreateBatchShortURLResponse, error) {
	user, ok := ctx.Value(security.UserID).(models.UserInfo)
	if !ok {
		return nil, errors.New("invalid user id")
	}
	urlsOrig := make([]models.OriginalURLInfoBatch, 0, len(in.URLS))
	for _, url := range in.URLS {
		urlsOrig = append(urlsOrig, models.OriginalURLInfoBatch{OriginalURL: url.OriginalUrl, CorrelationID: url.CorrelationId})
	}
	urlsShort, err := s.service.CreateBatchShortURL(ctx, user, urlsOrig)
	if err != nil {
		return nil, err
	}
	outURLS := make([]*CreateBatchShortURLResponseItem, 0, len(urlsShort))
	for _, url := range urlsShort {
		outURLS = append(outURLS, &CreateBatchShortURLResponseItem{
			CorrelationId: url.CorrelationID,
			ShortUrl:      s.serverConfig.BaseReturnURL + "/" + url.ShortURL,
		})
	}
	return &CreateBatchShortURLResponse{URLS: outURLS}, nil
}

// GetByShortURL возвращает исходный URL по сокращенному URL.
func (s *shortenerHandler) GetByShortURL(ctx context.Context, in *GetByShortURLRequest) (*GetByShortURLResponse, error) {
	_, ok := ctx.Value(security.UserID).(models.UserInfo)
	if !ok {
		return nil, errors.New("invalid user id")
	}
	urlOrig, err := s.service.GetByShortURL(ctx, in.ShortURL)
	if err != nil {
		return nil, err
	}
	return &GetByShortURLResponse{OriginalUrl: urlOrig}, nil
}

// PingStorage проверяет доступность хранилища данных.
func (s *shortenerHandler) PingStorage(ctx context.Context, in *PingStorageRequest) (*PingStorageResponse, error) {
	_, ok := ctx.Value(security.UserID).(models.UserInfo)
	if !ok {
		return nil, errors.New("invalid user id")
	}
	ping := s.service.PingStorage(ctx)
	return &PingStorageResponse{Ping: ping}, nil
}

// GetUrlsByUser возвращает список URL, созданных пользователем.
func (s *shortenerHandler) GetUrlsByUser(ctx context.Context, in *GetUrlsByUserRequest) (*GetUrlsByUserResponse, error) {
	user, ok := ctx.Value(security.UserID).(models.UserInfo)
	if !ok {
		return nil, errors.New("invalid user id")
	}
	urls, err := s.service.GetUrlsByUser(ctx, user)
	if err != nil {
		return nil, err
	}
	urlsOut := make([]*GetUrlsByUserResponseItem, 0, len(urls))
	for _, url := range urls {
		urlsOut = append(urlsOut, &GetUrlsByUserResponseItem{
			ShortUrl:    s.serverConfig.BaseReturnURL + "/" + url.ShortURL,
			OriginalUrl: url.OriginalURL,
		})
	}
	return &GetUrlsByUserResponse{URLS: urlsOut}, nil
}

// DeleteUrlsByUser удаляет список URL, созданных пользователем.
func (s *shortenerHandler) DeleteUrlsByUser(ctx context.Context, in *DeleteUrlsByUserRequest) (*DeleteUrlsByUserResponse, error) {
	user, ok := ctx.Value(security.UserID).(models.UserInfo)
	if !ok {
		return nil, errors.New("invalid user id")
	}
	urlsToDelete := make([]string, 0, len(in.URLS))
	for _, url := range in.URLS {
		urlsToDelete = append(urlsToDelete, url.ShortURL)
	}
	s.service.DeleteUrlsByUser(ctx, user, urlsToDelete)
	return &DeleteUrlsByUserResponse{}, nil
}

// GetStats возвращающий в ответ объект статистики.
func (s *shortenerHandler) GetStats(ctx context.Context, in *GetStatsRequest) (*GetStatsResponse, error) {
	_, ok := ctx.Value(security.UserID).(models.UserInfo)
	if !ok {
		return nil, errors.New("invalid user id")
	}
	stats, err := s.service.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	return &GetStatsResponse{URLS: int32(stats.URLS), Users: int32(stats.Users)}, nil
}
