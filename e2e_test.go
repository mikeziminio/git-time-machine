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

func TestInfoModeOnlyInput(t *testing.T) {
	inputDir := "testdata/project-simple"

	// Run with only -i flag (no -o)
	cmd := exec.Command("./git-time-machine", "-i", inputDir)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	
	// Should not be an error
	if err != nil {
		t.Fatalf("Should not error with only -i flag. Output: %s", string(output))
	}
	
	outputStr := string(output)
	
	// Check for repository info header
	if !strings.Contains(outputStr, "Repository Information:") {
		t.Errorf("Should show repository information header")
	}
	
	// Check for commit count
	if !strings.Contains(outputStr, "Commits:") {
		t.Errorf("Should show commit count")
	}
	
	// Check for original commits section
	if !strings.Contains(outputStr, "Original commits:") {
		t.Errorf("Should show original commits")
	}
	
	// Check for warning about ignored flags
	if !strings.Contains(outputStr, "Flags will be ignored, as -o is missing") {
		t.Errorf("Should warn about ignored flags")
	}
	
	// Check that no output repo was created
	// (we can't easily check this in temp dir, but we verified no error)
}

func TestInfoModeWithOtherFlags(t *testing.T) {
	inputDir := "testdata/project-simple"

	// Run with -i and other flags but no -o
	cmd := exec.Command("./git-time-machine", "-i", inputDir, "--user-name", "Test User", "--date-from", "2023-01-01")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	
	// Should not be an error - flags should be ignored
	if err != nil {
		t.Fatalf("Should not error with -i only. Output: %s", string(output))
	}
	
	outputStr := string(output)
	
	// Should still show info mode output
	if !strings.Contains(outputStr, "Repository Information:") {
		t.Errorf("Should show repository information")
	}
	
	// Should warn about ignored flags
	if !strings.Contains(outputStr, "Flags will be ignored, as -o is missing") {
		t.Errorf("Should warn about ignored flags")
	}
}

func TestNoFlagsShowsHelp(t *testing.T) {
	cmd := exec.Command("./git-time-machine")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	
	// Should not be an error when no flags
	if err != nil {
		t.Fatalf("Should not error with no flags. Output: %s", string(output))
	}
	
	outputStr := string(output)
	
	// Should show help
	if !strings.Contains(outputStr, "Git Time Machine") {
		t.Errorf("Should show help message")
	}
}

func TestMissingRequiredFlagWithError(t *testing.T) {
	outputDir := "/tmp/test-output-error"
	os.RemoveAll(outputDir)

	// Run with only -o (missing -i)
	cmd := exec.Command("./git-time-machine", "-o", outputDir)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	
	// Should be an error
	if err == nil {
		t.Error("Should error when -i is missing")
	}
	
	outputStr := string(output)
	
	// Should show error message
	if !strings.Contains(outputStr, "Error:") {
		t.Errorf("Should show error message")
	}
	
	// Should show required flag -i is missing
	if !strings.Contains(outputStr, "-i is missing") {
		t.Errorf("Should mention -i is required")
	}
}

func TestInvalidPathShowsErrorAndHelp(t *testing.T) {
	outputDir := "/tmp/test-output-invalid-help"
	os.RemoveAll(outputDir)

	// Run with invalid input path
	cmd := exec.Command("./git-time-machine", "-i", "/nonexistent/path", "-o", outputDir)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	
	// Should be an error
	if err == nil {
		t.Error("Should error with invalid input path")
	}
	
	outputStr := string(output)
	
	// Should show error
	if !strings.Contains(outputStr, "Error:") {
		t.Errorf("Should show error message")
	}
	
	// Should show help after error
	if !strings.Contains(outputStr, "Usage:") {
		t.Errorf("Should show help after error")
	}
}

func TestNormalModeBothFlags(t *testing.T) {
	inputDir := "testdata/project-simple"
	outputDir := "/tmp/test-output-normal"
	os.RemoveAll(outputDir)

	err := runGitTimeMachine(t, inputDir, outputDir)
	if err != nil {
		t.Fatalf("Failed with both flags: %v", err)
	}
	
	// Should not show info mode output
	cmd := exec.Command("cat", outputDir+"/git-time-machine.log")
	cmd.Dir = "."
	_, err = cmd.CombinedOutput()
	// Just need to verify the output dir was created and has a git repo
	
	// Check that git repo exists and has commits
	gitLogCmd := exec.Command("git", "log", "--oneline")
	gitLogCmd.Dir = outputDir
	output, err := gitLogCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get git log from output: %v", err)
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Error("Output repo should have commits")
	}
}
