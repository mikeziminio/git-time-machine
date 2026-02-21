package writer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyFiles(t *testing.T) {
	inputDir := "/tmp/test-copy-input"
	outputDir := "/tmp/test-copy-output"

	os.RemoveAll(inputDir)
	os.RemoveAll(outputDir)
	defer os.RemoveAll(inputDir)
	defer os.RemoveAll(outputDir)

	require.NoError(t, os.MkdirAll(filepath.Join(inputDir, "dir1", "subdir"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(inputDir, ".git"), 0755))

	require.NoError(t, os.WriteFile(filepath.Join(inputDir, "file1.txt"), []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(inputDir, "dir1", "file2.txt"), []byte("content2"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(inputDir, "dir1", "subdir", "file3.txt"), []byte("content3"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(inputDir, ".git", "config"), []byte("git config"), 0644))

	w := NewWriter(outputDir, true)
	require.NoError(t, w.InitRepository())

	err := w.CopyFiles(inputDir)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "file1.txt"))
	assert.FileExists(t, filepath.Join(outputDir, "dir1", "file2.txt"))
	assert.FileExists(t, filepath.Join(outputDir, "dir1", "subdir", "file3.txt"))

	_, err = os.Stat(filepath.Join(outputDir, ".git", "config"))
	if err == nil {
		content, _ := os.ReadFile(filepath.Join(outputDir, ".git", "config"))
		assert.NotEqual(t, "git config", string(content), ".git/config from input should not be copied")
	}
}
