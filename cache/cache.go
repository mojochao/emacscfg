// Package cache provides git repository caching support.
package cache

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mojochao/emacsctl/util"
)

// IsCached checks if a repository is cached in the cache directory.
func IsCached(dir, name string) bool {
	repoDir := filepath.Join(dir, name)
	_, err := os.Stat(repoDir)
	return !os.IsNotExist(err)
}

// AddRepo adds a repository to the cache directory and returns its location in it.
func AddRepo(dir, name, url string) (string, error) {
	repoDir := filepath.Join(dir, name)
	if err := util.EnsureDir(dir); err != nil {
		return repoDir, err
	}
	if err := cloneRepo(url, repoDir); err != nil {
		return repoDir, err
	}
	return repoDir, nil
}

// RemoveRepo removes a repository from the cache directory.
func RemoveRepo(dir, name string) error {
	repoDir := filepath.Join(dir, name)
	return os.RemoveAll(repoDir)
}

// cloneRepo clones a git repository into the cache directory.
func cloneRepo(url string, dir string) error {
	cmd := "git"
	args := []string{"clone", url, dir}
	return exec.Command(cmd, args...).Run()
}
