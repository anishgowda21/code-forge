package git

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
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
var ErrRefNotExists = errors.New("ref does not exist")

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

func GetRefDirectory(ctx context.Context, repoPath, refName, dirPath string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	_, err := os.Stat(repoPath)

	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", ErrRepoNotExists
		}
		return "", fmt.Errorf("stat repo: %w", err)
	}

	if refName == "" {
		refName, err = getDefaultRepoBranch(ctx, repoPath)
		if err != nil {
			return "", fmt.Errorf("get ref directory: %w", err)
		}
	}

	out, err := getBranchCommit(ctx, refName, repoPath)

	return out, err
}
