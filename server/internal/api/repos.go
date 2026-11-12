package api

import (
	"context"
	"encoding/json/v2"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anishgowda21/code-forge/server/internal/database"
	"github.com/anishgowda21/code-forge/server/internal/git"
)

var ErrInvalidQueryParam = errors.New("invalid query parameter")

type RepoRequest struct {
	Name string `json:"name"`
}

type RepoInfo struct {
	database.Repo
	git.RepoSummary
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
	if err := validateUsername(username); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
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
		log.Printf("get user repos: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJson(w, http.StatusOK, repos)
}

func (s *server) getUserRepoHandler(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	repoName := r.PathValue("repo")

	if err := validateUsername(username); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	if err := validateRepoName(repoName); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	repo, err := database.GetUserRepo(r.Context(), s.db, repoName, username)

	switch {
	case errors.Is(err, database.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
		return
	case err != nil:
		log.Printf("user repo get: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	repoPath := filepath.Join(s.repoPath, repo.Owner, repo.Name+".git")
	repoSummary, err := git.GetRepoSummary(r.Context(), repoPath)

	switch {
	case errors.Is(err, git.ErrRepoNotExists):
		log.Printf("repo %s/%s missing on disk: %v", repo.Owner, repo.Name, err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	case err != nil:
		log.Printf("repo summary get: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJson(w, http.StatusOK, RepoInfo{repo, repoSummary})
}

func (s *server) getRepoDirectoryTree(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	username := r.PathValue("username")
	repoName := r.PathValue("repo")
	refName := queryParams.Get("ref")
	dirPath := queryParams.Get("path")

	if dirPath != "/" {
		dirPath = strings.TrimSuffix(dirPath, "/")
	}

	if err := validateUsername(username); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	if err := validateRepoName(repoName); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	if err := validatePath(dirPath); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateRef(refName); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	repo, err := database.GetUserRepo(r.Context(), s.db, repoName, username)

	switch {
	case errors.Is(err, database.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
		return
	case err != nil:
		log.Printf("user repo get: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	repoPath := filepath.Join(s.repoPath, repo.Owner, repo.Name+".git")
	resp, err := git.GetRefDirectory(r.Context(), repoPath, refName, dirPath)

	if err != nil {
		switch {
		case errors.Is(err, git.ErrPathNotExists), errors.Is(err, git.ErrRefNotExists):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, git.ErrNotADirectory):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, git.ErrRepoNotExists):
			log.Printf("repo %s/%s missing on disk: %v", repo.Owner, repo.Name, err)
			writeError(w, http.StatusInternalServerError, "internal error")
		default:
			log.Printf("repo dir tree get: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJson(w, http.StatusOK, resp)
}

func (s *server) getRepoBlob(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	username := r.PathValue("username")
	repoName := r.PathValue("repo")
	refName := queryParams.Get("ref")
	filePath := queryParams.Get("path")

	if filePath == "" {
		writeError(w, http.StatusBadRequest, "path requierd")
		return
	}

	if err := validateUsername(username); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	if err := validateRepoName(repoName); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	if err := validatePath(filePath); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateRef(refName); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	repo, err := database.GetUserRepo(r.Context(), s.db, repoName, username)

	switch {
	case errors.Is(err, database.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
		return
	case err != nil:
		log.Printf("user repo get: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	repoPath := filepath.Join(s.repoPath, repo.Owner, repo.Name+".git")
	resp, err := git.GetBlobInfo(r.Context(), repoPath, refName, filePath)

	if err != nil {
		switch {
		case errors.Is(err, git.ErrPathNotExists), errors.Is(err, git.ErrRefNotExists):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, git.ErrNotAFile):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, git.ErrRepoNotExists):
			log.Printf("repo %s/%s missing on disk: %v", repo.Owner, repo.Name, err)
			writeError(w, http.StatusInternalServerError, "internal error")
		default:
			log.Printf("repo dir tree get: %v", err)
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJson(w, http.StatusOK, resp)
}
