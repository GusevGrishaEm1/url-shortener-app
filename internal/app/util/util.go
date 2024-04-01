// Package util предоставляет утилиты для работы с URL.
package util

import "math/rand"

// GenerateShortURL генерирует случайную короткую строку(размера 5) для использования в качестве сокращенного URL.
func GenerateShortURL() string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const shortURLLength = 5
	shortURL := make([]byte, shortURLLength)
	for i := range shortURL {
		shortURL[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(shortURL)
}
