package config

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestConfigParsing(t *testing.T) {
	expected := &Config{
		Reddit_secrets: RedditSecrets{
			ClientId:     "some_client_id",
			ClientSecret: "some_client_secret",
			Username:     "some_reddit_user",
			Password:     "some_password",
		},
		Subreddits: []string{
			"reddit1",
			"reddit2",
		},
		BackendStorePath: filepath.Join(storageDir, databaseFileName),
	}

	conf, err := GetTestConfig()
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(conf, expected) {
		t.Error("Config structure differed from expectation")
	}
}
