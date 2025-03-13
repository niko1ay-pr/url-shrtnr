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

func TestUrlShrtnr_SaveRedirectDelete(t *testing.T) {

	cases := []struct {
		name     string
		url      string
		alias    string
		user     string
		password string
		error    string
	}{
		{
			name:     "Success workflow",
			url:      "https://google.com",
			alias:    gofakeit.Word() + gofakeit.Word(),
			user:     "admin",
			password: "adminpass",
		}, {
			name:     "Empty alias",
			url:      "https://google.com",
			alias:    "",
			user:     "admin",
			password: "adminpass",
		},

		// TODO: Faulty case
	}

	t.Parallel()
	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}
			customClient := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse

				},
			}
			e := httpexpect.WithConfig(httpexpect.Config{
				BaseURL:  u.String(),
				Client:   customClient,
				Reporter: httpexpect.NewAssertReporter(t),
			})

			resp := e.POST("/url").
				WithJSON(save.Request{URL: tc.url, Alias: tc.alias}).
				WithBasicAuth(tc.user, tc.password).
				Expect()
			alias := resp.Status(http.StatusOK).
				JSON().Object().Value("alias").String().NotEmpty().Raw()

			e.GET("/url/"+alias).
				WithBasicAuth(tc.user, tc.password).
				Expect().
				Status(http.StatusFound).Header("Location").NotEmpty()

			e.DELETE("/url/"+alias).WithBasicAuth(tc.user, tc.password).
				Expect().
				Status(http.StatusOK).
				JSON().Object().
				HasValue("status", "Ok")
		})

	}
}
