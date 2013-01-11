package main

import (
	"os"
	"path/filepath"
	"strings"
)

// getPathDirs gets the user's GOPATH environment variable and splits it along
// the OS-specific path separator, returning a slice of strings, one per path
// directory.
func getPathDirs() []string {
	path := os.Getenv("GOPATH")
	if path == "" {
		quit(1, "GOPATH is not set")
	}
	return strings.Split(path, string(os.PathListSeparator))
}

// getPkgDirCandidates gets the list of possible locations for a given import
// string.  The paths are not guaranteed to exist or to even be valid; results
// are derived from the user's $GOPATH environment variable, which is not
// necessarily clean.
func getPkgDirCandidates(importPath string) []string {
	pathDirs := getPathDirs()
	candidates := make([]string, 0, len(pathDirs))
	for _, pathDir := range pathDirs {
		candidates = append(candidates, filepath.Join(append(
			[]string{pathDir, "src", "pkg"},
			strings.Split(importPath, "/")...,
		)...))
		candidates = append(candidates, filepath.Join(append(
			[]string{pathDir, "src"},
			strings.Split(importPath, "/")...,
		)...))
	}
	return candidates
}

// existingDirectories accepts a slice of strings and returns a slice of
// strings representing which of the given strings corresponds to an existing
// directory.  Note that this is prone to timing attacks, but it is presumed to
// not matter for this application; this may be an unsafe strategy to copy into
// other projects.
func existingDirectories(candidates []string) []string {
	dirs := make([]string, 0, len(candidates))
	for _, path := range candidates {
		if isDir(path) {
			dirs = append(dirs, path)
		}
	}
	return dirs
}

// isDir takes a file path and returns a boolean representing whether the path
// is or is not a valid directory.  Again, this is vulnerable to timing attacks
// and should be considered a Very Bad Idea in most contexts.
func isDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

// getPkgDir takes an import path and returns a string representing the
// directory path of the package on the current machine.  This is generally a
// stupid thing to do, because we're generally not worried about the location
// of a package's source code files, since they may be out of sync with the
// actual binary, but in our case, we're using it to look up assets that have
// been made go-gettable.  If no package dir can be found, an empty string is
// returned.  The package may reasonably be installed into multiple workspaces.
// In this case, it's the first package found, as dictated by the user's
// $GOPATH environment variable.
func getPkgDir(importPath string) string {
	dirs := existingDirectories(getPkgDirCandidates(importPath))
	if len(dirs) == 0 {
		return ""
	}
	return dirs[0]
}
