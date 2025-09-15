package models

type User struct {
	Id       int64
	Email    string
	Password string
	Salt     string
}

type UrlEntity struct {
	Alias string `json:"alias,required"`
	Url   string `json:"url,required"`
}

type RefreshToken struct {
	Id    int64
	Token string
	// TODO: add expired_at
}
