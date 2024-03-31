package models

import "time"

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type URLInfo struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type ShortURLInfoBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type OriginalURLInfoBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type URLByUser struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type URLToDelete struct {
	UserID   int    `json:"user_id"`
	ShortURL string `json:"short_url"`
}

type UserInfo struct {
	UserID int
}

// URL представляет собой модель хранимого URL-а.
type URL struct {
	ID          int       // ID идентификатор URL-а в хранилище.
	ShortURL    string    // ShortURL сокращенный URL.
	OriginalURL string    // OriginalURL исходный URL.
	CreatedBy   int       // CreatedBy идентификатор пользователя, который создал URL.
	CreatedTS   time.Time // CreatedTS время создания URL.
	IsDeleted   bool      // IsDeleted флаг, указывающий, был ли URL удален.
}
