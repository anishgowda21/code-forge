package api

import (
	"log"
	"net/http"
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func catchAll(w http.ResponseWriter, r *http.Request) {
	log.Printf("unhandled: %s %s", r.Method, r.URL.Path)
	http.NotFound(w, r)
}

func NewServer() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", healthCheck)
	mux.HandleFunc("/", catchAll)
	return mux
}
