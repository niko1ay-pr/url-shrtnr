package hash

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const SaltSize = 16

func HashPassword(password string, salt string) (string, error) {
	const op = "lib.hash.HashPassword"
	bytes, err := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)

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

func RandomSalt(saltSize int) (string, error) {
	const op = "lib.hash.RandomSalt"
	var salt = make([]byte, saltSize)
	_, err := rand.Read(salt[:])

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return string(salt), nil
}
