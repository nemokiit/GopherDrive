package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GenerateRefreshToken() (string, error) {
	const op = "lib.jwt.GenerateRefreshToken"

	bytes := make([]byte, 32)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return hex.EncodeToString(bytes), nil
}

func GenerateAccessToken(secretKey string, id uuid.UUID, ttl time.Duration) (string, error) {
	const op = "lib.jwt.GenerateAccessToken"

	claims := jwt.MapClaims{
		"uid": id.String(),
		"exp": time.Now().Add(ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	stringToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return stringToken, nil
}

func ParseAccessToken(accessToken string, secretToken string) (uuid.UUID, error) {
	const op = "lib.jwt.ParseAccessToken"

	token, err := jwt.Parse(accessToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%s: %w", op, errors.New("unexpected signing method"))
		}

		return []byte(secretToken), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, fmt.Errorf("%s: %w", op, errors.New("invalid token"))
	}
	uidStr, ok := claims["uid"].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("%s: %w", op, errors.New("uid not found in token"))
	}
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", op, err)
	}

	return uid, nil
}
