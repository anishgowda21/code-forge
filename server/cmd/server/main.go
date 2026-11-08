package main

import (
	"log"
	"net/http"

	"github.com/anishgowda21/code-forge/server/internal/api"
)

func main() {
	mux := api.NewServer()

	addr := ":8090"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
