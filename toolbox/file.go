package toolbox

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DoesFileExist(filepath string) bool {
	if stat, err := os.Stat(filepath); err == nil {
		if !stat.IsDir() {
			return true
		}
	}
	return false
}

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
	return "", errors.New(
		fmt.Sprintf("File %s not found anywhere in the paths: %s", filename, strings.Join(paths, " ")),
	)
}
