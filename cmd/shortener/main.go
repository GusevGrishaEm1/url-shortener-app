package main

import (
	"log/slog"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/logger"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server"
)

func main() {
	logger.Initialize(slog.LevelInfo)
	err := server.StartServer(parseFlagsAndEnv())
	if err != nil {
		panic(err)
	}
}
