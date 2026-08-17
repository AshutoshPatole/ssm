package cmd

import (
	"path/filepath"
	"strings"
	"testing"
)

func validateArchivePath(destDir, entryName string) bool {
	if filepath.IsAbs(entryName) || strings.HasPrefix(entryName, "/") || strings.HasPrefix(entryName, "\\") {
		return false
	}
	cleanDest := filepath.Clean(destDir) + string(filepath.Separator)
	targetPath := filepath.Clean(filepath.Join(destDir, entryName))
	return strings.HasPrefix(targetPath, cleanDest) || targetPath == filepath.Clean(destDir)
}

func TestPathTraversalProtection(t *testing.T) {
	destDir := filepath.Join("tmp", "test-extract")

	testCases := []struct {
		name      string
		entryName string
		isSafe    bool
	}{
		{"normal binary", "ssm", true},
		{"nested normal file", "bin/ssm", true},
		{"parent directory traversal", "../../../etc/passwd", false},
		{"sibling directory traversal", "../evil", false},
		{"rooted Unix path", "/etc/shadow", false},
		{"rooted Windows path", "\\Windows\\System32\\calc.exe", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isSafe := validateArchivePath(destDir, tc.entryName)
			if isSafe != tc.isSafe {
				t.Errorf("path %q validation mismatch: got safe=%v, want safe=%v", tc.entryName, isSafe, tc.isSafe)
			}
		})
	}
}
