package watch

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func Watch(ctx context.Context, build func() error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	watchDirs := []string{"models", "services"}
	for _, dir := range watchDirs {
		if _, err := os.Stat(dir); err == nil {
			if err := watcher.Add(dir); err != nil {
				return err
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if strings.HasSuffix(event.Name, ".cp") && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
				fmt.Printf("File changed: %s\n", event.Name)
				if err := build(); err != nil {
					fmt.Printf("Build failed: %v\n", err)
				} else {
					fmt.Println("Rebuild complete")
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}
