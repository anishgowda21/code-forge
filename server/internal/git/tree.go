package git

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

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
	Size *int64            `json:"size,omitzero"`
}

func checkDirPath(ctx context.Context, commit, dirPath, repoPath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	output, err := getTreeEntry(ctx, commit, dirPath, repoPath)

	if err != nil {
		return err
	}

	if dirPath != output.Name {
		return fmt.Errorf("check dir path error: unexpected output %q", output.Name)
	}

	if output.Type != TypeDir {
		return ErrNotADirectory
	}

	return nil
}

func getTreeEntry(ctx context.Context, commit, dirPath, repoPath string) (RefTreeEntry, error) {
	if err := ctx.Err(); err != nil {
		return RefTreeEntry{}, err
	}

	cmdArgs := []string{
		"git",
		"--git-dir=" + repoPath,
		"ls-tree",
		"-l",
		"-z",
		commit,
		"--",
		dirPath,
	}

	output, err := run(ctx, cmdArgs)
	if err != nil {
		return RefTreeEntry{}, fmt.Errorf("get tree entry: %w", err)
	}

	output = strings.TrimSuffix(output, "\x00")
	if output == "" {
		return RefTreeEntry{}, ErrPathNotExists
	}

	return getTreeEntryObject(output)
}

func getTreeEntryObject(output string) (RefTreeEntry, error) {
	rest, filename, exists := strings.Cut(output, "\t")

	if !exists || rest == "" || filename == "" {
		return RefTreeEntry{}, fmt.Errorf("get tree entry obj error: unexpected output %q", output)
	}

	restSlice := strings.Fields(rest)
	if len(restSlice) != 4 {
		return RefTreeEntry{}, fmt.Errorf("get tree entry obj error: unexpected output %q", rest)
	}

	mode, object, sha, sizeStr := restSlice[0], restSlice[1], restSlice[2], restSlice[3]
	treeObject, err := getTreeObjectType(mode, object)

	if err != nil {
		return RefTreeEntry{}, err
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
			return RefTreeEntry{}, fmt.Errorf("get tree entry obj error(parseInt): %w", err)
		}
		entry.Size = &size
	}

	return entry, nil
}

func getTreeObjectType(mode, object string) (RefTreeObjectType, error) {
	//Ref: https://git-scm.com/book/en/v2/Git-Internals-Git-Objects
	switch object {
	case "tree":
		return TypeDir, nil
	case "commit":
		return TypeSubmodule, nil
	case "blob":
		if mode == "120000" {
			return TypeSymlink, nil
		}
		return TypeFile, nil
	default:
		return "", fmt.Errorf("get tree object type: unexpected object:%q", object)
	}
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

		entry, err := getTreeEntryObject(item)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	slices.SortFunc(entries, func(a, b RefTreeEntry) int {
		aDir := a.Type == TypeDir
		bDir := b.Type == TypeDir

		if aDir && !bDir {
			return -1
		}

		if !aDir && bDir {
			return 1
		}

		return strings.Compare(a.Name, b.Name)
	})

	return entries, nil
}

func GetRefDirectory(ctx context.Context, repoPath, refName, dirPath string) (RefTreeResp, error) {
	if err := ctx.Err(); err != nil {
		return RefTreeResp{}, err
	}

	if err := ensureRepoExists(repoPath); err != nil {
		return RefTreeResp{}, err
	}

	refName, err := settleRefName(ctx, refName, repoPath)

	if err != nil {
		return RefTreeResp{}, err
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
