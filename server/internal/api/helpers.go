package api

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

var usernameRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var repoNameRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func initBareRepo(path, repoName string) error {
	repoPath := filepath.Join(path, repoName+".git")

	if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
		return errors.New("repo already exists")
	}

	cmd := exec.Command(
		"git",
		"init",
		"--bare",
		"--initial-branch=main",
		repoPath,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		cleanUpRepo(path, repoName)
		return fmt.Errorf("git init failed: %w: %s", err, output)
	}

	return nil
}

func cleanUpRepo(path, repoName string) {
	repoPath := filepath.Join(path, repoName+".git")
	if err := os.RemoveAll(repoPath); err != nil {
		log.Printf("failed to remove repo:%v", err)
	}
}
