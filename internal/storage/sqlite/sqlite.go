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

	stmt, err := db.PrepareContext(ctx, `CREATE TABLE IF NOT EXISTS urls (id INTEGER PRIMARY KEY,
	alias TEXT NOT NULL UNIQUE,
	url TEXT NOT NULL,
	owner_id INTEGER REFERENCES users(id));
	CREATE INDEX IF NOT EXISTS alias_idx ON urls (alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err = db.PrepareContext(ctx, `CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL);
    CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
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

func (s *Storage) SaveURL(ctx context.Context, urlToSave string, alias string, ownerID int) error {
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

func (s *Storage) UpdateURL(ctx context.Context, alias string, newAlias string, url string) error {
	op := "storage.sqlite.UpdateURL"

	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM url WHERE alias = ?)", alias).Scan(&exists)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return storage.ErrUrlNotFound
	}

	stmt, err := s.db.PrepareContext(ctx, "UPDATE url SET alias = ?, url = ? WHERE alias = ?;")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, newAlias, url, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetAll(ctx context.Context) ([]storage.UrlEntity, error) {
	const op = "storage.sqlite.GetAll"
	query := "SELECT alias, url FROM url"

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var entities []storage.UrlEntity

	for rows.Next() {
		var alias, url string

		if err := rows.Scan(&alias, &url); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		entities = append(entities, storage.UrlEntity{Alias: alias, Url: url})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return entities, nil
}

func (s *Storage) UserSave(ctx context.Context, email string, password string) (int, error) {
	const op = "storage.sqlite.UserSave"

	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = ?);", email).Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	if exists {
		return 0, storage.ErrUserExists
	}

	stmt, err := s.db.PrepareContext(ctx, `INSERT INTO users (email, password) VALUES (?, ?);`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, email, password)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return int(id), nil
}

func (s *Storage) GetUserHash(ctx context.Context, email string, password string) (int, string, error) {
	const op = "storage.sqlite.GetUserHash"

	stmt, err := s.db.PrepareContext(ctx, "SELECT id, password FROM users WHERE email = ?;")
	if err != nil {
		return 0, "", fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var userId int
	var hashedPass string
	err = stmt.QueryRowContext(ctx, email).Scan(&userId, &hashedPass)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", storage.ErrUserNotFound
		}
		return 0, "", fmt.Errorf("%s: %w", op, err)
	}

	return userId, hashedPass, nil
}
