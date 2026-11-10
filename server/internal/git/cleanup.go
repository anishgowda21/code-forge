package git

import (
	"log"
	"os"
	"path/filepath"
)

func CleanUpRepo(path, repoName string) {
	repoPath := filepath.Join(path, repoName+".git")
	if err := os.RemoveAll(repoPath); err != nil {
		log.Printf("failed to remove repo:%v", err)
	}
}
