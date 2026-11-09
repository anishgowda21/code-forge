package database

import (
	"database/sql"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS users(
	id		   	TEXT Primary Key,
	username   	TEXT NOT NULL UNIQUE,
	created_at 	TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS repos(
	id 		 	TEXT Primary Key,
	owner_id 	TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	name	 	TEXT NOT NULL,
	created_at	TEXT NOT NULL,
	UNIQUE(owner_id,name)
);
`

func OpenDB(path string) (*sql.DB, error) {
	path = filepath.Join(path, "data.db")
	dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}
	return db, nil
}
