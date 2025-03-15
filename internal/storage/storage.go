package storage

import "errors"

const (
	AliasLength = 6
)

var (
	ErrUrlNotFound = errors.New("url not found")
	ErrUrlExists   = errors.New("url already exists")
)
