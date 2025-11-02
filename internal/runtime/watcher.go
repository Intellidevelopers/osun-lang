package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func WatchAndRun(targetFile string, runFunc func(string)) {
	lastMods := make(map[string]time.Time)

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".os" {
			lastMods[path] = info.ModTime()
		}
		return nil
	})

	fmt.Println("👀 Watching all .os files...")

	for {
		time.Sleep(1 * time.Second)
		changed := false

		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && filepath.Ext(path) == ".os" {
				if info.ModTime() != lastMods[path] {
					changed = true
					lastMods[path] = info.ModTime()
				}
			}
			return nil
		})

		if changed {
			fmt.Println("♻️  Change detected, restarting Osun script...")
			data, _ := os.ReadFile(targetFile)
			runFunc(string(data))
		}
	}
}


func getModTime(file string) time.Time {
	info, err := os.Stat(file)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}
