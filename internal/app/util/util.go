package util

import "math/rand"

func GenerateShortURL() string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const shortURLLength = 5
	shortURL := make([]byte, shortURLLength)
	for i := range shortURL {
		shortURL[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(shortURL)
}
