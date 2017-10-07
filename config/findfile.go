package config

import (
	"flag"
	"fmt"
	"github.com/coverprice/contentscraper/toolbox"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	databaseFileName = "sqlite3.db"
)

var defaultConfigFileNames = []string{
	"contentscraper.yaml",
	"contentscraper.yml",
}

var configFileName string    // filename (incl. relative or absolute path)
var configFilePaths []string // If configFileName is relative, this list of paths is searched.
var storageDir string        // Directory where internal files are stored
var defaultPercentile int
var defaultMaxDailyPosts int

func init() {
	flag.StringVar(&configFileName, "config", "", fmt.Sprintf(
		"Configuration file (default: %s)", strings.Join(defaultConfigFileNames, ", ")))
	flag.IntVar(&defaultPercentile, "default-percentile", 80, "Default filtering percentile")
	flag.IntVar(&defaultMaxDailyPosts, "default-max-daily-posts", 100, "Default maximum daily posts")

	storageBaseEnvVar := "HOME"
	storageDirName := ".contentscraper"
	if runtime.GOOS == "windows" {
		storageBaseEnvVar = "LOCALAPPDATA"
		storageDirName = "ContentScraper"
	}
	storageBase, is_present := os.LookupEnv(storageBaseEnvVar)
	if !is_present {
		panic(fmt.Sprintf("Environment variable '%s' not set", storageBaseEnvVar))
	}
	storageDir = filepath.Join(storageBase, storageDirName)
	if !toolbox.DoesDirExist(storageDir) {
		if err := os.MkdirAll(storageDir, 0x755); err != nil {
			panic(fmt.Sprintf("Could not make the directory '%s'", storageDir))
		}
	}

	configFilePaths = append(configFilePaths, storageDir)
}

func StorageDir() string {
	return storageDir
}

func locateConfigFile() (filePath string, err error) {
	if configFileName == "" {
		// means use the default name & search paths
		for _, filename := range defaultConfigFileNames {
			if filePath, err = toolbox.FindFileInPaths(filename, configFilePaths); err == nil {
				return filePath, nil
			}
		}
		return "", err
	} else {
		if !toolbox.DoesFileExist(configFileName) {
			err = fmt.Errorf("Config file '%s' does not exist or insufficient permissions", configFileName)
			return
		}
		return configFileName, nil
	}
}
