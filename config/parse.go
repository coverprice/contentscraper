package config

import (
    "errors"
    "fmt"
    "log"
    //"github.com/davecgh/go-spew/spew"
    "github.com/ghodss/yaml"
)

type Config struct {
    Reddit_secrets RedditSecrets    `json:"redditsecrets"`
    Subreddits []string             `json:"subreddits"`
}

type RedditSecrets struct {
    ClientId                string  `json:"clientid"`
    ClientSecret            string  `json:"clientsecret"`
    Username                string  `json:"username"`
    Password                string  `json:"password"`
}

func parseFromString(configblob string) (conf *Config, err error) {
    conf = &Config{
        Reddit_secrets: RedditSecrets{},
    }

    if err = yaml.Unmarshal([]byte(configblob), conf); err != nil {
        log.Fatal(err)
    }
    /*
    log.Println(spew.Sdump(conf))
    log.Fatal("DONE!")
    */
    return
}

func Parse() (conf *Config, err error) {
    var configFilePath, configFileContents string
    if configFilePath, configFileContents, err = readFile(); err != nil {
        return nil, errors.New(
            fmt.Sprintf("Could not read config file: %v", err),
        )
    }

    if conf, err = parseFromString(configFileContents); err != nil {
        return nil, errors.New(
            fmt.Sprintf("Could not parse config file %s : %v", configFilePath, err),
        )
    }
    return
}
