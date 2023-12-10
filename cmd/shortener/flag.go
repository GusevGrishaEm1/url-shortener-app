package main

import (
	"flag"
	"os"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server/config"
)

func parseFlags() *config.Config {
	config := new(config.Config)
	setFromFlags(config)
	setFromEnv(config)
	return config
}

func setFromEnv(config *config.Config) {
	if addr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		config.SetServerURL(addr)
	}
	if addr, ok := os.LookupEnv("BASE_URL"); ok {
		config.SetBaseReturnURL(addr)
	}
}

func setFromFlags(config *config.Config) {
	var serverURL string
	var baseReturnURL string
	flag.StringVar(&serverURL, "a", "localhost:8080", "Net address host:port")
	flag.StringVar(&baseReturnURL, "b", "http://localhost:8080", "Return base address host:port")
	flag.Parse()
	config.SetServerURL(serverURL)
	config.SetBaseReturnURL(baseReturnURL)
}
