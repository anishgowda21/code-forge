package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	"uuid"
)

var (
	ErrExists   = errors.New("username already exists")
	ErrNotFound = errors.New("not found")
)

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
}

func CreateUser(ctx context.Context, db *sql.DB, username string) (User, error) {
	id := uuid.NewV7().String()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.ExecContext(ctx,
		`INSERT INTO users (id,username,created_at) VALUES(?,?,?)`,
		id, username, now)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return User{}, ErrExists
		}
		return User{}, err
	}

	return User{id, username, now}, nil
}

func FindByUsername(ctx context.Context, db *sql.DB, username string) (User, error) {
	var u User

	err := db.QueryRowContext(ctx, `
	SELECT id,username,created_at from users where username = ?`, username).Scan(
		&u.ID, &u.Username, &u.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}

	return u, err
}
