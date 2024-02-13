package main

import (
	"context"
	"log/slog"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/logger"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server"
)

func main() {
	if err := logger.Init(slog.LevelInfo); err != nil {
		panic(err)
	}
	ctx := context.Background()
	if err := server.StartServer(ctx, parseFlagsAndEnv()); err != nil {
		panic(err)
	}
}
