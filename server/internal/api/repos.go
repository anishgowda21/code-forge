package api

import (
	"encoding/json/v2"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/anishgowda21/code-forge/server/internal/database"
)

type RepoRequest struct {
	Name string `json:"name"`
}

func validateRepoName(repoName string) error {
	if len(repoName) == 0 || len(repoName) > 100 {
		return errors.New("invalid repo length")
	}

	if !repoNameRe.MatchString(repoName) || repoName == "." || repoName == ".." {
		return errors.New("invalid repo name")
	}

	if strings.HasSuffix(repoName, ".git") {
		return errors.New("repository name cannot end with .git")
	}

	return nil
}

func (s *server) createNewRepo(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	username := r.PathValue("username")
	var req RepoRequest

	if err := json.UnmarshalRead(r.Body, &req, json.RejectUnknownMembers(true)); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

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

	if err := validateRepoName(req.Name); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var repoCreated bool
	userRepoDir := filepath.Join(s.repoPath, user.Username)
	repo, err := database.CreateNewRepo(r.Context(),
		s.db, user, req.Name,
		func() error {
			if err := os.MkdirAll(userRepoDir, 0755); err != nil {
				return err
			}

			if err := initBareRepo(userRepoDir, req.Name); err != nil {
				return err
			}
			repoCreated = true
			return nil
		})

	switch {
	case errors.Is(err, database.ErrRepoExists):
		writeError(w, http.StatusConflict, err.Error())
		return
	case err != nil:
		log.Printf("repo: create error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		if repoCreated {
			cleanUpRepo(userRepoDir, req.Name)
		}
		return
	}

	writeJson(w, http.StatusCreated, repo)
}
