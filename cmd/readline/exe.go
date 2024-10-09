package main

import (
	"os"
	"path/filepath"
)

func IsExecOwner(mode os.FileMode) bool {
	return mode&0100 != 0
}

// Similarly for telling if executable by the group, use bitmask 0010:
func IsExecGroup(mode os.FileMode) bool {
	return mode&0010 != 0
}

// And by others, use bitmask 0001:
func IsExecOther(mode os.FileMode) bool {
	return mode&0001 != 0
}

// To tell if the file is executable by any of the above, use bitmask 0111:
func IsExecAny(mode os.FileMode) bool {
	return mode&0111 != 0
}

// To tell if the file is executable by all of the above, again use bitmask 0111 but check if the result equals to 0111:
func IsExecAll(mode os.FileMode) bool {
	return mode&0111 == 0111
}

// populate the valid binaries for tab completion
func (a *ash) findExes() {
	a.PathDirs, a.Executables = lookPath()
}

func lookPath() (dirs []string, exes []string) {
	path := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(path) {
		if dir == "" {
			dir = "."
		}

		dirent, err := os.ReadDir(dir)
		if err == nil {
			dirs = append(dirs, dir)
		} else {
			continue
		}

		for i := range dirent {
			file := dirent[i]
			fi, _ := file.Info()

			if IsExecAny(fi.Mode()) {
				exe := file.Name()
				exes = append(exes, exe)
			}
		}
	}

	return dirs, exes
}
