// Пакет config предоставляет функциональность для инициализации и конфигурации приложения.
package config

import (
	"encoding/json"
	"flag"
	"os"
)

// Структура Config представляет собой настройки конфигурации для приложения.
// Она включает в себя поля для URL сервера, базового URL возврата, пути к файловому хранилищу и URL базы данных.
type Config struct {
	ServerURL       string `json:"server_address"`    // ServerURL представляет сетевой адрес (хост:порт), где будет размещен сервер.
	BaseReturnURL   string `json:"base_url"`          // BaseReturnURL представляет собой базовый адрес возврата (хост:порт), используемый для создания коротких URL.
	FileStoragePath string `json:"file_storage_path"` // FileStoragePath представляет собой путь к каталогу, используемому для хранения файлов.
	DatabaseURL     string `json:"database_dsn"`      // DatabaseURL представляет собой URL базы данных, используемой приложением.
	EnableHTTPS     bool   `json:"enable_https"`      // EnableHTTPS представляет собой флаг, указывающий на включение HTTPS сервера.
	ConfigPath      string // ConfigPath представляет собой путь к конфигурационному файлу.
}

// GetDefault возвращает объект Config с значениями по умолчанию.
// К значениям по умолчанию относятся localhost сервера и базовые URL возврата.
func GetDefault() Config {
	return Config{
		ServerURL:     "localhost:8080",
		BaseReturnURL: "http://localhost:8080",
	}
}

// GetDefaultWithTestDB возвращает объект Config с значениями по умолчанию, включая URL тестовой базы данных.
func GetDefaultWithTestDB() Config {
	return Config{
		ServerURL:     "localhost:8080",
		BaseReturnURL: "http://localhost:8080",
		DatabaseURL:   "postgres://test:test@localhost:5432/test?sslmode=disable",
	}
}

// Init инициализирует конфигурацию приложения, объединяя настройки из переменных среды и флагов командной строки.
// Сначала он инициализирует пустой объект Config, а затем заполняет его значениями из переменных среды с помощью функции configFromEnv.
// Затем он переопределяет любые настройки с помощью флагов командной строки, разобранных пакетом flag.
func New() (Config, error) {
	config := Config{}
	var err error
	config, err = configFromFile(config)
	if err != nil {
		return config, err
	}
	config = configFromFlags(config)
	config = configFromEnv(config)
	return config, nil
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
	if os.Getenv("ENABLE_HTTPS") == "true" {
		config.EnableHTTPS = true
	}
	return config
}

func configFromFlags(config Config) Config {
	flag.StringVar(&config.ServerURL, "a", "localhost:8080", "Net address host:port")
	flag.StringVar(&config.BaseReturnURL, "b", "http://localhost:8080", "Return base address host:port")
	flag.StringVar(&config.FileStoragePath, "f", "", "File storage path")
	flag.StringVar(&config.DatabaseURL, "d", "", "Database URL")
	flag.BoolVar(&config.EnableHTTPS, "s", false, "Enable HTTPS")
	flag.Parse()
	return config
}

func configFromFile(config Config) (Config, error) {
	flag.StringVar(&config.ConfigPath, "c", "", "Config file path")
	flag.Parse()
	if config.ConfigPath == "" {
		flag.StringVar(&config.ConfigPath, "config", "", "Config file path")
		flag.Parse()
	}
	if configPath, ok := os.LookupEnv("CONFIG"); ok {
		config.ConfigPath = configPath
	}
	if config.ConfigPath == "" {
		return config, nil
	}
	file, err := os.ReadFile(config.ConfigPath)
	if err != nil {
		return Config{}, err
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
