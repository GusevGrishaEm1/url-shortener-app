package main

import (
	"log/slog"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/logger"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server"
)

func main() {
	if err := logger.Init(slog.LevelInfo); err != nil {
		panic(err)
	}
	config := parseFlagsAndEnv()
	err := server.StartServer(config)
	if err != nil {
		panic(err)
	}
}
