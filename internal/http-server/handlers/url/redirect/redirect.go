package redirect

import (
	"log/slog"
	"net/http"
)

type Request struct {
}
type Response struct {
}

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.url.get.New"

	}
}
