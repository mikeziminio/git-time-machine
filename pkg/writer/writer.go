package writer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Writer writes rewritten Git history to a new repository
type Writer struct {
	outputDir string
	quiet     bool
}

// NewWriter creates a new Git writer for the given output directory
func NewWriter(outputDir string, quiet bool) *Writer {
	return &Writer{
		outputDir: outputDir,
		quiet:     quiet,
	}
}

// InitRepository initializes a new Git repository
func (w *Writer) InitRepository() error {
	// Check if directory exists
	if _, err := os.Stat(w.outputDir); err == nil {
		// Directory exists, check if it's empty or non-empty git repo
		entries, err := os.ReadDir(w.outputDir)
		if err != nil {
			return fmt.Errorf("failed to read output directory: %w", err)
		}
		if len(entries) > 0 {
			return fmt.Errorf("output directory is not empty: %s", w.outputDir)
		}
	} else if os.IsNotExist(err) {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(w.outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	} else {
		return fmt.Errorf("failed to check output directory: %w", err)
	}

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = w.outputDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to init git repository: %w, output: %s", err, string(output))
	}

	return nil
}

// CreateCommit creates a new commit with the given parameters
func (w *Writer) CreateCommit(author, email string, date time.Time, message string) (string, error) {
	// Set environment variables for commit date and author
	env := os.Environ()
	env = append(env, fmt.Sprintf("GIT_AUTHOR_DATE=%s", date.Format(time.RFC3339)))
	env = append(env, fmt.Sprintf("GIT_COMMITTER_DATE=%s", date.Format(time.RFC3339)))
	env = append(env, fmt.Sprintf("GIT_AUTHOR_NAME=%s", author))
	env = append(env, fmt.Sprintf("GIT_AUTHOR_EMAIL=%s", email))
	env = append(env, fmt.Sprintf("GIT_COMMITTER_NAME=%s", author))
	env = append(env, fmt.Sprintf("GIT_COMMITTER_EMAIL=%s", email))

	// Create a dummy file for the commit
	tmpFile := fmt.Sprintf("%s/.tmp-commit-%d", w.outputDir, time.Now().UnixNano())
	if err := os.WriteFile(tmpFile, []byte(fmt.Sprintf("Commit at %s\n", date)), 0644); err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	// Add file to index
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = w.outputDir
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(tmpFile)
		return "", fmt.Errorf("failed to add file: %w, output: %s", err, string(output))
	}

	// Create commit
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = w.outputDir
	cmd.Env = env
	output, err = cmd.CombinedOutput()
	if err != nil {
		os.Remove(tmpFile)
		return "", fmt.Errorf("failed to create commit: %w, output: %s", err, string(output))
	}

	// Clean up temp file
	os.Remove(tmpFile)

	// Get the SHA of the newly created commit
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = w.outputDir
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit SHA: %w", err)
	}

	sha := strings.TrimSpace(string(output))
	
	if !w.quiet {
		fmt.Printf("Created commit: %s (author: %s <%s>, date: %s)\n", sha[:7], author, email, date.Format("2006-01-02 15:04:05"))
	}

	return sha, nil
}

// CreateBranch creates a new branch at the given commit
func (w *Writer) CreateBranch(branchName, commitSHA string) error {
	cmd := exec.Command("git", "branch", branchName, commitSHA)
	cmd.Dir = w.outputDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create branch: %w, output: %s", err, string(output))
	}

	return nil
}

// CreateTag creates a new tag at the given commit
func (w *Writer) CreateTag(tagName, commitSHA string) error {
	cmd := exec.Command("git", "tag", tagName, commitSHA)
	cmd.Dir = w.outputDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create tag: %w, output: %s", err, string(output))
	}

	return nil
}
