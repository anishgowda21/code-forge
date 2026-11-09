package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	"uuid"
)

var ErrRepoExists = errors.New("repository already exists")

type Repo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Owner     string `json:"owner"`
	CreatedAt string `json:"created_at"`
}

func CreateNewRepo(ctx context.Context, db *sql.DB, owner User, repoName string, createRepoOnDisk func() error) (Repo, error) {
	tx, err := db.BeginTx(ctx, nil)

	if err != nil {
		return Repo{}, err
	}

	defer tx.Rollback()

	id := uuid.NewV7().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err = tx.ExecContext(
		ctx, `INSERT INTO repos (id,owner_id,name,created_at) VALUES(?,?,?,?)`,
		id, owner.ID, repoName, now)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return Repo{}, ErrRepoExists
		}
		return Repo{}, err
	}

	if err = createRepoOnDisk(); err != nil {
		return Repo{}, err
	}

	return Repo{ID: id, Name: repoName, Owner: owner.Username, CreatedAt: now}, tx.Commit()
}
