package main

import (
	"context"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server"
)

func main() {
	ctx := context.Background()
	config := config.Init()
	if err := server.StartServer(ctx, config); err != nil {
		panic(err)
	}
}
