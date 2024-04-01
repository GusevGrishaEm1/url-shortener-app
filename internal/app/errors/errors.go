// Пакет errors предоставляет пользовательские типы ошибок и функции для создания новых ошибок с различными статусами HTTP.
//
// Структура CustomError представляет пользовательскую ошибку. Она содержит оригинальную ошибку Err, статус HTTP, тело ответа, тип содержимого и короткий URL.
// Метод Error() позволяет структуре CustomError удовлетворять интерфейсу error.
// Функции NewCustomError, NewCustomErrorInternal и NewCustomErrorBadRequest создают новые экземпляры CustomError с различными статусами HTTP.
package errors

import (
	"net/http"
)

// CustomError представляет пользовательскую ошибку.
type CustomError struct {
	Err         error
	Status      int
	Body        []byte
	ContentType string
	ShortURL    string
}

// Error возвращает строку, представляющую пользовательскую ошибку.
func (customErr CustomError) Error() string {
	return customErr.Err.Error()
}

// NewCustomError создает новый экземпляр CustomError с заданной оригинальной ошибкой.
func NewCustomError(err error) *CustomError {
	return &CustomError{
		Err: err,
	}
}

// NewCustomErrorInternal создает новый экземпляр CustomError с оригинальной ошибкой и статусом HTTP 500 (внутренняя ошибка сервера).
func NewCustomErrorInternal(err error) *CustomError {
	return &CustomError{
		Err:    err,
		Status: http.StatusInternalServerError,
	}
}

// NewCustomErrorBadRequest создает новый экземпляр CustomError с оригинальной ошибкой и статусом HTTP 400 (неверный запрос).
func NewCustomErrorBadRequest(err error) *CustomError {
	return &CustomError{
		Err:    err,
		Status: http.StatusBadRequest,
	}
}
