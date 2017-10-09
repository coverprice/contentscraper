package types

import (
	"fmt"
	"github.com/coverprice/contentscraper/drivers"
	"github.com/turnage/graw/reddit"
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

// Create a new RedditPost object from the scraper client's format
func NewRedditPostFromBotPost(bp *reddit.Post) (p RedditPost) {
	// Populate drivers.Post fields
	p.Id = bp.ID
	p.Name = bp.Name
	p.Score = int64(bp.Score)
	p.TimeCreated = int64(bp.CreatedUTC)
	p.TimeStored = int64(bp.CreatedUTC)
	p.Permalink = bp.Permalink
	p.IsActive = !bp.Deleted
	p.IsSticky = bp.Stickied
	p.Title = bp.Title
	p.Url = bp.URL
	p.SubredditName = bp.Subreddit
	p.SubredditId = bp.SubredditID
	return
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
