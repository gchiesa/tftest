package tfdir

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/git"
)

// GetRelativePathFromGitRepo returns the relative path to the git repo root.
// It is used to get the relative path to the git repo root for the Terraform directory.
// If the Terraform directory is an absolute path, it returns the relative path to the git repo root.
// If the Terraform directory is a relative path, it returns the relative path to the git repo root.
func GetRelativePathFromGitRepo(tfDir string, t *testing.T) (string, string, error) {
	if tfDir == "" {
		return "", "", fmt.Errorf("tfDir is required")
	}

	repoRoot := git.GetRepoRoot(t)

	t.Logf("The git repo root is %s", repoRoot)

	if !filepath.IsAbs(tfDir) {
		currentDir, err := os.Getwd()
		if err != nil {
			return "", "", fmt.Errorf("failed to get current directory: %v", err)
		}
		resolvedPath, err := filepath.Rel(repoRoot, filepath.Join(currentDir, tfDir))
		if err != nil {
			return "", "", fmt.Errorf("failed to get relative path to git repo: %v", err)
		}

		return resolvedPath, repoRoot, nil
	}

	if !strings.HasPrefix(tfDir, repoRoot) {
		return "", "", fmt.Errorf("the tfdir passed %s does not start with the repo root %s", tfDir, repoRoot)
	}

	return strings.TrimPrefix(tfDir, repoRoot), repoRoot, nil
}