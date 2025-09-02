package services

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	secret []byte
}

func NewAuthService(secret string) *AuthService { return &AuthService{secret: []byte(secret)} }

func (a *AuthService) IssueJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(a.secret)
}

func (a *AuthService) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authH := r.Header.Get("Authorization")
		if !strings.HasPrefix(strings.ToLower(authH), "bearer ") {
			http.Error(w, "missing bearer token", 401); return
		}
		tok := strings.TrimSpace(authH[len("Bearer "):])
		_, err := jwt.Parse(tok, func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 { return nil, jwt.ErrTokenUnverifiable }
			return a.secret, nil
		})
		if err != nil { http.Error(w, "invalid token", 401); return }
		next.ServeHTTP(w, r)
	})
}
