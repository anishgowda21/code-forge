package api

import (
	"encoding/json/v2"
	"errors"
	"log"
	"net/http"

	"github.com/anishgowda21/code-forge/server/internal/database"
)

type signUpRequest struct {
	Username string `json:"username"`
}

func (s *server) signUpHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)

	var req signUpRequest
	if err := json.UnmarshalRead(r.Body, &req, json.RejectUnknownMembers(true)); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if err := validateUsername(req.Username); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	newUser, err := database.CreateUser(r.Context(), s.db, req.Username)

	switch {
	case errors.Is(err, database.ErrExists):
		writeError(w, http.StatusConflict, "username already taken")
		return
	case err != nil:
		log.Printf("signup: error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJson(w, http.StatusCreated, newUser)
}
