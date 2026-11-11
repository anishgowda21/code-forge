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

func GetUserRepos(ctx context.Context, db *sql.DB, owner User, limit, offset int) ([]Repo, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, name,created_at 
	 	 FROM repos where owner_id = ? 
	  	 ORDER BY created_at DESC 
	 	 LIMIT ? OFFSET ?`,
		owner.ID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	repos := make([]Repo, 0)
	for rows.Next() {
		var r Repo
		if err := rows.Scan(&r.ID, &r.Name, &r.CreatedAt); err != nil {
			return nil, err
		}
		r.Owner = owner.Username
		repos = append(repos, r)
	}
	return repos, rows.Err()
}

func GetUserRepo(ctx context.Context, db *sql.DB, repoName, owner string) (Repo, error) {
	var repo Repo
	err := db.QueryRowContext(ctx,
		`SELECT r.id,r.name,u.username,r.created_at
		 FROM repos r JOIN users u ON r.owner_id = u.id
		 WHERE r.name = ? AND u.username = ?`, repoName, owner).Scan(&repo.ID, &repo.Name, &repo.Owner, &repo.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return Repo{}, ErrNotFound
	}

	if err != nil {
		return Repo{}, err
	}

	return repo, nil
}
