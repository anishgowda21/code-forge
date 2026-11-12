package git

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

type BlobResponse struct {
	Ref       string  `json:"ref"`
	Path      string  `json:"path"`
	Sha       string  `json:"sha"`
	Mode      string  `json:"mode"`
	Size      int64   `json:"size"`
	Binary    bool    `json:"binary"`
	Truncated bool    `json:"truncated"`
	Content   *string `json:"content,omitzero"`
}

const maxBlobSize = 1 << 20

func readBlob(ctx context.Context, sha, repoPath string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	cmdArgs := []string{
		"git",
		"--git-dir=" + repoPath,
		"cat-file",
		"blob",
		sha,
	}

	output, err := run(ctx, cmdArgs)
	if err != nil {
		return "", fmt.Errorf("read blob error: %w", err)
	}

	return output, nil
}

func getBlobObject(ctx context.Context, commit, filePath, repoPath string) (BlobResponse, error) {
	if err := ctx.Err(); err != nil {
		return BlobResponse{}, err
	}

	entry, err := getTreeEntry(ctx, commit, filePath, repoPath)

	if err != nil {
		return BlobResponse{}, err
	}

	if filePath != entry.Name {
		return BlobResponse{}, fmt.Errorf("get blob object error: unexpected output %q", entry.Name)
	}

	if entry.Type != TypeFile || (entry.Mode != "100644" && entry.Mode != "100755") {
		return BlobResponse{}, ErrNotAFile
	}

	resp := BlobResponse{
		Path:    entry.Name,
		Sha:     entry.Sha,
		Mode:    entry.Mode,
		Content: nil,
	}

	if entry.Size == nil {
		return BlobResponse{}, fmt.Errorf("get blob object error: file size not avilable %v", entry.Size)
	}
	resp.Size = *entry.Size

	if resp.Size > maxBlobSize {
		resp.Truncated = true
		return resp, nil
	}

	fileContent, err := readBlob(ctx, resp.Sha, repoPath)

	if err != nil {
		return BlobResponse{}, err
	}

	if strings.IndexByte(fileContent[:min(len(fileContent), 8000)], 0) >= 0 ||
		!utf8.ValidString(fileContent) {
		resp.Binary = true
		return resp, nil
	}

	resp.Content = &fileContent
	return resp, nil
}

func GetBlobInfo(ctx context.Context, repoPath, refName, filePath string) (BlobResponse, error) {
	if err := ctx.Err(); err != nil {
		return BlobResponse{}, err
	}

	if err := ensureRepoExists(repoPath); err != nil {
		return BlobResponse{}, err
	}

	refName, err := settleRefName(ctx, refName, repoPath)

	if err != nil {
		return BlobResponse{}, err
	}

	commit, err := getBranchCommit(ctx, refName, repoPath)

	if err != nil {
		return BlobResponse{}, err
	}

	resp, err := getBlobObject(ctx, commit, filePath, repoPath)

	if err != nil {
		return BlobResponse{}, err
	}

	resp.Ref = refName

	return resp, nil
}
