package all_test

import (
	"niko1ay-pr/url-shrtnr/internal/models"
	"testing"
)

var cases = []struct {
	name        string
	ownerID     int64
	entities    []models.UrlEntity
	expectedErr string
}{
	{
		name:    "Success several entities",
		ownerID: 1,
		entities: []models.UrlEntity{
			{
				Alias: "someAlias",
				Url:   "https://validurl.com",
			},
			{
				Alias: "anotherAlias",
				Url:   "https://anothervalidurl.net",
			},
		},
	},
	{
		name:     "Success zero entitites",
		ownerID:  1,
		entities: []models.UrlEntity{},
	},
	{
		name:        "No owner id specified",
		ownerID:     0,
		expectedErr: "no owner_id speciefied",
	},
}

func GetAllTest(t *testing.T) {
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(*testing.T) {
			t.Parallel()
		})
	}
}
