package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	http_server "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server/http"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)
	ctx := context.Background()
	config, err := config.New()
	if err != nil {
		panic(err)
	}
	if err := http_server.StartServer(ctx, config); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		panic(err)
	}
}
