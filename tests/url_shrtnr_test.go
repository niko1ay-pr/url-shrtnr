package tests

import (
	"net/http"
	"net/url"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/url/save"
	"niko1ay-pr/url-shrtnr/internal/lib/random"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gavv/httpexpect/v2"
)

const (
	host = "localhost:8080"
)

func TestUrlShrtnr_HappyPath(t *testing.T) {

	u := url.URL{
		Scheme: "http",
		Host:   host,
	}

	e := httpexpect.Default(t, u.String())
	randStr, err := random.NewRandomString(6)
	if err != nil {
		return
	}

	e.POST("/url").
		WithJSON(save.Request{URL: gofakeit.URL(),
			Alias: randStr}).
		WithBasicAuth("admin", "adminpass").
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		ContainsKey("alias")
}
