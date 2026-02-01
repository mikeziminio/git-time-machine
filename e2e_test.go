package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func runGitTimeMachine(t *testing.T, inputDir, outputDir string, args ...string) error {
	cmd := exec.Command("./git-time-machine", append([]string{"-i", inputDir, "-o", outputDir}, args...)...)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	t.Logf("Command output: %s", string(output))
	return err
}

func getGitLog(t *testing.T, dir string) []string {
	cmd := exec.Command("git", "log", "--format=%H|%an|%ae|%ad|%s", "--date=format:%Y-%m-%d %H:%M:%S")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get git log: %v", err)
	}
	
	logEntries := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	// Parse dates and sort by date (chronological order)
	type entry struct {
		line   string
		date   time.Time
	}
	entries := make([]entry, 0, len(logEntries))
	for _, line := range logEntries {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			tm, err := time.Parse("2006-01-02 15:04:05", parts[3])
			if err == nil {
				entries = append(entries, entry{line: line, date: tm})
			}
		}
	}
	
	// Sort by date (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].date.Before(entries[i].date) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	
	result := make([]string, len(entries))
	for i, e := range entries {
		result[i] = e.line
	}
	return result
}

func getGitLogCount(t *testing.T, dir string) int {
	cmd := exec.Command("git", "log", "--oneline")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get git log: %v", err)
	}
	return len(strings.Split(strings.TrimSpace(string(output)), "\n"))
}

func TestProjectSimple(t *testing.T) {
	inputDir := "testdata/project-simple"
	outputDir := "/tmp/test-output-simple"
	os.RemoveAll(outputDir)

	// Run with author change
	err := runGitTimeMachine(t, inputDir, outputDir, "--user-name", "New Author")
	if err != nil {
		t.Fatalf("Failed to run git-time-machine: %v", err)
	}

	log := getGitLog(t, outputDir)
	if len(log) == 0 {
		t.Fatal("No commits in output repository")
	}

	// Check author was changed
	if !strings.Contains(log[0], "New Author") {
		t.Errorf("Author not changed. First commit: %s", log[0])
	}
}

func TestTimeSlotConstraint(t *testing.T) {
	inputDir := "testdata/project-time-constrained"
	outputDir := "/tmp/test-output-time"
	os.RemoveAll(outputDir)

	// Run with time slot 10:00-12:00
	err := runGitTimeMachine(t, inputDir, outputDir, "--date-from", "2023-06-01", "--date-to", "2023-06-02", "--time-from", "10", "--time-to", "12")
	if err != nil {
		t.Fatalf("Failed to run git-time-machine: %v", err)
	}

	log := getGitLog(t, outputDir)
	
	// Parse times and check they are within the slot
	for _, entry := range log {
		parts := strings.Split(entry, "|")
		if len(parts) >= 3 {
			timeStr := parts[2]
			tm, err := time.Parse("2006-01-02 15:04:05", timeStr)
			if err != nil {
				t.Logf("Failed to parse time %s: %v", timeStr, err)
				continue
			}
			hour := tm.Hour()
			if hour < 10 || hour >= 12 {
				t.Errorf("Time %s is outside 10:00-12:00 slot", timeStr)
			}
		}
	}
}

func TestMinInterval(t *testing.T) {
	inputDir := "testdata/project-simple"
	outputDir := "/tmp/test-output-interval"
	os.RemoveAll(outputDir)

	// Run with min-interval of 1 hour on a 1-day range
	err := runGitTimeMachine(t, inputDir, outputDir, "--date-from", "2023-01-01", "--date-to", "2023-01-03", "--min-interval", "1")
	if err != nil {
		t.Fatalf("Failed to run git-time-machine: %v", err)
	}

	log := getGitLog(t, outputDir)
	
	// Parse times and check intervals
	times := make([]time.Time, 0)
	for _, entry := range log {
		parts := strings.Split(entry, "|")
		if len(parts) >= 3 {
			tm, err := time.Parse("2006-01-02 15:04:05", parts[2])
			if err == nil {
				times = append(times, tm)
			}
		}
	}

	// Check minimum 1 hour intervals
	for i := 1; i < len(times); i++ {
		diff := times[i].Sub(times[i-1]).Hours()
		if diff < 1 {
			t.Errorf("Interval of %.1f hours is less than required 1 hour", diff)
		}
	}
}

func TestQuietMode(t *testing.T) {
	inputDir := "testdata/project-simple"
	outputDir := "/tmp/test-output-quiet"
	os.RemoveAll(outputDir)

	// Run in quiet mode
	cmd := exec.Command("./git-time-machine", "-i", inputDir, "-o", outputDir, "-q")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run git-time-machine: %v", err)
	}
	
	outputStr := string(output)
	// Quiet mode should not show detailed commit info
	if strings.Contains(outputStr, "Author:") || strings.Contains(outputStr, "SHA:") {
		t.Errorf("Quiet mode should not show commit details. Output: %s", outputStr)
	}
}

func TestInvalidInput(t *testing.T) {
	outputDir := "/tmp/test-output-invalid"
	os.RemoveAll(outputDir)

	err := runGitTimeMachine(t, "/nonexistent/path", outputDir)
	if err == nil {
		t.Error("Should fail with invalid input path")
	}
}

func TestHelp(t *testing.T) {
	cmd := exec.Command("./git-time-machine", "--help")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run git-time-machine --help: %v", err)
	}
	
	outputStr := string(output)
	if !strings.Contains(outputStr, "Git Time Machine") || !strings.Contains(outputStr, "git-time-machine") {
		t.Error("Help message not displayed correctly")
	}
}

func TestAuthorAndEmail(t *testing.T) {
	inputDir := "testdata/project-simple"
	outputDir := "/tmp/test-output-author"
	os.RemoveAll(outputDir)

	err := runGitTimeMachine(t, inputDir, outputDir, "--user-name", "John Doe", "--user-email", "john@example.com")
	if err != nil {
		t.Fatalf("Failed to run git-time-machine: %v", err)
	}

	log := getGitLog(t, outputDir)
	
	// Check both author and email were changed
	if !strings.Contains(log[0], "John Doe") {
		t.Errorf("Author name not changed")
	}
	if !strings.Contains(log[0], "john@example.com") {
		t.Errorf("Email not changed")
	}
}
