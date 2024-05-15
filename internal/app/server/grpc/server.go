package grpc

import (
	"context"
	"log"
	"net"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/service"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/storage"
	"google.golang.org/grpc"
)

func StartServer(ctx context.Context, config config.Config) error {
	listen, err := net.Listen("tcp", config.ServerURL)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	storage, err := storage.NewShortenerStorage(storage.GetStorageTypeByConfig(config), config)
	if err != nil {
		return err
	}
	service, err := service.NewShortenerService(ctx, config, storage)
	if err != nil {
		return err
	}
	handler := NewShortenerHandler(config, service)
	RegisterShortenerServiceServer(s, handler)

	log.Println("Starting gRPC server on port :50051")
	if err := s.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
	return nil
}
