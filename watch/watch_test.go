package watch

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchTriggersBuild(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(dir)

	os.Mkdir("models", 0755)
	file := filepath.Join("models", "a.cp")
	os.WriteFile(file, []byte("initial"), 0644)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	builds := 0
	go func() {
		Watch(ctx, func() error { builds++; return nil })
	}()

	time.Sleep(200 * time.Millisecond)
	os.WriteFile(file, []byte("changed"), 0644)

	for i := 0; i < 20; i++ {
		if builds > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	cancel()

	if builds == 0 {
		t.Fatal("expected build to be triggered")
	}
}
