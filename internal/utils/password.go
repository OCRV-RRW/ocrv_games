package utils

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, fmt.Errorf("could not hash password %w", err)
	}
	return hashedPassword, nil
}

func VerifyPassword(hashedPassword []byte, candidatePassword string) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(candidatePassword))
}
