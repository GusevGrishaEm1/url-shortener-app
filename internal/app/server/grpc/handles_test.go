package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateShortURL(t *testing.T) {
	mockService := new(MockShortenerService)
	handler := NewShortenerHandler(config.Config{BaseReturnURL: "http://short.url"}, mockService)
	ctx := context.WithValue(context.Background(), models.UserID, 1)
	request := &CreateShortURLRequest{URL: "http://example.com"}

	mockService.On("CreateShortURL", ctx, models.UserInfo{UserID: 1}, "http://example.com").Return("abc123", nil)

	response, err := handler.CreateShortURL(ctx, request)
	assert.NoError(t, err)
	assert.Equal(t, "http://short.url/abc123", response.URL)

	mockService.AssertExpectations(t)
}

func TestCreateShortURLError(t *testing.T) {
	mockService := new(MockShortenerService)
	handler := NewShortenerHandler(config.Config{BaseReturnURL: "http://short.url"}, mockService)
	ctx := context.WithValue(context.Background(), models.UserID, 1)
	request := &CreateShortURLRequest{URL: "http://example.com"}

	mockService.On("CreateShortURL", ctx, models.UserInfo{UserID: 1}, "http://example.com").Return("", errors.New("service error"))

	_, err := handler.CreateShortURL(ctx, request)
	assert.Error(t, err)
	assert.Equal(t, "service error", err.Error())

	mockService.AssertExpectations(t)
}

func TestCreateBatchShortURL(t *testing.T) {
	mockService := new(MockShortenerService)
	handler := NewShortenerHandler(config.Config{BaseReturnURL: "http://short.url"}, mockService)
	ctx := context.WithValue(context.Background(), models.UserID, 1)
	request := &CreateBatchShortURLRequest{
		URLS: []*CreateBatchShortURLRequestItem{
			{OriginalUrl: "http://example1.com", CorrelationId: "1"},
			{OriginalUrl: "http://example2.com", CorrelationId: "2"},
		},
	}

	mockService.On("CreateBatchShortURL", ctx, models.UserInfo{UserID: 1}, []models.OriginalURLInfoBatch{
		{OriginalURL: "http://example1.com", CorrelationID: "1"},
		{OriginalURL: "http://example2.com", CorrelationID: "2"},
	}).Return([]models.ShortURLInfoBatch{
		{ShortURL: "short1", CorrelationID: "1"},
		{ShortURL: "short2", CorrelationID: "2"},
	}, nil)

	response, err := handler.CreateBatchShortURL(ctx, request)
	assert.NoError(t, err)
	assert.Len(t, response.URLS, 2)
	assert.Equal(t, "http://short.url/short1", response.URLS[0].ShortUrl)
	assert.Equal(t, "http://short.url/short2", response.URLS[1].ShortUrl)

	mockService.AssertExpectations(t)
}

func TestGetByShortURL(t *testing.T) {
	mockService := new(MockShortenerService)
	handler := NewShortenerHandler(config.Config{}, mockService)
	ctx := context.WithValue(context.Background(), models.UserID, 1)
	request := &GetByShortURLRequest{ShortURL: "short1"}

	mockService.On("GetByShortURL", ctx, "short1").Return("http://example1.com", nil)

	response, err := handler.GetByShortURL(ctx, request)
	assert.NoError(t, err)
	assert.Equal(t, "http://example1.com", response.OriginalURL)

	mockService.AssertExpectations(t)
}

func TestPingStorage(t *testing.T) {
	mockService := new(MockShortenerService)
	handler := NewShortenerHandler(config.Config{}, mockService)
	ctx := context.WithValue(context.Background(), models.UserID, 1)
	request := &PingStorageRequest{}

	mockService.On("PingStorage", ctx).Return(true)

	response, err := handler.PingStorage(ctx, request)
	assert.NoError(t, err)
	assert.True(t, response.Ping)

	mockService.AssertExpectations(t)
}

func TestGetUrlsByUser(t *testing.T) {
	mockService := new(MockShortenerService)
	handler := NewShortenerHandler(config.Config{BaseReturnURL: "http://short.url"}, mockService)
	ctx := context.WithValue(context.Background(), models.UserID, 1)
	request := &GetUrlsByUserRequest{}

	mockService.On("GetUrlsByUser", ctx, models.UserInfo{UserID: 1}).Return([]models.URLByUser{
		{ShortURL: "short1", OriginalURL: "http://example1.com"},
		{ShortURL: "short2", OriginalURL: "http://example2.com"},
	}, nil)

	response, err := handler.GetUrlsByUser(ctx, request)
	assert.NoError(t, err)
	assert.Len(t, response.URLS, 2)
	assert.Equal(t, "http://short.url/short1", response.URLS[0].ShortUrl)
	assert.Equal(t, "http://short.url/short2", response.URLS[1].ShortUrl)

	mockService.AssertExpectations(t)
}

func TestDeleteUrlsByUser(t *testing.T) {
	mockService := new(MockShortenerService)
	handler := NewShortenerHandler(config.Config{}, mockService)
	ctx := context.WithValue(context.Background(), models.UserID, 1)
	request := &DeleteUrlsByUserRequest{
		URLS: []*DeleteUrlsByUserRequestItem{
			{ShortURL: "short1"},
			{ShortURL: "short2"},
		},
	}

	mockService.On("DeleteUrlsByUser", ctx, models.UserInfo{UserID: 1}, []string{"short1", "short2"}).Return()

	_, err := handler.DeleteUrlsByUser(ctx, request)
	assert.NoError(t, err)

	mockService.AssertExpectations(t)
}

func TestGetStats(t *testing.T) {
	mockService := new(MockShortenerService)
	handler := NewShortenerHandler(config.Config{}, mockService)
	ctx := context.WithValue(context.Background(), models.UserID, 1)
	request := &GetStatsRequest{}

	mockService.On("GetStats", ctx).Return(models.Stats{URLS: 100, Users: 10}, nil)

	response, err := handler.GetStats(ctx, request)
	assert.NoError(t, err)
	assert.Equal(t, int32(100), response.URLS)
	assert.Equal(t, int32(10), response.Users)

	mockService.AssertExpectations(t)
}
