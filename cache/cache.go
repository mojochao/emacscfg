// Package cache provides git repository caching support.
package cache

import (
	"os"
	"os/exec"
	"path/filepath"
)

// IsCached checks if a repository is cached in the cache directory.
func IsCached(cacheDir, repoName string) bool {
	repoDir := filepath.Join(cacheDir, repoName)
	_, err := os.Stat(repoDir)
	return !os.IsNotExist(err)
}

// AddRepo adds a repository to the cache directory and returns its location in it.
func AddRepo(cacheDir, repoName, repoUrl string) (string, error) {
	repoDir := filepath.Join(cacheDir, repoName)
	if err := cloneRepo(repoDir, repoUrl); err != nil {
		return repoDir, err
	}
	return repoDir, nil
}

// RemoveRepo removes a repository from the cache directory.
func RemoveRepo(cacheDir, repoName string) error {
	repoDir := filepath.Join(cacheDir, repoName)
	return os.RemoveAll(repoDir)
}

// cloneRepo clones a git repository into the cache directory.
func cloneRepo(repoDir, repoUrl string) error {
	cmd := "git"
	args := []string{"clone", repoUrl, repoDir}
	return exec.Command(cmd, args...).Run()
}
