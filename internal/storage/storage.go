package storage

import "errors"

const (
	AliasLength = 6
)

type UrlEntity struct {
	Alias string `json:"alias,required"`
	Url   string `json:"url,required"`
}

var (
	ErrUrlNotFound  = errors.New("url not found")
	ErrUrlExists    = errors.New("url already exists")
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)
