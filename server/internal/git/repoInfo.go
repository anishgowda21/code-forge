package git

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

type RepoSummary struct {
	DefaultBranch string   `json:"default_branch"`
	Branches      []Branch `json:"branches"`
	Empty         bool     `json:"empty"`
}

type Branch struct {
	Name   string `json:"name"`
	Commit string `json:"commit"`
}

var ErrRepoNotExists = errors.New("repo does not exist")

func getDefaultRepoBranch(ctx context.Context, repoPath string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	var cmdArgs = []string{
		"git", "--git-dir=" + repoPath,
		"symbolic-ref", "--short", "HEAD",
	}
	output, err := run(ctx, cmdArgs)
	if err != nil {
		return "", fmt.Errorf(
			"get default branch error: %w",
			err,
		)
	}

	output = strings.TrimSpace(output)

	return output, nil
}

func getAllBranches(ctx context.Context, repoPath string) ([]Branch, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cmdArgs := []string{
		"git",
		"--git-dir=" + repoPath,
		"for-each-ref",
		"--format=%(refname:short) %(objectname)",
		"--sort=refname",
		"refs/heads",
	}
	output, err := run(ctx, cmdArgs)

	if err != nil {
		return nil, fmt.Errorf("get all branches error: %w", err)
	}

	branches := make([]Branch, 0)

	for line := range strings.SplitSeq(output, "\n") {
		name, commit, found := strings.Cut(line, " ")

		if found && name != "" && commit != "" {
			branches = append(branches, Branch{Name: name, Commit: commit})
		}
	}

	return branches, nil
}

func GetRepoSummary(ctx context.Context, repoPath string) (RepoSummary, error) {
	if err := ctx.Err(); err != nil {
		return RepoSummary{}, err
	}

	_, err := os.Stat(repoPath)

	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return RepoSummary{}, ErrRepoNotExists
		}
		return RepoSummary{}, fmt.Errorf("stat repo: %w", err)
	}

	defaultBranch, err := getDefaultRepoBranch(ctx, repoPath)

	if err != nil {
		return RepoSummary{}, fmt.Errorf("get repo summary: %w", err)
	}

	branches, err := getAllBranches(ctx, repoPath)

	if err != nil {
		return RepoSummary{}, fmt.Errorf("get repo summary: %w", err)
	}

	isEmpty := len(branches) == 0

	return RepoSummary{DefaultBranch: defaultBranch, Branches: branches, Empty: isEmpty}, nil
}
