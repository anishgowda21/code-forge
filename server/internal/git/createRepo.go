package git

import (
	"context"
	"fmt"
	"path/filepath"
)

func InitBareRepo(ctx context.Context, path, repoName string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	repoPath := filepath.Join(path, repoName+".git")
	exists, err := repoDirExists(repoPath)

	if err != nil {
		return err
	}
	if exists {
		return ErrRepoDirExists
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
