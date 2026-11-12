// helper.go
package git

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// repoDirExists reports whether repoPath exists and is a directory.
func repoDirExists(repoPath string) (bool, error) {
	info, err := os.Stat(repoPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("stat repo: %w", err)
	}

	if !info.IsDir() {
		return false, fmt.Errorf("stat repo: %q is not a directory", repoPath)
	}

	return true, nil
}

// ensureRepoExists is for the read paths: a missing repo is ErrRepoNotExists.
func ensureRepoExists(repoPath string) error {
	exists, err := repoDirExists(repoPath)
	if err != nil {
		return err
	}
	if !exists {
		return ErrRepoNotExists
	}
	return nil
}
