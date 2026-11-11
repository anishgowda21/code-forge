package api

import (
	"errors"
	"regexp"
	"slices"
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

func validatePath(path string) error {
	if path == "" {
		return nil
	}

	if strings.HasPrefix(path, "/") || strings.ContainsRune(path, '\x00') {
		return errors.New("invalid path")
	}

	for seg := range strings.SplitSeq(path, "/") {
		if seg == "." || seg == ".." || seg == "" {
			return errors.New("invalid path")
		}
	}

	return nil
}

func validateRef(name string) error {
	if name == "" {
		return nil
	}

	//prefix
	if strings.HasPrefix(name, "-") {
		return errors.New("invalid ref")
	}

	//suffix
	if strings.HasSuffix(name, ".") {
		return errors.New("invalid ref")
	}

	//contains
	if strings.Contains(name, "..") || strings.Contains(name, "@{") {
		return errors.New("invalid ref")
	}
	var mustNotContain = []byte{'~', '^', ':', '?', '*', '[', '\\', ' '}
	for i := 0; i < len(name); i++ {
		char := name[i]

		if slices.Contains(mustNotContain, char) || char < 32 || char == 127 {
			return errors.New("invalid ref")
		}
	}

	//single @
	if name == "@" {
		return errors.New("invalid ref")
	}

	//path component
	for seg := range strings.SplitSeq(name, "/") {
		if strings.HasPrefix(seg, ".") ||
			strings.HasSuffix(seg, ".lock") || seg == "" {
			return errors.New("invalid ref")
		}
	}

	return nil
}
