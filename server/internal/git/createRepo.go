package git

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var ErrRepoDirExists = errors.New("repo already exists")

func InitBareRepo(ctx context.Context, path, repoName string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	repoPath := filepath.Join(path, repoName+".git")

	if _, err := os.Stat(repoPath); err == nil {
		return ErrRepoDirExists
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("stat repo path: %w", err)
	}

	var cmdArgs = []string{"git", "init", "--bare", "--initial-branch=main", repoPath}

	if output, err := run(ctx, cmdArgs); err != nil {
		CleanUpRepo(path, repoName)
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("git init aborted: %w", ctxErr)
		}
		return fmt.Errorf("git init failed: %w: %s", err, output)
	}

	return nil
}
