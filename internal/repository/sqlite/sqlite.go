package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"niko1ay-pr/url-shrtnr/internal/models"
	"niko1ay-pr/url-shrtnr/internal/repository"

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

	stmt, err := db.PrepareContext(ctx, `CREATE TABLE IF NOT EXISTS urls (
	id INTEGER PRIMARY KEY,
	alias TEXT NOT NULL UNIQUE,
	url TEXT NOT NULL,
	owner_id INTEGER REFERENCES users(id));
	CREATE INDEX IF NOT EXISTS owner_id_idx ON urls (owner_id);
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
    password TEXT NOT NULL,
    salt TEXT);
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
	op := "repository.sqlite.SaveURL"

	stmt, err := s.db.PrepareContext(ctx, `INSERT INTO urls (alias, url, owner_id) VALUES (?, ?, ?);`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, alias, urlToSave, ownerID)
	if err != nil {
		if sqlErr, ok := err.(sqlite3.Error); ok && sqlErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return fmt.Errorf("%s: %w", op, repository.ErrUrlExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetURL(ctx context.Context, alias string) (string, error) {
	op := "storage.sqlite.GetURL"

	stmt, err := s.db.PrepareContext(ctx, "SELECT url FROM urls WHERE alias = ?;")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var resUrl string
	err = stmt.QueryRowContext(ctx, alias).Scan(&resUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", repository.ErrUrlNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return resUrl, nil
}

func (s *Storage) DeleteURL(ctx context.Context, alias string) error {
	op := "repository.sqlite.DeleteURL"

	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM urls WHERE alias = ?)", alias).Scan(&exists)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return repository.ErrUrlNotFound
	}

	stmt, err := s.db.PrepareContext(ctx, "DELETE FROM urls WHERE alias = ?;")
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
	op := "repository.sqlite.UpdateURL"

	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM urls WHERE alias = ?)", alias).Scan(&exists)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !exists {
		return repository.ErrUrlNotFound
	}

	stmt, err := s.db.PrepareContext(ctx, "UPDATE urls SET alias = ?, url = ? WHERE alias = ?;")
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

func (s *Storage) GetAll(ctx context.Context) ([]models.UrlEntity, error) {
	const op = "repository.sqlite.GetAll"
	query := "SELECT alias, url FROM urls"

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var entities []models.UrlEntity

	for rows.Next() {
		var alias, url string

		if err := rows.Scan(&alias, &url); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		entities = append(entities, models.UrlEntity{Alias: alias, Url: url})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return entities, nil
}

func (s *Storage) UserSave(ctx context.Context, user models.User) error {
	const op = "repository.sqlite.UserSave"

	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = ?);", user.Email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if exists {
		return repository.ErrUserExists
	}

	stmt, err := s.db.PrepareContext(ctx, `INSERT INTO users (email, password, salt) VALUES (?, ?, ?);`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, user.Email, user.Password, user.Salt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetUserHash(ctx context.Context, email string) (*models.User, error) {
	const op = "repository.sqlite.GetUserHash"

	stmt, err := s.db.PrepareContext(ctx, "SELECT id, password, salt FROM users WHERE email = ?;")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	var userId int
	var hashedPass string
	var salt string
	err = stmt.QueryRowContext(ctx, email).Scan(&userId, &hashedPass, &salt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.User{Id: userId, Password: hashedPass, Salt: salt}, nil
}
