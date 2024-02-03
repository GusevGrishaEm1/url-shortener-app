package security

import (
	"context"
	"errors"
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

type UserInfo string

const (
	UserId UserInfo = "UserID"
)

type ShortenerService interface {
	GetUserID(context.Context) int
}

type SecurityHandlerImpl struct {
	ShortenerService
}

func New(config *config.Config, service ShortenerService) *SecurityHandlerImpl {
	return &SecurityHandlerImpl{
		service,
	}
}

func (securityHandler *SecurityHandlerImpl) RequestSecurityOnlyUserID(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("UserID")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userID, err := getUserIDFromToken(cookie.Value)
		if err != nil || userID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h(w, r.WithContext(context.WithValue(r.Context(), UserId, userID)))
	}
}

func (securityHandler *SecurityHandlerImpl) RequestSecurity(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("UserID")
		if errors.Is(err, http.ErrNoCookie) {
			newUserID := securityHandler.GetUserID(r.Context())
			token, _ := buildJWTString(newUserID)
			cookie = &http.Cookie{
				Name:  "UserID",
				Value: token,
			}
			http.SetCookie(w, cookie)
		}
		userID, err := getUserIDFromToken(cookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h(w, r.WithContext(context.WithValue(r.Context(), UserId, userID)))
	}
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
