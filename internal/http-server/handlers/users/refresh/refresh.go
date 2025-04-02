package refresh

import (
	"context"
	"log/slog"
	"net/http"
)

func New(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.refresh.New"

		
	}
}
