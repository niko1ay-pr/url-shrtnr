package jwt

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sl"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	DefaultSigningMethod       = jwt.SigningMethodHS256
	ErrUnexpectedSigningMethod = jwt.ErrSignatureInvalid
)

type CustomClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateToken(log *slog.Logger, secret []byte, UserID int, UserEmail string) (string, error) {
	token := jwt.NewWithClaims(DefaultSigningMethod, &CustomClaims{
		int64(UserID),
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "url-shrtnr",
			Subject:   UserEmail,
		},
	})
	signedString, err := token.SignedString(secret)
	if err != nil {
		log.Error("failed to sign token", sl.Err(err))
		// render.JSON(w, r, response.Error("failed to login user"))
		return "", errors.New("failed to sign token")
	}
	return signedString, nil
}

func ExtractToken(r *http.Request) (string, error) {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("Authorization header missing")
	}

	authParts := strings.SplitN(authHeader, " ", 2)
	if len(authParts) != 2 || strings.ToLower(authParts[0]) != "bearer" {
		return "", errors.New("Invalid authorization header")
	}
	return authParts[1], nil
}

func ExtractClaims(tokenString string, secret []byte) (*CustomClaims, error) {
	const op = "internal.lib.jwt.ValidateToken"
	keyFn := func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if token.Method != DefaultSigningMethod {
			return nil, jwt.ErrSignatureInvalid
		}
		// TODO: check if token is expired and return proper error. Keep refreshing in mind
		//  errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet)
		return secret, nil
	}
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, keyFn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return claims, nil
}
