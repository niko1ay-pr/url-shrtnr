package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"niko1ay-pr/url-shrtnr/internal/storage"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(ctx context.Context, storagePath string) (*Storage, error) {
	op := "sqlite.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.PrepareContext(ctx, `CREATE TABLE IF NOT EXISTS url (id INTEGER PRIMARY KEY,
	alias TEXT NOT NULL UNIQUE,
	url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS alias_idx ON url (alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(ctx context.Context, urlToSave string, alias string) error {
	op := "storage.sqlite.SaveURL"

	stmt, err := s.db.PrepareContext(ctx, `INSERT INTO url (alias, url) VALUES (?, ?);`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, alias, urlToSave)
	if err != nil {
		if sqlErr, ok := err.(sqlite3.Error); ok && sqlErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return fmt.Errorf("%s: %w", op, storage.ErrUrlExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetURL(ctx context.Context, alias string) (string, error) {
	op := "storage.sqlite.GetURL"

	stmt, err := s.db.PrepareContext(ctx, "SELECT url FROM url WHERE alias = ?;")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var resUrl string
	err = stmt.QueryRowContext(ctx, alias).Scan(&resUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrUrlNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return resUrl, nil
}

func (s *Storage) DeleteURL(ctx context.Context, alias string) error {
	op := "storage.sqlite.DeleteURL"

	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM url WHERE alias = ?)", alias).Scan(&exists)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return storage.ErrUrlNotFound
	}

	stmt, err := s.db.PrepareContext(ctx, "DELETE FROM url WHERE alias = ?;")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
