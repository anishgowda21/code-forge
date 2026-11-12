package git

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func getBranchCommit(ctx context.Context, branchName, repoPath string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	arg := "refs/heads/" + branchName + "^{commit}"

	cmdArgs := []string{
		"git",
		"--git-dir=" + repoPath,
		"rev-parse",
		"--verify",
		"--quiet",
		arg,
	}

	output, err := run(ctx, cmdArgs)

	if err != nil {
		/*
			So I did couple of variations on this
			first was returning the actual stderr and checking if it has "fatal: Needed a single revision",
			had second thoughts on it and asked claude and it mentioned
			As per git source rev-parse --verify --quiet calls exit(1) and been like this for long time
			so checking error code while passing --quite is more safer and sturdier then string litral check
		*/
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", ErrRefNotExists
		}
		return "", fmt.Errorf("get branch commit error: %w", err)
	}

	output = strings.TrimSpace(output)

	return output, nil
}

func settleRefName(ctx context.Context, refName, repoPath string) (string, error) {
	if refName != "" {
		return refName, nil
	}

	refName, err := getDefaultRepoBranch(ctx, repoPath)
	if err != nil {
		return "", err
	}

	return refName, nil
}
