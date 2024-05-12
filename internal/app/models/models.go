// Package models предоставляет структуры данных для работы с URL.
package models

import "time"

// Request представляет модель запроса на сокращение URL.
type Request struct {
	URL string `json:"url"` // URL URL для сокращения.
}

// Response представляет модель ответа с сокращенным URL.
type Response struct {
	Result string `json:"result"` // Result сокращенный URL.
}

// URLInfo представляет информацию о сокращенном URL.
type URLInfo struct {
	UUID        int    `json:"uuid"`         // UUID идентификатор URL.
	ShortURL    string `json:"short_url"`    // ShortURL сокращенный URL.
	OriginalURL string `json:"original_url"` // OriginalURL исходный URL.
}

// ShortURLInfoBatch представляет информацию о сокращенном URL для пакетной обработки.
type ShortURLInfoBatch struct {
	CorrelationID string `json:"correlation_id"` // CorrelationID идентификатор корреляции.
	ShortURL      string `json:"short_url"`      // ShortURL сокращенный URL.
}

// OriginalURLInfoBatch представляет информацию об исходном URL для пакетной отдачи.
type OriginalURLInfoBatch struct {
	CorrelationID string `json:"correlation_id"` // CorrelationID идентификатор корреляции.
	OriginalURL   string `json:"original_url"`   // OriginalURL исходный URL.
}

// URLByUser представляет информацию о URL, созданных пользователем.
type URLByUser struct {
	ShortURL    string `json:"short_url"`    // ShortURL сокращенный URL.
	OriginalURL string `json:"original_url"` // OriginalURL исходный URL.
}

// URLToDelete представляет информацию о URL, которые нужно удалить.
type URLToDelete struct {
	UserID   int    `json:"user_id"`   // UserID идентификатор пользователя.
	ShortURL string `json:"short_url"` // ShortURL сокращенный URL.
}

// Stats представляет статистику по сокращенным URL.
type Stats struct {
	URLS  int `json:"urls"`  // URLS количество сокращенных URL.
	Users int `json:"users"` // USERS количество пользователей.
}

// UserInfo представляет информацию о пользователе.
type UserInfo struct {
	UserID int // UserID идентификатор пользователя.
}

// URL представляет модель хранимого URL.
type URL struct {
	ID          int       // ID идентификатор URL в хранилище.
	ShortURL    string    // ShortURL сокращенный URL.
	OriginalURL string    // OriginalURL исходный URL.
	CreatedBy   int       // CreatedBy идентификатор пользователя, который создал URL.
	CreatedTS   time.Time // CreatedTS время создания URL.
	IsDeleted   bool      // IsDeleted флаг, указывающий, был ли URL удален.
}
