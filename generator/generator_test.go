package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateRecord(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(dir)

	GenerateRecord("User")

	content, err := os.ReadFile(filepath.Join("models", "user.cp"))
	if err != nil {
		t.Fatalf("expected generated file, got error: %v", err)
	}
	if !strings.Contains(string(content), "record User") {
		t.Fatalf("unexpected content: %s", string(content))
	}
}
