package config

import (
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	MEDIA_TYPE_TEXT  = "text"
	MEDIA_TYPE_IMAGE = "image"
)

// Config is a struct that stores the configs of each type of data source.
// (Note: While Twitter's config is supported, the actual harvesting code has
// not been implemented yet)
type Config struct {
	Reddit           RedditConfig  `json:"reddit"`
	Twitter          TwitterConfig `json:"twitter"`
	BackendStorePath string        // Path to the database file
}

// RedditConfig is a struct that stores all Reddit-related configuration.
type RedditConfig struct {
	Secrets RedditSecrets `json:"secrets"`
	Feeds   []RedditFeed  `json:"feeds"`
}

// RedditSecrets stores the credentials used by the harvesting robot account.
type RedditSecrets struct {
	ClientId     string `json:"clientid"`
	ClientSecret string `json:"clientsecret"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

// RedditFeed describes the filtering configuration for feeds from Reddit sources, i.e. 1-many Subreddits.
// It's expected that subreddits primarily share the same media type (text or graphics). E.g. a "images"
// feed might include image-heavy subreddits like "funny" and "gifs", and a "text" feed might include
// text-heavy subreddits like "legaladvice" or "showerthoughts".
type RedditFeed struct {
	Name                 string      `json:"name"`
	Description          string      `json:"description"`
	Media                string      `json:"media"`
	Subreddits           []Subreddit `json:"subreddits"`
	DefaultPercentile    float64     `json:"percentile"`
	DefaultMaxDailyPosts int         `json:"max_daily_posts"`
}

// Validate returns nil if the RedditFeed structure is syntactically valid, or an error if it is not.
func (this RedditFeed) Validate() (err error) {
	if this.Name == "" {
		return fmt.Errorf("Empty feed name")
	}
	if !regexp.MustCompile("^[-_a-zA-Z0-9]+$").MatchString(this.Name) {
		return fmt.Errorf("Invalid feed name, must contain only chars from A-Z, a-z, 0-9, '_', & '-'")
	}
	if this.Description == "" {
		return fmt.Errorf("Empty feed description")
	}
	if !(this.Media == MEDIA_TYPE_IMAGE || this.Media == MEDIA_TYPE_TEXT) {
		return fmt.Errorf(
			"Invalid media type: '%s', must be one of '%s' or '%s'",
			this.Media,
			MEDIA_TYPE_IMAGE,
			MEDIA_TYPE_TEXT,
		)
	}
	return nil
}

// Subreddit describes the filtering configuration for specific subreddit, e.g. /r/funny.
// It is an element of the RedditFeed structure.
type Subreddit struct {
	Name          string  `json:"name"`            // The subreddit name, without the leading '/r/'
	Percentile    float64 `json:"percentile"`      // Percent of posts to include from this subreddit (0-100)
	MaxDailyPosts int     `json:"max_daily_posts"` // Maximum # of posts to include per day from this subreddit.
}

// Validate returns nil if the Subreddit structure is syntactically valid, or an error if it is not.
func (this Subreddit) Validate() (err error) {
	if this.Name == "" {
		return fmt.Errorf("Empty subreddit name")
	}
	if this.Percentile < 0.0 || this.Percentile > 100.0 {
		return fmt.Errorf("Percentile out of 0-100 range. : %f", this.Percentile)
	}
	if this.MaxDailyPosts < 0 {
		return fmt.Errorf("MaxDailyPosts must be a +ve integer. : %d", this.MaxDailyPosts)
	}
	return nil
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

func (this TwitterFeed) Validate() (err error) {
	if this.Name == "" {
		return fmt.Errorf("Empty feed name")
	}
	if this.Description == "" {
		return fmt.Errorf("Empty feed description")
	}
	return nil
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
	if this.Percentile < 0.0 || this.Percentile > 100.0 {
		return fmt.Errorf("Percentile out of 0-100 range: %f", this.Percentile)
	}
	if this.MaxDailyPosts < 0 {
		return fmt.Errorf("MaxDailyPosts must be a +ve integer: %d", this.MaxDailyPosts)
	}
	return nil
}

// Validate returns nil if the Config structure is syntactically and semantically valid, otherwise it returns an error.
func (this *Config) Validate() (err error) {
	// validation is mainly concerned with ensuring that all feed names are unique
	// across source types, that percentile values are within range, etc.

	var feednames = make(map[string]bool)
	var subredditnames = make(map[string]bool)

	log.Debug("Validating config file")
	for idx, redditFeed := range this.Reddit.Feeds {
		var feedname = redditFeed.Name
		var feederr_template = fmt.Sprintf("Problem in Reddit feed '%s', index %d ", feedname, idx+1)
		if err := redditFeed.Validate(); err != nil {
			return fmt.Errorf("%s: %s", feederr_template, err)
		}
		if _, is_present := feednames[feedname]; is_present {
			return fmt.Errorf("%s: Duplicate name detected. Feed names must be globally unique.", feederr_template)
		}
		feednames[feedname] = true

		for sub_idx, subreddit := range redditFeed.Subreddits {
			var subredditerr_template = fmt.Sprintf("%s, subreddit: '%s' (index %d) ", feederr_template, subreddit.Name, sub_idx+1)
			if err := subreddit.Validate(); err != nil {
				return fmt.Errorf("%s: %s", subredditerr_template, err)
			}
			if _, is_present := subredditnames[subreddit.Name]; is_present {
				return fmt.Errorf("%s: Duplicate subreddit name. Subreddits must be unique across feeds.", subredditerr_template)
			}
		}
	}

	for idx, twitterFeed := range this.Twitter.Feeds {
		var feederr_template = fmt.Sprintf("Problem in Twitter feed #%d, name: '%s' ", idx+1, twitterFeed.Name)
		if err := twitterFeed.Validate(); err != nil {
			return fmt.Errorf("%s: %s", feederr_template, err)
		}
		if _, is_present := feednames[twitterFeed.Name]; is_present {
			return fmt.Errorf("%s: Duplicate name detected. Feed names must be globally unique.", feederr_template)
		}
		feednames[twitterFeed.Name] = true

		for sub_idx, twitterFilter := range twitterFeed.Filters {
			var filtererr_template = fmt.Sprintf("%s, filter index %d ", feederr_template, sub_idx+1)
			if err := twitterFilter.Validate(); err != nil {
				return fmt.Errorf("%s: %s", filtererr_template, err)
			}
		}
	}
	return nil
}

func (this *Config) populateDefaults() {
	// Populate defaults
	for idx, redditfeed := range this.Reddit.Feeds {
		if redditfeed.DefaultPercentile == 0 {
			this.Reddit.Feeds[idx].DefaultPercentile = float64(defaultPercentile)
		}
		// Due to the way the Unmarshaller works, we can't differentiate between the config file specifying
		// [Default]MaxDailyPosts = 0 vs not specifying anything at all. So the convention is that if the user
		// wants [Default]MaxDailyPosts to be 0, then they should specify a negative number.
		if redditfeed.DefaultMaxDailyPosts < 0 {
			this.Reddit.Feeds[idx].DefaultMaxDailyPosts = 0
		} else if redditfeed.DefaultMaxDailyPosts == 0 {
			this.Reddit.Feeds[idx].DefaultMaxDailyPosts = defaultMaxDailyPosts
		}

		if redditfeed.Media == "" {
			this.Reddit.Feeds[idx].Media = MEDIA_TYPE_TEXT
		}
		for subidx, subreddit := range redditfeed.Subreddits {
			// Canonicalize subreddit name (i.e. lowercase)
			this.Reddit.Feeds[idx].Subreddits[subidx].Name = strings.ToLower(subreddit.Name)
			if subreddit.Percentile == 0 {
				this.Reddit.Feeds[idx].Subreddits[subidx].Percentile = this.Reddit.Feeds[idx].DefaultPercentile
			}
			// See above.
			if subreddit.MaxDailyPosts < 0 {
				this.Reddit.Feeds[idx].Subreddits[subidx].MaxDailyPosts = 0
			} else if subreddit.MaxDailyPosts == 0 {
				this.Reddit.Feeds[idx].Subreddits[subidx].MaxDailyPosts = this.Reddit.Feeds[idx].DefaultMaxDailyPosts
			}
		}
	}
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

	// This will read the YAML and convert it into the structure defined above.
	// If the file has a syntax mismatch with the object structure we've defined,
	// the unmarshaller will return an error.
	err = yaml.Unmarshal([]byte(configblob), conf)
	return
}

// GetConfig finds and loads the config file into a Config structure, populates
// default values where they have not been specified, and validates the result.
// It returns a new Config structure.
func GetConfig() (conf *Config, err error) {
	configFilePath, err := locateConfigFile()
	if err != nil {
		return nil, fmt.Errorf("Failed to locate config file: %v", err)
	}

	var raw_contents []byte
	if raw_contents, err = ioutil.ReadFile(configFilePath); err != nil {
		return nil, fmt.Errorf("Failed to read config file: %v", err)
	}
	var contents = string(raw_contents)

	if conf, err = parseFromString(contents); err != nil {
		return nil, fmt.Errorf("Failed to parse config file: %v", err)
	}

	conf.populateDefaults()

	err = conf.Validate()
	// log.Fatal(spew.Sdump(conf))
	return
}
