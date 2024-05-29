// Package security предоставляет middleware для обеспечения безопасности HTTP-запросов.
package security

import (
	"context"
	"errors"
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/golang-jwt/jwt/v4"
)

// Claims определяет структуру для хранения JWT.
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

// ShortenerService определяет методы, необходимые для работы с сервисом сокращения URL.
type ShortenerService interface {
	GetUserID(context.Context) int
}

// securityJWT определяет middleware для обеспечения безопасности с использованием JWT.
type securityJWT struct {
	ShortenerService
}

// NewSecurityMiddleware создает новый экземпляр middleware для обеспечения безопасности.
func NewSecurityMiddleware(service ShortenerService) *securityJWT {
	return &securityJWT{
		service,
	}
}

// RequiredUserID проверяет наличие идентификатора пользователя в запросе.
func (security *securityJWT) RequiredUserID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(string(models.UserID))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userID, err := getUserIDFromToken(cookie.Value)
		if err != nil || userID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), models.UserID, userID)))
	})
}

// Security обеспечивает безопасность обработки HTTP-запросов с использованием JWT.
func (security *securityJWT) Security(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(string(models.UserID))
		if errors.Is(err, http.ErrNoCookie) {
			newUserID := security.GetUserID(r.Context())
			token, _ := buildJWTString(newUserID)
			cookie = &http.Cookie{
				Name:  string(models.UserID),
				Value: token,
			}
			http.SetCookie(w, cookie)
		}
		userID, err := getUserIDFromToken(cookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), models.UserID, userID)))
	})
}

func buildJWTString(userID int) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, &Claims{
			RegisteredClaims: jwt.RegisteredClaims{},
			UserID:           userID,
		},
	)
	tokenString, err := token.SignedString([]byte("secretkey"))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func getUserIDFromToken(token string) (int, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte("secretkey"), nil
	})
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
