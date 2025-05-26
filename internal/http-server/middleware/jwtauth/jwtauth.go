package jwtauth

import (
	"context"
	"net/http"
	jwtlib "niko1ay-pr/url-shrtnr/internal/lib/jwt"
)

type contextKey struct{}

func ExtractClaims(secret []byte) func(http.Handler) http.Handler {
	const op = "internal.http-server.middleware.jwtauth.ExtractClaims"
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			tokenString, err := jwtlib.ExtractToken(r)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			claims, err := jwtlib.ExtractClaims(tokenString, secret)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), contextKey{}, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

func GetUserID(ctx context.Context) int {
	if ctx == nil {
		return 0
	}

	if userID, ok := ctx.Value(contextKey{}).(int); ok {
		return userID
	}
	return 0
}
