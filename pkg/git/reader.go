package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Commit represents a Git commit with its metadata
type Commit struct {
	SHA        string
	Author     string
	Email      string
	Date       string
	DateParsed time.Time
	Message    string
	Parents    []string
	// NewDate is the rewritten date (optional)
	NewDate time.Time
}

// Branch represents a Git branch with its commits
type Branch struct {
	Name    string
	Commits []Commit
}

// Reader reads Git repository information
type Reader struct {
	repoPath string
}

// NewReader creates a new Git reader for the given repository path
func NewReader(repoPath string) *Reader {
	return &Reader{repoPath: repoPath}
}

// Validate checks if the path is a valid Git repository
func (r *Reader) Validate() error {
	gitDir := fmt.Sprintf("%s/.git", r.repoPath)
	_, err := os.Stat(gitDir)
	if err != nil {
		return fmt.Errorf("not a Git repository: %w", err)
	}
	return nil
}

// GetBranches reads all branches from the repository
func (r *Reader) GetBranches() ([]Branch, error) {
	cmd := exec.Command("git", "for-each-ref", "--format=%(refname:short)", "refs/heads")
	cmd.Dir = r.repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	var branches []Branch
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		branchName := scanner.Text()
		commits, err := r.GetBranchCommits(branchName)
		if err != nil {
			return nil, fmt.Errorf("failed to get commits for branch %s: %w", branchName, err)
		}
		branches = append(branches, Branch{
			Name:    branchName,
			Commits: commits,
		})
	}

	return branches, nil
}

// GetBranchCommits reads all commits for a specific branch
func (r *Reader) GetBranchCommits(branchName string) ([]Commit, error) {
	// Get commit log with SHA, author, email, date, and message
	// Using a delimiter that's unlikely to appear in commit data
	format := "%H%x1f%an%x1f%ae%x1f%ad%x1f%s%x1f%P"
	cmd := exec.Command("git", "log", "--format="+format, branchName)
	cmd.Dir = r.repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	var commits []Commit
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	delimiter := "\x1f"

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Split(line, delimiter)
		if len(parts) < 5 {
			continue
		}

		commit := Commit{
			SHA:     parts[0],
			Author:  parts[1],
			Email:   parts[2],
			Date:    parts[3],
			Message: parts[4],
		}

		// Parse parent commits
		if len(parts) > 5 && parts[5] != "" {
			commit.Parents = strings.Split(parts[5], " ")
		}

		// Parse date to time.Time
		if parsedDate, err := time.Parse("Mon Jan 2 15:04:05 2006 -0700", commit.Date); err == nil {
			commit.DateParsed = parsedDate
		} else if parsedDate, err = time.Parse("2006-01-02 15:04:05", commit.Date); err == nil {
			commit.DateParsed = parsedDate
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

// GetLatestCommit reads the latest commit from the repository
func (r *Reader) GetLatestCommit() (*Commit, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%H%x1f%an%x1f%ae%x1f%ad%x1f%s%x1f%P")
	cmd.Dir = r.repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest commit: %w", err)
	}

	line := strings.TrimSpace(string(output))
	if line == "" {
		return nil, fmt.Errorf("no commits found in repository")
	}

	parts := strings.Split(line, "\x1f")
	if len(parts) < 5 {
		return nil, fmt.Errorf("unexpected commit format")
	}

	commit := &Commit{
		SHA:     parts[0],
		Author:  parts[1],
		Email:   parts[2],
		Date:    parts[3],
		Message: parts[4],
	}

	if len(parts) > 5 && parts[5] != "" {
		commit.Parents = strings.Split(parts[5], " ")
	}

	return commit, nil
}
