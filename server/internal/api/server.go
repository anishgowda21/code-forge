package api

import (
	"database/sql"
	"log"
	"net/http"
)

type server struct {
	db *sql.DB
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func catchAll(w http.ResponseWriter, r *http.Request) {
	log.Printf("unhandled: %s %s", r.Method, r.URL.Path)
	http.NotFound(w, r)
}

func NewServer(db *sql.DB) *http.ServeMux {
	s := &server{db: db}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthCheck)
	mux.HandleFunc("POST /users", s.signUpHandler)
	mux.HandleFunc("/", catchAll)
	return mux
}
