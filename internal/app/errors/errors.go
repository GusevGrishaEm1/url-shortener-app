package errors

import (
	"net/http"
)

type CustomError struct {
	Err         error
	Status      int
	Body        []byte
	ContentType string
	ShortURL    string
}

func (customErr CustomError) Error() string {
	return customErr.Err.Error()
}

func NewCustomError(err error) *CustomError {
	return &CustomError{
		Err: err,
	}
}

func NewCustomErrorInternal(err error) *CustomError {
	return &CustomError{
		Err:    err,
		Status: http.StatusInternalServerError,
	}
}

func NewCustomErrorBadRequest(err error) *CustomError {
	return &CustomError{
		Err:    err,
		Status: http.StatusBadRequest,
	}
}
