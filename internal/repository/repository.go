package repository

import "errors"

const (
	AliasLength = 6
)

var (
	ErrUrlNotFound   = errors.New("url not found")
	ErrUrlExists     = errors.New("url already exists")
	ErrAliasExists   = errors.New("alias already exists")
	ErrAliasNotFound = errors.New("alias not found")
	ErrUserExists    = errors.New("user already exists")
	ErrUserNotFound  = errors.New("user not found")
)
