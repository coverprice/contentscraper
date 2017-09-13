package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/coverprice/contentscraper/toolbox"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	defaultConfigFileName = "contentscraper.yaml"
	databaseFileName      = "sqlite3.db"
)

var configFileName string    // filename (incl. relative or absolute path)
var configFilePaths []string // If configFileName is relative, this list of paths is searched.
var storageDir string        // Directory where internal files are stored

func init() {
	flag.StringVar(&configFileName, "configfile", defaultConfigFileName, "Configuration file")

	configFilePaths = append(configFilePaths, ".")
	configFilePaths = append(configFilePaths, "/etc")

	home_dir, is_present := os.LookupEnv("HOME")
	if !is_present {
		panic("HOME environment variable not set")
	}
	storageDir = filepath.Join(home_dir, ".contentscraper")

	if !toolbox.DoesDirExist(storageDir) {
		if err := os.MkdirAll(storageDir, 0x755); err != nil {
			panic(fmt.Sprintf("Could not make the directory %s", storageDir))
		}
	}
	configFilePaths = append(configFilePaths, storageDir)
}

// findFile attempts to locate a file within the config paths.
// If an absolute path is given, it skips the config paths and tries to locate that file directly.
func findFile(filename string) (filepath string, err error) {
	if len(filename) == 0 {
		panic("Empty configuration file name")
	}
	if filename[0] == '/' {
		if !toolbox.DoesFileExist(filename) {
			return "", errors.New(fmt.Sprintf("File %s not found.", filename))
		}
		return filename, nil
	}

	if filepath, err = toolbox.FindFileInPaths(filename, configFilePaths); err != nil {
		return
	}
	return
}

// readFile searches for the config file, and reads its contents into memory.
func readFile() (file_path string, contents string, err error) {
	if file_path, err = findFile(configFileName); err != nil {
		return "", "", err
	}
	var raw_contents []byte
	if raw_contents, err = ioutil.ReadFile(file_path); err != nil {
		return "", "", err
	}
	contents = string(raw_contents)
	return
}
