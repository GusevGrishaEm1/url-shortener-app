package grpc

import (
	"context"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/stretchr/testify/mock"
)

type MockShortenerService struct {
	mock.Mock
}

func (m *MockShortenerService) CreateShortURL(ctx context.Context, userInfo models.UserInfo, originalURL string) (string, error) {
	args := m.Called(ctx, userInfo, originalURL)
	return args.String(0), args.Error(1)
}

func (m *MockShortenerService) CreateBatchShortURL(ctx context.Context, userInfo models.UserInfo, arr []models.OriginalURLInfoBatch) ([]models.ShortURLInfoBatch, error) {
	args := m.Called(ctx, userInfo, arr)
	return args.Get(0).([]models.ShortURLInfoBatch), args.Error(1)
}

func (m *MockShortenerService) GetByShortURL(ctx context.Context, shortURL string) (string, error) {
	args := m.Called(ctx, shortURL)
	return args.String(0), args.Error(1)
}

func (m *MockShortenerService) PingStorage(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *MockShortenerService) GetUrlsByUser(ctx context.Context, userInfo models.UserInfo) ([]models.URLByUser, error) {
	args := m.Called(ctx, userInfo)
	return args.Get(0).([]models.URLByUser), args.Error(1)
}

func (m *MockShortenerService) DeleteUrlsByUser(ctx context.Context, userInfo models.UserInfo, urls []string) {
	m.Called(ctx, userInfo, urls)
}

func (m *MockShortenerService) GetStats(ctx context.Context) (models.Stats, error) {
	args := m.Called(ctx)
	return args.Get(0).(models.Stats), args.Error(1)
}
