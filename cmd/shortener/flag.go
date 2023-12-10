package main

import (
	"flag"
	"os"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server/config"
)

func parseFlags() *config.Config {
	config := new(config.Config)
	setFromEnv(config)
	setFromFlags(config)
	return config
}

func setFromEnv(config *config.Config) {
	if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
		config.SetServerURL(addr)
	}
	if addr := os.Getenv("BASE_URL"); addr != "" {
		config.SetBaseReturnURL(addr)
	}
}

func setFromFlags(config *config.Config) {
	var serverURL string
	var baseReturnURL string
	if config.GetServerURL() == "" {
		flag.StringVar(&serverURL, "a", "localhost:8080", "Net address host:port")
	}
	if config.GetBaseReturnURL() == "" {
		flag.StringVar(&baseReturnURL, "b", "http://localhost:8080", "Return base address host:port")
	}
	flag.Parse()
	if serverURL != "" {
		config.SetServerURL(serverURL)
	}
	if baseReturnURL != "" {
		config.SetBaseReturnURL(baseReturnURL)
	}
}
