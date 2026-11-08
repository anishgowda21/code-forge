package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	"uuid"
)

var ErrExists = errors.New("username already exists")

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateUser(ctx context.Context, db *sql.DB, username string) (User, error) {
	id := uuid.NewV7().String()
	now := time.Now().UTC()
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
