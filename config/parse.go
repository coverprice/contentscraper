package config

import (
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"path/filepath"
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
          subreddits:
            - name: "subreddit1"
              percentile: 80.0
              max_daily_posts: 100
            - name: "subreddit2"
              percentile: 90
              max_daily_posts: 73 
        - name: "bar"
          description: "bar bar"
          subreddits:
            - name: "subreddit3"
              percentile: 70.0
              max_daily_posts: 22

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

// -----------------------------

type Config struct {
	Reddit           RedditConfig  `json:"reddit"`
	Twitter          TwitterConfig `json:"twitter"`
	BackendStorePath string        // Path to the database file
}

type RedditConfig struct {
	Secrets RedditSecrets `json:"secrets"`
	Feeds   []RedditFeed  `json:"feeds"`
}

type RedditSecrets struct {
	ClientId     string `json:"clientid"`
	ClientSecret string `json:"clientsecret"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

type RedditFeed struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Subreddits  []Subreddit `json:"subreddits"`
}

type Subreddit struct {
	Name          string  `json:"name"`
	Percentile    float64 `json:"percentile"`
	MaxDailyPosts int     `json:"max_daily_posts"`
}

type TwitterConfig struct {
	Secrets TwitterSecrets `json:"secrets"`
	Feeds   []TwitterFeed  `json:"feeds"`
}

type TwitterSecrets struct {
	ClientSecret string `json:"clientsecret"`
}

type TwitterFeed struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Filters     []TwitterFilter `json:"filters"`
}

type TwitterFilter struct {
	AccountName   string  `json:"account"`
	FilterType    string  `json:"filtertype"`
	Percentile    float64 `json:"percentile"`
	MaxDailyPosts int     `json:"max_daily_posts"`
}

func (this TwitterFilter) Validate() (err error) {
	if this.AccountName == "" {
		return fmt.Errorf("Empty account name")
	}
	if !(this.FilterType == "original" || this.FilterType == "retweets") {
		return fmt.Errorf("Invalid filter type: '%s'", this.FilterType)
	}
	return nil
}

func (this *Config) Validate() (err error) {
	// validation is mainly concerned with ensuring that all feed names are unique
	// across source types, that percentile values are within range, etc.

	var feednames = make(map[string]bool)
	var subredditnames = make(map[string]bool)

	log.Debug("Validating config file")
	for idx, redditFeed := range this.Reddit.Feeds {
		var feedname = redditFeed.Name
		var feederr_template = fmt.Sprintf("Problem in Reddit feed '%s', index %d ", feedname, idx+1)
		if redditFeed.Name == "" {
			return fmt.Errorf("%s: Empty feed name", feederr_template)
		}
		if redditFeed.Description == "" {
			return fmt.Errorf("%s: Empty feed description", feederr_template)
		}
		if _, is_present := feednames[feedname]; is_present {
			return fmt.Errorf("%s: Duplicate name detected. Feed names must be globally unique.", feederr_template)
		}
		feednames[feedname] = true

		for sub_idx, subreddit := range redditFeed.Subreddits {
			var subredditerr_template = fmt.Sprintf("%s, subreddit: '%s' (index %d) ", feederr_template, subreddit.Name, sub_idx+1)
			if subreddit.Name == "" {
				return fmt.Errorf("%s: Empty subreddit name", subredditerr_template)
			}
			if _, is_present := subredditnames[subreddit.Name]; is_present {
				return fmt.Errorf("%s: Duplicate subreddit name. Subreddits must be unique across feeds.", subredditerr_template)
			}
			if subreddit.Percentile < 0.0 || subreddit.Percentile > 100.0 {
				return fmt.Errorf("%s: Percentile out of 0-100 range. : %f", subredditerr_template, subreddit.Percentile)
			}
			if subreddit.MaxDailyPosts < 0 {
				return fmt.Errorf("%s: MaxDailyPosts must be a +ve integer. : %d", subredditerr_template, subreddit.MaxDailyPosts)
			}
		}
	}

	for idx, twitterFeed := range this.Twitter.Feeds {
		var feedname = twitterFeed.Name
		var feederr_template = fmt.Sprintf("Problem in Twitter feed #%d ", idx+1)
		if twitterFeed.Name == "" {
			return fmt.Errorf("%s: Empty feed name", feederr_template)
		}
		if twitterFeed.Description == "" {
			return fmt.Errorf("%s: Empty feed description", feederr_template)
		}
		feederr_template = fmt.Sprintf("%s, name: '%s' ", feederr_template, feedname)
		if _, is_present := feednames[feedname]; is_present {
			return fmt.Errorf("%s: Duplicate name detected. Feed names must be globally unique.", feederr_template)
		}
		feednames[feedname] = true

		for sub_idx, twitterFilter := range twitterFeed.Filters {
			var filtererr_template = fmt.Sprintf("%s, filter index %d ", feederr_template, sub_idx+1)
			if err := twitterFilter.Validate(); err != nil {
				return fmt.Errorf("%s: %s", filtererr_template, err)
			}
			if twitterFilter.Percentile < 0.0 || twitterFilter.Percentile > 100.0 {
				return fmt.Errorf("%s: Percentile out of 0-100 range. : %f", filtererr_template, twitterFilter.Percentile)
			}
			if twitterFilter.MaxDailyPosts < 0 {
				return fmt.Errorf("%s: MaxDailyPosts must be a +ve integer. : %d", filtererr_template, twitterFilter.MaxDailyPosts)
			}
		}
	}
	return nil
}

func parseFromString(configblob string) (conf *Config, err error) {
	conf = &Config{
		Reddit: RedditConfig{
			Secrets: RedditSecrets{},
		},
		Twitter: TwitterConfig{
			Secrets: TwitterSecrets{},
		},
		BackendStorePath: filepath.Join(storageDir, databaseFileName),
	}

	if err = yaml.Unmarshal([]byte(configblob), conf); err != nil {
		return
	}

	// log.Fatal(spew.Sdump(conf))
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
