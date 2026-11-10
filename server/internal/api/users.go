package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/anishgowda21/code-forge/server/internal/database"
)

func (s *server) getUserHandler(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	user, err := database.FindByUsername(r.Context(), s.db, username)

	switch {
	case errors.Is(err, database.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
		return
	case err != nil:
		log.Printf("user get: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJson(w, http.StatusOK, user)
}
