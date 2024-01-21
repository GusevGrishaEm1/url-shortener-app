package main

import (
	"log/slog"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/logger"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server"
)

func main() {
	var err error
	err = logger.Init(slog.LevelInfo)
	if err != nil {
		panic(err)
	}
	err = server.StartServer(config.GetDefault())
	if err != nil {
		panic(err)
	}
}
