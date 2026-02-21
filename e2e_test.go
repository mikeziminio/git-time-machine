package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err, "Failed to get git log")

	logEntries := strings.Split(strings.TrimSpace(string(output)), "\n")

	type entry struct {
		line string
		date time.Time
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
	require.NoError(t, err, "Failed to get git log")
	return len(strings.Split(strings.TrimSpace(string(output)), "\n"))
}

func TestProjectSimple(t *testing.T) {
	inputDir := "testdata/project-simple"
	outputDir := "/tmp/test-output-simple"
	os.RemoveAll(outputDir)

	err := runGitTimeMachine(t, inputDir, outputDir, "--user-name", "New Author")
	require.NoError(t, err)

	log := getGitLog(t, outputDir)
	assert.NotEmpty(t, log)
	assert.Contains(t, log[0], "New Author")
}

func TestTimeSlotConstraint(t *testing.T) {
	inputDir := "testdata/project-time-constrained"
	outputDir := "/tmp/test-output-time"
	os.RemoveAll(outputDir)

	err := runGitTimeMachine(t, inputDir, outputDir, "--date-from", "2023-06-01", "--date-to", "2023-06-02", "--time-from", "10", "--time-to", "12")
	require.NoError(t, err)

	log := getGitLog(t, outputDir)

	for _, entry := range log {
		parts := strings.Split(entry, "|")
		if len(parts) >= 3 {
			timeStr := parts[2]
			tm, err := time.Parse("2006-01-02 15:04:05", timeStr)
			if err != nil {
				continue
			}
			hour := tm.Hour()
			assert.GreaterOrEqual(t, hour, 10, "Hour should be >= 10")
			assert.Less(t, hour, 12, "Hour should be < 12")
		}
	}
}

func TestMinInterval(t *testing.T) {
	inputDir := "testdata/project-simple"
	outputDir := "/tmp/test-output-interval"
	os.RemoveAll(outputDir)

	err := runGitTimeMachine(t, inputDir, outputDir, "--date-from", "2023-01-01", "--date-to", "2023-01-03", "--min-interval", "1")
	require.NoError(t, err)

	log := getGitLog(t, outputDir)

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

	for i := 1; i < len(times); i++ {
		diff := times[i].Sub(times[i-1]).Hours()
		assert.GreaterOrEqual(t, diff, float64(1), "Interval should be at least 1 hour")
	}
}

func TestQuietMode(t *testing.T) {
	inputDir := "testdata/project-simple"
	outputDir := "/tmp/test-output-quiet"
	os.RemoveAll(outputDir)

	cmd := exec.Command("./git-time-machine", "-i", inputDir, "-o", outputDir, "-q")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	outputStr := string(output)
	assert.NotContains(t, outputStr, "Author:")
	assert.NotContains(t, outputStr, "SHA:")
}

func TestInvalidInput(t *testing.T) {
	outputDir := "/tmp/test-output-invalid"
	os.RemoveAll(outputDir)

	err := runGitTimeMachine(t, "/nonexistent/path", outputDir)
	assert.Error(t, err)
}

func TestHelp(t *testing.T) {
	cmd := exec.Command("./git-time-machine", "--help")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	require.NoError(t, err)

	outputStr := string(output)
	assert.Contains(t, outputStr, "Git Time Machine")
	assert.Contains(t, outputStr, "git-time-machine")
}

func TestAuthorAndEmail(t *testing.T) {
	inputDir := "testdata/project-simple"
	outputDir := "/tmp/test-output-author"
	os.RemoveAll(outputDir)

	err := runGitTimeMachine(t, inputDir, outputDir, "--user-name", "John Doe", "--user-email", "john@example.com")
	require.NoError(t, err)

	log := getGitLog(t, outputDir)
	assert.Contains(t, log[0], "John Doe")
	assert.Contains(t, log[0], "john@example.com")
}

func TestInfoModeOnlyInput(t *testing.T) {
	inputDir := "testdata/project-simple"

	cmd := exec.Command("./git-time-machine", "-i", inputDir)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()

	require.NoError(t, err)
	outputStr := string(output)
	assert.Contains(t, outputStr, "Repository Information:")
	assert.Contains(t, outputStr, "Commits:")
	assert.Contains(t, outputStr, "Original commits:")
	assert.Contains(t, outputStr, "Flags will be ignored, as -o is missing")
}

func TestInfoModeWithOtherFlags(t *testing.T) {
	inputDir := "testdata/project-simple"

	cmd := exec.Command("./git-time-machine", "-i", inputDir, "--user-name", "Test User", "--date-from", "2023-01-01")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()

	require.NoError(t, err)
	outputStr := string(output)
	assert.Contains(t, outputStr, "Repository Information:")
	assert.Contains(t, outputStr, "Flags will be ignored, as -o is missing")
}

func TestNoFlagsShowsHelp(t *testing.T) {
	cmd := exec.Command("./git-time-machine")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()

	require.NoError(t, err)
	outputStr := string(output)
	assert.Contains(t, outputStr, "Git Time Machine")
}

func TestMissingRequiredFlagWithError(t *testing.T) {
	outputDir := "/tmp/test-output-error"
	os.RemoveAll(outputDir)

	cmd := exec.Command("./git-time-machine", "-o", outputDir)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()

	assert.Error(t, err)
	outputStr := string(output)
	assert.Contains(t, outputStr, "Error:")
	assert.Contains(t, outputStr, "-i is missing")
}

func TestInvalidPathShowsErrorAndHelp(t *testing.T) {
	outputDir := "/tmp/test-output-invalid-help"
	os.RemoveAll(outputDir)

	cmd := exec.Command("./git-time-machine", "-i", "/nonexistent/path", "-o", outputDir)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()

	assert.Error(t, err)
	outputStr := string(output)
	assert.Contains(t, outputStr, "Error:")
	assert.Contains(t, outputStr, "Usage:")
}

func TestNormalModeBothFlags(t *testing.T) {
	inputDir := "testdata/project-simple"
	outputDir := "/tmp/test-output-normal"
	os.RemoveAll(outputDir)

	err := runGitTimeMachine(t, inputDir, outputDir)
	require.NoError(t, err)

	gitLogCmd := exec.Command("git", "log", "--oneline")
	gitLogCmd.Dir = outputDir
	output, err := gitLogCmd.CombinedOutput()
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	assert.NotEmpty(t, lines)
	assert.NotEmpty(t, lines[0])
}
