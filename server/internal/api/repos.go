package api

import (
	"context"
	"encoding/json/v2"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/anishgowda21/code-forge/server/internal/database"
	"github.com/anishgowda21/code-forge/server/internal/git"
)

var ErrInvalidQueryParam = errors.New("invalid query parameter")

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

func validatePagParama(limitStr, offsetStr string) (limit, offset int, err error) {
	limit, err = strconv.Atoi(limitStr)
	if err != nil {
		return 0, 0, ErrInvalidQueryParam
	}

	offset, err = strconv.Atoi(offsetStr)
	if err != nil {
		return 0, 0, ErrInvalidQueryParam
	}

	if limit < 1 || limit > 100 || offset < 0 {
		return 0, 0, errors.New("query parameter out of range")
	}

	return limit, offset, nil
}

func (s *server) createNewRepoHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	username := r.PathValue("username")
	var req RepoRequest

	if err := json.UnmarshalRead(r.Body, &req, json.RejectUnknownMembers(true)); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := validateRepoName(req.Name); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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

	var repoCreated bool
	userRepoDir := filepath.Join(s.repoPath, user.Username)
	repo, err := database.CreateNewRepo(r.Context(),
		s.db, user, req.Name,
		func() error {
			if err := os.MkdirAll(userRepoDir, 0755); err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()
			if err := git.InitBareRepo(ctx, userRepoDir, req.Name); err != nil {
				return err
			}
			repoCreated = true
			return nil
		})

	switch {
	case errors.Is(err, database.ErrRepoExists):
		writeError(w, http.StatusConflict, err.Error())
		return
	case errors.Is(err, git.ErrRepoDirExists):
		log.Printf("repo: orphan dir for %s/%s", user.Username, req.Name)
		writeError(w, http.StatusConflict, err.Error())
		return
	case err != nil:
		log.Printf("repo: create error: %v", err)
		if repoCreated {
			git.CleanUpRepo(userRepoDir, req.Name)
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJson(w, http.StatusCreated, repo)
}

func (s *server) getUserReposHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	username := r.PathValue("username")
	limitStr := queryParams.Get("limit")
	if limitStr == "" {
		limitStr = "20"
	}
	offsetStr := queryParams.Get("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}

	limit, offset, err := validatePagParama(limitStr, offsetStr)

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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

	repos, err := database.GetUserRepos(r.Context(), s.db, user, limit, offset)

	if err != nil {
		log.Printf("get repos: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJson(w, http.StatusOK, repos)
}
