package main

import (
	"flag"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server/config"
)

func parseFlags() *config.Config {
	config := new(config.Config)
	var serverURL string
	var baseReturnURL string
	flag.StringVar(&serverURL, "a", "localhost:8080", "Net address host:port")
	flag.StringVar(&baseReturnURL, "b", "http://localhost:8080/", "Return base address host:port")
	flag.Parse()
	config.SetServerURL(serverURL)
	config.SetBaseReturnURLURL(baseReturnURL)
	return config
}
