package main

import (
	"log"
	"net"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	grpc_server "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server/grpc"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// Запуск тестового сервера для тестирования
func startTestServer(config config.Config) error {
	listen, err := net.Listen("tcp", config.ServerURL)
	if err != nil {
		return err
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(grpc_server.UnarySecurityInterceptor))

	mockService := new(grpc_server.MockShortenerService)
	mockService.On("CreateShortURL", mock.Anything, mock.Anything, "http://example.com").Return("abc123", nil)
	mockService.On("GetByShortURL", mock.Anything, "abc123").Return("http://example.com", nil)

	handler := grpc_server.NewShortenerHandler(config, mockService)
	grpc_server.RegisterShortenerServiceServer(s, handler)

	log.Println("Starting gRPC server")
	if err := s.Serve(listen); err != nil {
		return err
	}
	return nil
}

// Запуск сервера для тестирования
func main() {
	config := config.Config{}
	config.ServerURL = "localhost:50051"
	config.BaseReturnURL = "http://short.url"
	err := startTestServer(config)
	if err != nil {
		panic(err)
	}
}
