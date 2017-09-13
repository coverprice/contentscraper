package config

import (
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/ghodss/yaml"
	"path/filepath"
)

const testConfig = `
redditsecrets:
    clientid: "some_client_id"
    clientsecret: "some_client_secret"
    username: "some_reddit_user"
    password: "some_password"

subreddits:
- "reddit1"
- "reddit2"
`

type Config struct {
	Reddit_secrets   RedditSecrets `json:"redditsecrets"`
	Subreddits       []string      `json:"subreddits"`
	BackendStorePath string        // Path to the database file
}

type RedditSecrets struct {
	ClientId     string `json:"clientid"`
	ClientSecret string `json:"clientsecret"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

func parseFromString(configblob string) (conf *Config, err error) {
	conf = &Config{
		Reddit_secrets:   RedditSecrets{},
		BackendStorePath: filepath.Join(storageDir, databaseFileName),
	}

	if err = yaml.Unmarshal([]byte(configblob), conf); err != nil {
		return
	}

	/*
	   log.Println(spew.Sdump(conf))
	   log.Fatal("DONE!")
	*/
	return
}

func GetConfigFromFile() (conf *Config, err error) {
	var configFilePath, configFileContents string
	if configFilePath, configFileContents, err = readFile(); err != nil {
		return nil, fmt.Errorf("Could not read config file: %v", err)
	}

	if conf, err = parseFromString(configFileContents); err != nil {
		return nil, fmt.Errorf("Could not parse config file %s : %v", configFilePath, err)
	}
	return
}

func GetTestConfig() (conf *Config, err error) {
	if conf, err = parseFromString(testConfig); err != nil {
		return nil, fmt.Errorf("Could not parse fake config file: %v", err)
	}
	return
}
