package main

import (
	"os"
	"path/filepath"
)

func lookPath() ([]string, []string) {
	path := os.Getenv("PATH")
	var dirs []string
	var exes []string
	for _, dir := range filepath.SplitList(path) {
		if dir == "" {
			dir = "."
		}

		dirs = append(dirs, dir)

		dirent, err := os.ReadDir(dir)
		if err != nil {
		}

		for x := range dirent {
			exe := dirent[x].Name()
			exes = append(exes, exe)
		}
	}

	return dirs, exes
}
