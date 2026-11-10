package git

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"time"
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

	cmd := exec.CommandContext(
		ctx,
		"git",
		"init",
		"--bare",
		"--initial-branch=main",
		repoPath,
	)

	cmd.WaitDelay = 5 * time.Second

	if output, err := cmd.CombinedOutput(); err != nil {
		CleanUpRepo(path, repoName)
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("git init aborted: %w", ctxErr)
		}
		return fmt.Errorf("git init failed: %w: %s", err, output)
	}

	return nil
}
