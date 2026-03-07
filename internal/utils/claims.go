package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	// Поле для confirm token при регистрации — пусто в обычных JWT
	ConfirmLogin string `json:"confirm_login,omitempty"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID uuid.UUID, secret string, duration time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateConfirmToken — временный токен на 15 минут для шага заполнения профиля.
// Содержит login (email или телефон) вместо userID.
func GenerateConfirmToken(login, secret string) (string, error) {
	claims := Claims{
		UserID:       uuid.Nil, // не используется
		ConfirmLogin: login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseConfirmToken — проверяет confirm token и возвращает login
func ParseConfirmToken(tokenString, secret string) (string, error) {
	claims, err := VerifyJWT(tokenString, secret)
	if err != nil {
		return "", err
	}
	if claims.ConfirmLogin == "" {
		return "", errors.New("недействительный токен подтверждения")
	}
	return claims.ConfirmLogin, nil
}

func VerifyJWT(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}
