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
					Name:                 "foo",
					Description:          "foo foo",
					Media:                "image",
					DefaultPercentile:    87.0,
					DefaultMaxDailyPosts: 103,
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
					Name:                 "bar",
					Description:          "bar bar",
					Media:                "text",
					DefaultPercentile:    85.0,
					DefaultMaxDailyPosts: 100, // The global default
					Subreddits: []Subreddit{
						Subreddit{
							Name:          "subreddit3",
							Percentile:    70.0,
							MaxDailyPosts: 22,
						},
						Subreddit{
							Name: "subreddit4",
							// Percentile + MaxDailyPosts inherited from Feed default
							Percentile:    85.0,
							MaxDailyPosts: 100,
						},
						Subreddit{
							Name:       "subreddit5",
							Percentile: 72.0,
							// MaxDailyPosts inherited from Feed default
							MaxDailyPosts: 100,
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
		t.Error("Config structure differed from expectation: (actual, expected)\n", spew.Sdump(conf), spew.Sdump(expected))
	}
}
