package tests

import (
	"net/http"
	"net/url"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/urls/save"
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

	payload := save.Request{Entities: []save.Entity{
		save.Entity{
			URL:   gofakeit.URL(),
			Alias: randStr},
		save.Entity{
			URL:   gofakeit.URL(),
			Alias: "",
		},
	}}

	resp := e.POST("/url").
		WithJSON(payload).
		WithBasicAuth("admin", "adminpass").
		Expect()

	resp.Status(http.StatusOK).
		JSON().Object().
		ContainsKey("entities")
}

func TestUrlShrtnr_SaveRedirectDelete(t *testing.T) {

	cases := []struct {
		name     string
		entities []save.Entity
		user     string
		password string
		error    string
	}{
		{
			name: "Success workflow",
			entities: []save.Entity{
				{
					URL:   "https://google.com",
					Alias: gofakeit.Word() + gofakeit.Word(),
				},
				{
					URL:   "https://yandex.ru",
					Alias: gofakeit.Word() + gofakeit.Word(),
				},
			},
			user:     "admin",
			password: "adminpass",
		}, {
			name: "Empty alias",
			entities: []save.Entity{
				{
					URL:   "https://google.com",
					Alias: "",
				},
			},
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

			payload := save.Request{Entities: []save.Entity{}}
			for _, entity := range tc.entities {
				payload.Entities = append(payload.Entities, entity)
			}

			resp := e.POST("/url").
				WithJSON(payload).
				WithBasicAuth(tc.user, tc.password).
				Expect()
			entity := resp.Status(http.StatusOK).
				JSON().Object().Value("entities").Array().First()
			// JSON().Object().Value("alias").String().NotEmpty().Raw()

			alias := entity.String().Raw()
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

// TODO:
func TestUrlShrtnrUnauthorizedAccess(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}

	e := httpexpect.Default(t, u.String())

	e.POST("/url").
		WithJSON(save.Request{Entities: []save.Entity{}}).
		Expect().
		Status(http.StatusUnauthorized)

	e.GET("/urls").
		Expect().
		Status(http.StatusUnauthorized)

	e.PUT("/url/testalias").
		WithJSON(map[string]string{"url": "https://example.com"}).
		Expect().
		Status(http.StatusUnauthorized)

	e.DELETE("/url/testalias").
		Expect().
		Status(http.StatusUnauthorized)

	// Test save with invalid token
	e.POST("/url").
		WithJSON(save.Request{Entities: []save.Entity{}}).
		WithHeader("Authorizatioin", "Bearer invalid-token").
		Expect().
		Status(http.StatusUnauthorized)
}
