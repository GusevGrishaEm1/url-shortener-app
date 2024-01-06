package main

import (
	"flag"
	"os"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
)

func parseFlagsAndEnv() *config.Config {
	config := new(config.Config)
	setFromFlags(config)
	setFromEnv(config)
	return config
}

func setFromEnv(config *config.Config) {
	if addr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		config.ServerURL = addr
	}
	if addr, ok := os.LookupEnv("BASE_URL"); ok {
		config.BaseReturnURL = addr
	}
	if path, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		config.FileStoragePath = path
	}
}

func setFromFlags(config *config.Config) {
	var serverURL string
	var baseReturnURL string
	var fileStoragePath string
	flag.StringVar(&serverURL, "a", "localhost:8080", "Net address host:port")
	flag.StringVar(&baseReturnURL, "b", "http://localhost:8080", "Return base address host:port")
	flag.StringVar(&fileStoragePath, "f", "/tmp/short-url-db.json", "File storage path")
	flag.Parse()
	config.ServerURL = serverURL
	config.BaseReturnURL = baseReturnURL
	config.FileStoragePath = fileStoragePath
}
