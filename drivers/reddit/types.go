package reddit

import (
	"fmt"
	"github.com/turnage/graw/reddit"
	"time"
)

type RedditPost struct {
	Id            string
	RawId         string
	Permalink     string
	TimeCreated   uint64 `mapstructure:"time_created"`
	TimeUpdated   uint64 `mapstructure:"time_updated"`
	IsActive      bool
	IsSticky      bool
	Score         int64
	Title         string
	Url           string
	SubredditName string `mapstructure:"subreddit_name"`
	SubredditId   string `mapstructure:"subreddit_id"`
}

func redditPostFromBotPost(bp *reddit.Post) (p RedditPost) {
	p.Id = fmt.Sprintf("%s/%s", bp.ID, bp.Subreddit)
	p.RawId = bp.ID
	p.Permalink = bp.Permalink
	p.TimeCreated = bp.CreatedUTC
	p.TimeUpdated = uint64(time.Now().Unix())
	p.IsActive = !bp.Deleted
	p.IsSticky = bp.Stickied
	p.Score = int64(bp.Score)
	p.Title = bp.Title
	p.Url = bp.URL
	p.SubredditName = bp.Subreddit
	p.SubredditId = bp.SubredditID
	return
}
