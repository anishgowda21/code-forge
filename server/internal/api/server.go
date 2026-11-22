package api

import (
	"database/sql"
	"log"
	"net/http"
)

type server struct {
	gitPath  string
	repoPath string
	db       *sql.DB
	gitSem   chan struct{}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func catchAll(w http.ResponseWriter, r *http.Request) {
	log.Printf("unhandled: %s %s", r.Method, r.URL.Path)
	http.NotFound(w, r)
}

func NewServer(db *sql.DB, gitPath, repoPath string) *http.ServeMux {
	s := &server{db: db, gitPath: gitPath, repoPath: repoPath, gitSem: make(chan struct{}, 10)}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthCheck)

	mux.HandleFunc("POST /users/{username}/repos", s.createNewRepoHandler)
	mux.HandleFunc("POST /users", s.signUpHandler)

	mux.HandleFunc("/{username}/{repo}/{rest...}", s.gitHTTPHandler)

	mux.HandleFunc("GET /users/{username}", s.getUserHandler)
	mux.HandleFunc("GET /users/{username}/repos", s.getUserReposHandler)
	mux.HandleFunc("GET /users/{username}/repos/{repo}", s.getUserRepoHandler)
	mux.HandleFunc("GET /users/{username}/repos/{repo}/tree", s.getRepoDirectoryTree)
	mux.HandleFunc("GET /users/{username}/repos/{repo}/blob", s.getRepoBlob)
	mux.HandleFunc("GET /users/{username}/repos/{repo}/commits", s.getRepoCommitHistory)

	mux.HandleFunc("/", catchAll)
	return mux
}
