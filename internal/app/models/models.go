package models

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

type URLInfoBatchRequest struct {
	Array []OriginalURLInfoBatch
}

type URLInfoBatchResponse struct {
	Array []ShortURLInfoBatch
}

type ShortURLInfoBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"original_url"`
}

type OriginalURLInfoBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"short_url"`
}
