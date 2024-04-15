// Пакет config предоставляет функциональность для инициализации и конфигурации приложения.
package config

import (
	"flag"
	"os"
)

// Структура Config представляет собой настройки конфигурации для приложения.
// Она включает в себя поля для URL сервера, базового URL возврата, пути к файловому хранилищу и URL базы данных.
type Config struct {
	ServerURL       string // ServerURL представляет сетевой адрес (хост:порт), где будет размещен сервер.
	BaseReturnURL   string // BaseReturnURL представляет собой базовый адрес возврата (хост:порт), используемый для создания коротких URL.
	FileStoragePath string // FileStoragePath представляет собой путь к каталогу, используемому для хранения файлов.
	DatabaseURL     string // DatabaseURL представляет собой URL базы данных, используемой приложением.
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
func New() Config {
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
