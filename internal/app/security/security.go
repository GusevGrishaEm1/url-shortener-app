package security

import (
	"context"
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

type UserInfo string

const (
	UserID UserInfo = "UserID"
)

type ShortenerService interface {
	GetUserID(context.Context) int
}

type securityJWT struct {
	ShortenerService
}

func NewSecurityMiddleware(service ShortenerService) *securityJWT {
	return &securityJWT{
		service,
	}
}

func (security *securityJWT) RequiredUserID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(string(UserID))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userID, err := getUserIDFromToken(cookie.Value)
		if err != nil || userID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserID, userID)))
	})
}

func (security *securityJWT) Security(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(string(UserID))
		if errors.Is(err, http.ErrNoCookie) {
			newUserID := security.GetUserID(r.Context())
			token, _ := buildJWTString(newUserID)
			cookie = &http.Cookie{
				Name:  string(UserID),
				Value: token,
			}
			http.SetCookie(w, cookie)
		}
		userID, err := getUserIDFromToken(cookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserID, userID)))
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
