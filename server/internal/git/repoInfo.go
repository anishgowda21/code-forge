package git

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strconv"
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

type RefTreeObjectType string

const (
	TypeDir       RefTreeObjectType = "dir"
	TypeFile      RefTreeObjectType = "file"
	TypeSymlink   RefTreeObjectType = "symlink"
	TypeSubmodule RefTreeObjectType = "submodule"
)

type RefTreeResp struct {
	Ref     string         `json:"ref"`
	Path    string         `json:"path"`
	Entries []RefTreeEntry `json:"entries"`
}

type RefTreeEntry struct {
	Name string            `json:"name"`
	Type RefTreeObjectType `json:"type"`
	Mode string            `json:"mode"`
	Sha  string            `json:"sha"`
	Size *int64            `json:"size,omitempty"`
}

var ErrRepoNotExists = errors.New("repo does not exist")
var ErrRefNotExists = errors.New("ref does not exist")
var ErrPathNotExists = errors.New("path does not exist")
var ErrNotADirectory = errors.New("not a directory")

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

func checkDirPath(ctx context.Context, commit, dirPath, repoPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	cmdArgs := []string{
		"git",
		"--git-dir=" + repoPath,
		"ls-tree",
		"-z",
		"--format=%(objecttype) %(path)",
		commit,
		"--",
		dirPath,
	}

	output, err := run(ctx, cmdArgs)
	if err != nil {
		return fmt.Errorf("check dir path error: %w", err)
	}

	output = strings.TrimSuffix(output, "\x00")
	if output == "" {
		return ErrPathNotExists
	}

	object, path, exists := strings.Cut(output, " ")
	if !exists || object == "" || dirPath != path {
		return fmt.Errorf("check dir path error: unexpected output %q", output)
	}

	if object != "tree" {
		return ErrNotADirectory
	}

	return nil
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

func getCommitTree(ctx context.Context, commit, dirPath, repoPath string) ([]RefTreeEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	treeish := commit

	if dirPath != "" {
		err := checkDirPath(ctx, commit, dirPath, repoPath)
		if err != nil {
			return nil, err
		}
		treeish = treeish + ":" + dirPath
	}

	cmdArgs := []string{
		"git",
		"--git-dir=" + repoPath,
		"ls-tree",
		"-l",
		"-z",
		treeish,
	}

	output, err := run(ctx, cmdArgs)

	if err != nil {
		return nil, fmt.Errorf("get commit tree: %w", err)
	}

	entries := make([]RefTreeEntry, 0)

	for item := range strings.SplitSeq(output, "\x00") {
		if item == "" {
			continue
		}

		rest, filename, exists := strings.Cut(item, "\t")
		if !exists || rest == "" || filename == "" {
			return nil, fmt.Errorf("get commit tree error: unexpected output %q", item)
		}

		restSlice := strings.Fields(rest)
		if len(restSlice) != 4 {
			return nil, fmt.Errorf("get commit tree error: unexpected output %q", rest)
		}

		mode, object, sha, sizeStr := restSlice[0], restSlice[1], restSlice[2], restSlice[3]
		treeObject, err := getTreeObjectType(mode, object)

		if err != nil {
			return nil, err
		}

		entry := RefTreeEntry{
			Name: filename,
			Type: treeObject,
			Mode: mode,
			Sha:  sha,
			Size: nil,
		}

		if treeObject == TypeFile {
			size, err := strconv.ParseInt(sizeStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("get commit tree error(parseInt): %w", err)
			}
			entry.Size = &size
		}

		entries = append(entries, entry)
	}

	return entries, nil
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

func GetRefDirectory(ctx context.Context, repoPath, refName, dirPath string) (RefTreeResp, error) {
	if err := ctx.Err(); err != nil {
		return RefTreeResp{}, err
	}

	_, err := os.Stat(repoPath)

	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return RefTreeResp{}, ErrRepoNotExists
		}
		return RefTreeResp{}, fmt.Errorf("stat repo: %w", err)
	}

	if refName == "" {
		refName, err = getDefaultRepoBranch(ctx, repoPath)
		if err != nil {
			return RefTreeResp{}, err
		}
	}

	commit, err := getBranchCommit(ctx, refName, repoPath)

	if err != nil {
		return RefTreeResp{}, err
	}

	treeEntries, err := getCommitTree(ctx, commit, dirPath, repoPath)

	if err != nil {
		return RefTreeResp{}, err
	}

	resp := RefTreeResp{
		Ref:     refName,
		Path:    dirPath,
		Entries: treeEntries,
	}

	return resp, err
}
