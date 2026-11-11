package api

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var usernameRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var repoNameRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

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

func validateUsername(username string) error {
	if len(username) == 0 || len(username) > 39 {
		return errors.New("invalid username length")
	}

	if !usernameRe.MatchString(username) {
		return errors.New("invalid username")
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
