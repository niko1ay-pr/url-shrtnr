package hash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	const op = "lib.hash.HashPassword"
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("%s: %v", op, err)
	}

	return string(bytes), nil
}

func CheckPasswordHash(hashedPass string, password string) (bool, error) {
	const op = "lib.hash.CheckPasswordHash"

	err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(password))

	if err != nil {
		return false, fmt.Errorf("%s, %v", op, err)
	}

	return true, nil
}
