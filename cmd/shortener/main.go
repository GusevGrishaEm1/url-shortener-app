package main

import (
	"context"
	"fmt"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server"
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
	if err := server.StartServer(ctx, config); err != nil {
		panic(err)
	}
}
