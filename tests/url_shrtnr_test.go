package tests

import (
	"net/http"
	"net/url"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/urls/save"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/urls/update"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/users/register"
	"niko1ay-pr/url-shrtnr/internal/lib/random"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gavv/httpexpect/v2"
)

const (
	host = "localhost:8080"
)

func TestUrlShrtnr_HappyPath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

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

	randStr, err := random.NewRandomString(6)
	if err != nil {
		return
	}

	req := register.Request{
		Email:    "test_" + gofakeit.Email(),
		Password: gofakeit.Password(true, true, false, false, false, 8),
	}

	t.Log("Registering new user")
	regResp := e.POST("/users/register").WithJSON(req).Expect().
		Status(http.StatusCreated).JSON().Object()

	regResp.Value("status").String().IsEqual("Ok")
	regResp.Value("email").String().IsEqual(req.Email)

	t.Log("Logging user to get JWT-token")
	loginResp := e.POST("/users/login").WithJSON(req).Expect().
		Status(http.StatusOK).JSON().Object()

	token := loginResp.Value("token").String().NotEmpty().Raw()
	loginResp.Value("status").String().NotEmpty().IsEqual("Ok")
	loginResp.Value("email").String().NotEmpty().IsEqual(req.Email)

	t.Log("Saving URLs with JWT token")
	payload := save.Request{Entities: []save.Entity{
		save.Entity{
			URL:   gofakeit.URL(),
			Alias: randStr},
		save.Entity{
			URL:   gofakeit.URL(),
			Alias: "",
		},
	}}

	saveResp := e.POST("/url").
		WithJSON(payload).
		WithHeader("Authorization", "Bearer "+token).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	saveResp.Value("status").String().IsEqual("Ok")
	entities := saveResp.Value("entities").Array()
	entities.Length().IsEqual(2)

	// Store aliases for further deleting
	aliases := make(map[string]string)
	for i := range int(entities.Length().Raw()) {
		entity := entities.Value(i).Object()
		url := entity.Value("url").String().NotEmpty().Raw()
		alias := entity.Value("alias").String().NotEmpty().Raw()
		aliases[alias] = url
	}

	t.Log("Test redirecting URL")
	// check redirection without request to real external source
	var alias string
	var url string

	for a, u := range aliases {
		alias = a
		url = u
		break
	}
	e.GET("/url/" + alias).
		Expect().
		Status(http.StatusFound).
		Header("Location").
		IsEqual(url)

	t.Log("Updating URL")
	newURL := gofakeit.URL()
	respUpdate := e.PUT("/url/"+alias).WithJSON(update.Request{URL: newURL}).
		WithHeader("Authorization", "Bearer "+token).
		Expect().Status(http.StatusOK).JSON().Object()
	respUpdate.Value("url").String().IsEqual(newURL)
	respUpdate.Value("alias").String().IsEqual(alias)

	t.Log("Deleting URLs")
	for alias, _ := range aliases {
		deleteResp := e.DELETE("/url/"+alias).
			WithHeader("Authorization", "Bearer "+token).
			Expect().Status(http.StatusOK).JSON().Object()
		deleteResp.Value("status").String().IsEqual("Ok")
	}

	t.Log("Happy path test completed!")
}

func TestUrlShrtnr_SaveRedirectDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	return
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

func TestUrlShrtnr_UnauthorizedAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

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
