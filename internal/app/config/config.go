package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerURL       string
	BaseReturnURL   string
	FileStoragePath string
	DatabaseURL     string
}

func GetDefault() Config {
	return Config{
		ServerURL:     "localhost:8080",
		BaseReturnURL: "http://localhost:8080",
	}
}

func GetDefaultWithTestDB() Config {
	return Config{
        ServerURL:     "localhost:8080",
        BaseReturnURL: "http://localhost:8080",
        DatabaseURL: "postgres://test:test@localhost:5432/test?sslmode=disable",
    }
}

func Init() Config {
	config := Config{}
	config = configFromFlags(config)
	config = configFromEnv(config)
	return config
}

func configFromEnv(config Config) Config {
	if addr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		config.ServerURL = addr
	}
	if addr, ok := os.LookupEnv("BASE_URL"); ok {
		config.BaseReturnURL = addr
	}
	if path, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		config.FileStoragePath = path
	}
	if url, ok := os.LookupEnv("DATABASE_DSN"); ok {
		config.DatabaseURL = url
	}
	return config
}

func configFromFlags(config Config) Config {
	flag.StringVar(&config.ServerURL, "a", "localhost:8080", "Net address host:port")
	flag.StringVar(&config.BaseReturnURL, "b", "http://localhost:8080", "Return base address host:port")
	flag.StringVar(&config.FileStoragePath, "f", "", "File storage path")
	flag.StringVar(&config.DatabaseURL, "d", "", "Database URL")
	flag.Parse()
	return config
}