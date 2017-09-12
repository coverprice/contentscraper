package config

import (
	"testing"
	"reflect"
)

func TestConfigParsing(t *testing.T) {
	expected := &Config{
		Reddit_secrets: RedditSecrets{
			ClientId: "some_client_id",
			ClientSecret: "some_client_secret",
			Username: "some_reddit_user",
			Password: "some_password",
		},
		Subreddits: []string{
			"reddit1",
			"reddit2",
		},
	}

	conf, err := parseFromString(testConfig)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(conf, expected) {
		t.Error("Config structure differed from expectation")
	}
}

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
