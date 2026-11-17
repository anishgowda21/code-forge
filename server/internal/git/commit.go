package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type CommitHistoryObj struct {
	Sha        string   `json:"sha"`
	Parents    []string `json:"parents"`
	Author     string   `json:"author"`
	AuthoredAt string   `json:"authored_at"`
	Message    string   `json:"message"`
}

func commitLog(ctx context.Context, sha, filePath, repoPath string, limit, offset int) ([]CommitHistoryObj, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	cmdArgs := []string{
		"git",
		"--git-dir=" + repoPath,
		"log",
		fmt.Sprintf("--max-count=%d", limit),
		fmt.Sprintf("--skip=%d", offset),
		"-z",
		"--format=%H%x00%P%x00%an%x00%at%x00%s",
		sha,
	}

	if filePath != "" {
		cmdArgs = append(cmdArgs, "--", filePath)
	}

	output, err := run(ctx, cmdArgs)
	if err != nil {
		return nil, fmt.Errorf("commit logs error: %w", err)
	}

	output = strings.TrimSuffix(output, "\x00")

	if output == "" {
		return []CommitHistoryObj{}, nil
	}

	pieces := strings.Split(output, "\x00")

	if len(pieces)%5 != 0 {
		return nil, fmt.Errorf("commit logs error: unexpected output: %d fields", len(pieces))

	}

	commitObjs := make([]CommitHistoryObj, 0)

	for i := 0; i < len(pieces); i = i + 5 {
		commitSha := pieces[i]

		parents := []string{}
		if pieces[i+1] != "" {
			parents = strings.Split(pieces[i+1], " ")
		}

		author := pieces[i+2]

		authoredAtUnix, err := strconv.ParseInt(pieces[i+3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("commit logs error: bad timestamp %q: %w", pieces[i+3], err)
		}
		authoredAt := time.Unix(authoredAtUnix, 0).UTC().Format(time.RFC3339)

		message := pieces[i+4]

		commitObj := CommitHistoryObj{
			Sha:        commitSha,
			Parents:    parents,
			Author:     author,
			AuthoredAt: authoredAt,
			Message:    message,
		}

		commitObjs = append(commitObjs, commitObj)
	}

	return commitObjs, nil
}

func GetCommitHistory(ctx context.Context, repoPath, refName, filePath string, limit, offset int) ([]CommitHistoryObj, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := ensureRepoExists(repoPath); err != nil {
		return nil, err
	}

	refName, err := settleRefName(ctx, refName, repoPath)
	if err != nil {
		return nil, err
	}

	commit, err := getBranchCommit(ctx, refName, repoPath)

	if err != nil {
		return nil, err
	}

	logs, err := commitLog(ctx, commit, filePath, repoPath, limit, offset)
	if err != nil {
		return nil, err
	}

	return logs, nil
}
