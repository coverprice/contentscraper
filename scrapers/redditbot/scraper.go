package redditbot

import (
	"fmt"
	"github.com/coverprice/contentscraper/backingstore"
	"github.com/coverprice/contentscraper/config"
	"github.com/mitchellh/mapstructure"
	"github.com/turnage/graw/reddit"
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
	p.TimeUpdated = bp.CreatedUTC
	p.IsActive = !bp.Deleted
	p.IsSticky = bp.Stickied
	p.Score = int64(bp.Score)
	p.Title = bp.Title
	p.Url = bp.URL
	p.SubredditName = bp.Subreddit
	p.SubredditId = bp.SubredditID
	return
}

func MakeScraper(conf *config.Config) (scraper Scraper, err error) {
	if scraper.Bot, err = makeBot(conf); err != nil {
		return
	}
	if scraper.DbConn, err = makeDbConn(conf); err != nil {
		return
	}
	return
}

type Scraper struct {
	reddit.Bot
	*backingstore.DbConn
}

func (s *Scraper) QuerySql(sql string, params ...interface{}) (posts []RedditPost, err error) {
	var rows backingstore.MultiRowResult
	if rows, err = s.DbConn.GetAllRows(sql, params...); err != nil {
		return nil, err
	}
	for _, row := range rows {
		var reddit_post RedditPost
		err = mapstructure.Decode(row, &reddit_post)
		if err != nil {
			panic(err)
		}
		posts = append(posts, reddit_post)
	}
	return
}

func (s *Scraper) GetPosts(subreddit_name string) (posts []RedditPost, err error) {
	subreddit_name = fmt.Sprintf("/r/%s", subreddit_name)

	harvest, err := s.Bot.Listing(subreddit_name, "")
	if err != nil {
		return posts, fmt.Errorf("Failed to fetch listing for subreddit '%s': %v", subreddit_name, err)
	}

	for _, botpost := range harvest.Posts {
		reddit_post := redditPostFromBotPost(botpost)
		posts = append(posts, reddit_post)
	}
	return
}
