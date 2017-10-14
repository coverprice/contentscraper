package types

import (
	"fmt"
	"github.com/coverprice/contentscraper/drivers"
)

type RedditPost struct {
	Id            string
	Name          string // Note that this more of an ID, used in the "after" parameter of the scraper
	TimeCreated   int64  `mapstructure:"time_created"`
	TimeStored    int64  `mapstructure:"time_stored"`
	Permalink     string
	IsActive      bool
	IsSticky      bool
	Score         int64
	Title         string
	Url           string
	SubredditName string `mapstructure:"subreddit_name"`
	SubredditId   string `mapstructure:"subreddit_id"`
}

// Compile time interface check
var _ drivers.ISourceConfig = &SubredditSourceConfig{}

type SubredditSourceConfig struct {
	Subreddit string
}

func (this *SubredditSourceConfig) GetSourceConfigId() drivers.SourceConfigId {
	return drivers.SourceConfigId(fmt.Sprintf("reddit:%s", this.Subreddit))
}

type SubredditFeed struct {
	Name          string
	Description   string
	Media         string
	SourceConfigs []SubredditSourceConfig
}
