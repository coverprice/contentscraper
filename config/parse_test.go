package config

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"path/filepath"
	"reflect"
	"testing"
)

const testConfig = `
reddit:
    secrets:
        clientid: "some_client_id"
        clientsecret: "some_client_secret"
        username: "some_reddit_user"
        password: "some_password"

    feeds:
        - name: "foo"
          description: "foo foo"
          media: "image"
          percentile: 87.0
          max_daily_posts: 103
          subreddits:
            - name: "subreddit1"
              percentile: 80.0
              max_daily_posts: 100
            - name: "subreddit2"
              percentile: 90
              max_daily_posts: 73 
        - name: "bar"
          description: "bar bar"
          media: "text"
          percentile: 85.0
          subreddits:
            - name: "subreddit3"
              percentile: 70.0
              max_daily_posts: 22
            - name: "subreddit4"
            - name: "subreddit5"
              percentile: 72.0

twitter:
    secrets:
        clientsecret: "twitterapikey"
    feeds:
        - name: "news"
          description: "news tweets"
          filters:
            - account: "mhaberman"
              filtertype: "original"
              percentile: 70.0
              max_daily_posts: 22
            - account: "mhaberman"
              filtertype: "retweets"
              percentile: 50.0
              max_daily_posts: 44
`

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

	var conf *Config
	var err error
	if conf, err = parseFromString(testConfig); err != nil {
		t.Error(fmt.Errorf("Could not parse fake config file: %v", err))
	}
	conf.populateDefaults()

	if err = conf.Validate(); err != nil {
		t.Error("Did not validate", err)
	}
	if !reflect.DeepEqual(conf, expected) {
		t.Error("Config structure differed from expectation: (actual, expected)\n", spew.Sdump(conf), spew.Sdump(expected))
	}
}
