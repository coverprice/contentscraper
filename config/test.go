package config

import (
	"fmt"
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

func GetTestConfig() (conf *Config, err error) {
	if conf, err = parseFromString(testConfig); err != nil {
		return nil, fmt.Errorf("Could not parse fake config file: %v", err)
	}
	return
}
