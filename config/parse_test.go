package config

import (
	"github.com/davecgh/go-spew/spew"
	"path/filepath"
	"reflect"
	"testing"
)

func TestConfigParsing(t *testing.T) {
	expected := &Config{
		Reddit: RedditConfig{
			Secrets: RedditSecrets{
				ClientId:     "some_client_id",
				ClientSecret: "some_client_secret",
				Username:     "some_reddit_user",
				Password:     "some_password",
			},
			Feeds: []RedditFeed{
				RedditFeed{
					Name:        "foo",
					Description: "foo foo",
					Subreddits: []Subreddit{
						Subreddit{
							Name:          "subreddit1",
							Percentile:    80.0,
							MaxDailyPosts: 100,
						},
						Subreddit{
							Name:          "subreddit2",
							Percentile:    90.0,
							MaxDailyPosts: 73,
						},
					},
				},
				RedditFeed{
					Name:        "bar",
					Description: "bar bar",
					Subreddits: []Subreddit{
						Subreddit{
							Name:          "subreddit3",
							Percentile:    70.0,
							MaxDailyPosts: 22,
						},
					},
				},
			},
		},
		Twitter: TwitterConfig{
			Secrets: TwitterSecrets{
				ClientSecret: "twitterapikey",
			},
			Feeds: []TwitterFeed{
				TwitterFeed{
					Name:        "news",
					Description: "news tweets",
					Filters: []TwitterFilter{
						TwitterFilter{
							AccountName:   "mhaberman",
							FilterType:    "original",
							Percentile:    70.0,
							MaxDailyPosts: 22,
						},
						TwitterFilter{
							AccountName:   "mhaberman",
							FilterType:    "retweets",
							Percentile:    50.0,
							MaxDailyPosts: 44,
						},
					},
				},
			},
		},
		BackendStorePath: filepath.Join(storageDir, databaseFileName),
	}

	conf, err := GetTestConfig()
	if err != nil {
		t.Error(err)
	}

	if err := conf.Validate(); err != nil {
		t.Error("Did not validate", err)
	}
	if !reflect.DeepEqual(conf, expected) {
		t.Error("Config structure differed from expectation", spew.Sdump(conf), spew.Sdump(expected))
	}
}
