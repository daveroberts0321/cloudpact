package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindCloudPactFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.cp"), []byte(""), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "generated"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "generated", "b.cp"), []byte(""), 0644); err != nil {
		t.Fatalf("write generated file: %v", err)
	}

	files, err := FindCloudPactFiles(dir)
	if err != nil {
		t.Fatalf("FindCloudPactFiles error: %v", err)
	}
	expected := filepath.Join(dir, "a.cp")
	if len(files) != 1 || files[0] != expected {
		t.Fatalf("unexpected files: %v", files)
	}
}
