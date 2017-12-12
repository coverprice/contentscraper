package toolbox

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DoesFileExist returns true if the given file exists and is not a directory.
func DoesFileExist(filepath string) bool {
	if stat, err := os.Stat(filepath); err == nil {
		if !stat.IsDir() {
			return true
		}
	}
	return false
}

// DoesDirExist returns true if the given path exists and is a directory.
func DoesDirExist(dirpath string) bool {
	if stat, err := os.Stat(dirpath); err == nil {
		return stat.IsDir()
	}
	return false
}

// FindFileInPaths attempts to locate a file with the given name within the given list of directories.
// If successful, it returns a string with the absolute path to the file.
func FindFileInPaths(filename string, paths []string) (fullpath string, err error) {
	for _, path := range paths {
		fullpath = filepath.Join(path, filename)
		if DoesFileExist(fullpath) {
			return fullpath, nil
		}
	}
	return "", fmt.Errorf("File %s not found anywhere in the paths: %s", filename, strings.Join(paths, ", "))
}
