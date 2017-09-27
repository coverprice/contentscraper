package config

import (
	"flag"
	"fmt"
	"github.com/coverprice/contentscraper/toolbox"
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
var defaultPercentile int
var defaultMaxDailyPosts int

func init() {
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

	flag.StringVar(&configFileName, "configfile", "", "Configuration file")
	flag.IntVar(&defaultPercentile, "default-percentile", 80, "Default filtering percentile")
	flag.IntVar(&defaultMaxDailyPosts, "default-max-daily-posts", 100, "Default maximum daily posts")
	configFilePaths = append(configFilePaths, ".")
	configFilePaths = append(configFilePaths, "/etc")
	configFilePaths = append(configFilePaths, storageDir)
}

func locateConfigFile() (filePath string, err error) {
	if configFileName == "" {
		// means use the default name & search paths
		return toolbox.FindFileInPaths(defaultConfigFileName, configFilePaths)
	} else {
		if !toolbox.DoesFileExist(configFileName) {
			err = fmt.Errorf("Config file '%s' does not exist or insufficient permissions", configFileName)
			return
		}
		return configFileName, nil
	}
}
