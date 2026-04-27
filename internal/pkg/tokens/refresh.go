package tokens

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateRefreshToken() (string, error) {
	const op = "pkg.jwt.GenerateRefreshToken"

	bytes := make([]byte, 32)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return hex.EncodeToString(bytes), nil
}
