package config

type Config struct {
	ServerURL       string
	BaseReturnURL   string
	FileStoragePath string
}

func GetDefault() *Config {
	config := Config{
		ServerURL:       "localhost:8080",
		BaseReturnURL:   "http://localhost:8080",
		FileStoragePath: "/tmp/short-url-db.json",
	}
	return &config
}
