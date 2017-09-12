package config

import (
    "os"
    "errors"
    "fmt"
    "sync"
    "strings"
    "path/filepath"
    "io/ioutil"
    "log"
    "flag"
)

const (
    defaultConfigFileName = "contentscraper.yaml"
)

var configFileName string
func init() {
    flag.StringVar(&configFileName, "configfile", defaultConfigFileName, "Configuration file")
}

var configFilePaths []string
var initConfigOnce sync.Once

// Initializes the list of directories that will be searched for the configuration file
func initConfigFilePaths() {
    configFilePaths = append(configFilePaths, ".")
    configFilePaths = append(configFilePaths, "/etc")
    if home, is_present := os.LookupEnv("HOME"); is_present {
        configFilePaths = append(configFilePaths, home + "/.contentscraper")
    }
}

func isFileExist(filepath string) bool {
    if stat, err := os.Stat(filepath); err == nil {
        if !stat.IsDir() {
            return true
        }
    }
    return false
}

// findFileInPaths attempts to locate a file with the given name within the given list of directories.
// If successful, it returns a string with the absolute path to the file.
func findFileInPaths(filename string, paths []string) (fullpath string, err error) {
    for _, path := range paths {
        fullpath = filepath.Join(path, filename)
        if isFileExist(fullpath) {
            return fullpath, nil
        }
    }
    return "", errors.New(
        fmt.Sprintf("File %s not found anywhere in the paths: %s", filename, strings.Join(paths, " ")),
    )
}

// findFile attempts to locate a file within the config paths.
// If an absolute path is given, it skips the config paths and tries to locate that file directly.
func findFile(filename string) (filepath string, err error) {
    initConfigOnce.Do(initConfigFilePaths)
    if len(filename) == 0 {
        log.Fatal("Empty configuration file name")
    }
    if filename[0] == '/' {
        if !isFileExist(filename) {
            return "", errors.New(fmt.Sprintf("File %s not found.", filename))
        }
        return filename, nil
    }

    if filepath, err = findFileInPaths(filename, configFilePaths); err != nil {
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
