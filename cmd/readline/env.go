package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	HOME string = "HOME"
)

// Getwd returns path of the working directory in a format suitable as the
// prompt.
func Getwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		return "?"
	}
	return TildeAbbr(pwd)
}

// TildeAbbr abbreviates the user's home directory to ~.
func TildeAbbr(path string) string {
	home, err := GetHome("")
	if home == "" || home == "/" {
		// If home is "" or "/", do not abbreviate because (1) it is likely a
		// problem with the environment and (2) it will make the path actually
		// longer.
		return path
	}
	if err == nil {
		if path == home {
			return "~"
		} else if strings.HasPrefix(path, home+"/") || (runtime.GOOS == "windows" && strings.HasPrefix(path, home+"\\")) {
			return "~" + path[len(home):]
		}
	}
	return path
}

// expand ~ in any string to an absolute path purely for path completion
func ExpandTilde(path string) string {
	if strings.HasPrefix(strings.TrimSpace(path), "~/") {
		home, err := GetHome("")
		if err != nil {
			return path
		}

		return strings.Replace(path, "~/", home+string(filepath.Separator), 1)
	}

	return path
}

// GetHome finds the home directory of a specified user. When given an empty
// string, it finds the home directory of the current user.
func GetHome(uname string) (string, error) {
	if uname == "" {
		// Use $HOME as override if we are looking for the home of the current
		// variable.
		home := os.Getenv(HOME)
		if home != "" {
			if runtime.GOOS == "windows" {
				return strings.TrimRight(home, "/\\"), nil
			} else {
				return strings.TrimRight(home, "/"), nil
			}
		}
	}

	// Look up the user.
	var u *user.User
	var err error
	if uname == "" {
		u, err = user.Current()
	} else {
		u, err = user.Lookup(uname)
	}
	if err != nil {
		return "", fmt.Errorf("can't resolve ~%s: %s", uname, err.Error())
	}
	return strings.TrimRight(u.HomeDir, "/"), nil
}
