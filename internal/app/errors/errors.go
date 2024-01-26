package errors

import "errors"

var (
	ErrOriginalIsEmpty     = errors.New("original url is empty")
	ErrOriginalURLNotFound = errors.New("original url isn't found")
)

type OriginalURLAlreadyExists struct {
	ShortURL string
}

func (err *OriginalURLAlreadyExists) Error() string {
	return "original url already exists"
}

func NewErrOriginalURLAlreadyExists(shortURL string) *OriginalURLAlreadyExists {
	return &OriginalURLAlreadyExists{
		ShortURL: shortURL,
	}
}
