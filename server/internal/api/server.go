package api

import (
	"database/sql"
	"log"
	"net/http"
)

type server struct {
	repoPath string
	db       *sql.DB
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func catchAll(w http.ResponseWriter, r *http.Request) {
	log.Printf("unhandled: %s %s", r.Method, r.URL.Path)
	http.NotFound(w, r)
}

func NewServer(db *sql.DB, repoPath string) *http.ServeMux {
	s := &server{db: db, repoPath: repoPath}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthCheck)
	mux.HandleFunc("POST /users", s.signUpHandler)
	mux.HandleFunc("/users/{username}", s.getUserHandler)
	mux.HandleFunc("/users/{username}/repos", s.getUserReposHandler)
	mux.HandleFunc("POST /users/{username}/repos", s.createNewRepoHandler)
	mux.HandleFunc("/", catchAll)
	return mux
}
