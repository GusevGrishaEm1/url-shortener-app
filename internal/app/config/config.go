package config

type Config struct {
	ServerURL       string
	BaseReturnURL   string
	FileStoragePath string
	DatabaseURL     string
}

func GetDefault() *Config {
	return &Config{
		ServerURL:     "localhost:8080",
		BaseReturnURL: "http://localhost:8080",
	}
}
