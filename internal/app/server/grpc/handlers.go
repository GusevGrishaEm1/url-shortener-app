package grpc

import (
	"context"
	"errors"
	"strconv"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

func UnarySecurityInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("user ID not found in metadata")
	}
	userIDStr := md.Get(string(models.UserID))
	if len(userIDStr) == 0 {
		return nil, errors.New("user ID not found in metadata")
	}
	userID, err := strconv.Atoi(userIDStr[0])
	if err != nil {
		return nil, err
	}
	resp, err = handler(context.WithValue(ctx, models.UserID, userID), req)
	return resp, err
}

// CreateShortURL создает сокращенный URL на основе исходного URL.
func (s *shortenerHandler) CreateShortURL(ctx context.Context, in *CreateShortURLRequest) (*CreateShortURLResponse, error) {
	userID, ok := ctx.Value(models.UserID).(int)
	if !ok {
		return nil, errors.New("invalid user id")
	}
	shortURL, err := s.service.CreateShortURL(ctx, models.UserInfo{UserID: userID}, in.URL)
	if err != nil {
		return nil, err
	}
	return &CreateShortURLResponse{
		URL: s.serverConfig.BaseReturnURL + "/" + shortURL,
	}, nil
}

// CreateBatchShortURL создает несколько сокращенных URL на основе списка исходных URL.
func (s *shortenerHandler) CreateBatchShortURL(ctx context.Context, in *CreateBatchShortURLRequest) (*CreateBatchShortURLResponse, error) {
	userID, ok := ctx.Value(models.UserID).(int)
	if !ok {
		return nil, errors.New("invalid user id")
	}
	urlsOrig := make([]models.OriginalURLInfoBatch, 0, len(in.URLS))
	for _, url := range in.URLS {
		urlsOrig = append(urlsOrig, models.OriginalURLInfoBatch{OriginalURL: url.OriginalUrl, CorrelationID: url.CorrelationId})
	}
	urlsShort, err := s.service.CreateBatchShortURL(ctx, models.UserInfo{UserID: userID}, urlsOrig)
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
	_, ok := ctx.Value(models.UserID).(int)
	if !ok {
		return nil, errors.New("invalid user id")
	}
	urlOrig, err := s.service.GetByShortURL(ctx, in.ShortURL)
	if err != nil {
		return nil, err
	}
	return &GetByShortURLResponse{OriginalURL: urlOrig}, nil
}

// PingStorage проверяет доступность хранилища данных.
func (s *shortenerHandler) PingStorage(ctx context.Context, in *PingStorageRequest) (*PingStorageResponse, error) {
	_, ok := ctx.Value(models.UserID).(int)
	if !ok {
		return nil, errors.New("invalid user id")
	}
	ping := s.service.PingStorage(ctx)
	return &PingStorageResponse{Ping: ping}, nil
}

// GetUrlsByUser возвращает список URL, созданных пользователем.
func (s *shortenerHandler) GetUrlsByUser(ctx context.Context, in *GetUrlsByUserRequest) (*GetUrlsByUserResponse, error) {
	userID, ok := ctx.Value(models.UserID).(int)
	if !ok {
		return nil, errors.New("invalid user id")
	}
	urls, err := s.service.GetUrlsByUser(ctx, models.UserInfo{UserID: userID})
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
	userID, ok := ctx.Value(models.UserID).(int)
	if !ok {
		return nil, errors.New("invalid user id")
	}
	urlsToDelete := make([]string, 0, len(in.URLS))
	for _, url := range in.URLS {
		urlsToDelete = append(urlsToDelete, url.ShortURL)
	}
	s.service.DeleteUrlsByUser(ctx, models.UserInfo{UserID: userID}, urlsToDelete)
	return &DeleteUrlsByUserResponse{}, nil
}

// GetStats возвращающий в ответ объект статистики.
func (s *shortenerHandler) GetStats(ctx context.Context, in *GetStatsRequest) (*GetStatsResponse, error) {
	_, ok := ctx.Value(models.UserID).(int)
	if !ok {
		return nil, errors.New("invalid user id")
	}
	stats, err := s.service.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	return &GetStatsResponse{URLS: int32(stats.URLS), Users: int32(stats.Users)}, nil
}
