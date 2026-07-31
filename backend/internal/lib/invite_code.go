package lib

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
)

func GenerateInviteCode() (string, error) {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate invite code : %w", err)
	}

	code := base32.StdEncoding.EncodeToString(b)
	return strings.TrimRight(code, "="), nil
}
