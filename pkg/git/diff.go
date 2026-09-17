package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GetChangedFiles returns a list of staged files if HEAD exists.
// If HEAD does not exist, it compares against the empty tree hash.
// If baseRef is provided, it diffs against that ref instead of index.
func GetChangedFiles(baseRef string) ([]string, error) {
	// First check if HEAD exists
	hasHead := true
	headCheck := exec.Command("git", "rev-parse", "--verify", "HEAD")
	if err := headCheck.Run(); err != nil {
		hasHead = false
	}

	var cmd *exec.Cmd
	if baseRef != "" {
		cmd = exec.Command("git", "diff", "--name-only", baseRef)
	} else if hasHead {
		cmd = exec.Command("git", "diff", "--cached", "--name-only")
	} else {
		// Empty tree hash for new repos
		cmd = exec.Command("git", "diff", "--cached", "--name-only", "4b825dc642cb6eb9a060e54bf8d69288fbee4904")
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %v", err)
	}

	var files []string
	for _, line := range strings.Split(out.String(), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}
