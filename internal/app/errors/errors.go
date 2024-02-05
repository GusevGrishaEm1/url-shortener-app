package errors

import (
	"errors"
)

var (
	ErrOriginalIsEmpty      = errors.New("original url is empty")
	ErrOriginalURLNotFound  = errors.New("original url isn't found")
	ErrOriginalURLIsDeleted = errors.New("original url is deleted")
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

type CustomError struct {
	Err         error
	Status      int
	Body        []byte
	ContentType string
}

func (customErr CustomError) Error() string {
	return customErr.Err.Error()
}

func NewCustomErrorWithMessage(message string) *CustomError {
	return &CustomError{
		Err: errors.New(message),
	}
}

func NewCustomError(err error) *CustomError {
	return &CustomError{
		Err: err,
	}
}
