package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(raw_pass string) (string, error) {
	hashed_pass, err := bcrypt.GenerateFromPassword([]byte(raw_pass), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("Failed to hash password: %w", err)
	}

	return string(hashed_pass), nil

}

func CheckPassword(pass string, hashed_pass string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed_pass), []byte(pass))
}
