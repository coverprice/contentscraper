package reddit

import (
	"fmt"
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	// log "github.com/sirupsen/logrus"
	"github.com/turnage/graw/reddit"
	"strings"
)

const (
	useragent = "Fedora:github.com/coverprice/contentscraper:0.1.0 (by /u/jayzefrashe)"
)

type Scraper struct {
	bot reddit.Bot
}

// Parameters for a scrape request
type Context struct {
	Subreddit         string // Name of the subreddit to scrape
	UrlPath           string // listing URL path, e.g. "/r/somereddit/new"
	After             string // used for pagination
	NumPostsPerScrape int
}

func NewDefaultContext(subreddit string) Context {
	return Context{
		Subreddit:         subreddit,
		UrlPath:           fmt.Sprintf("/r/%s", subreddit),
		After:             "",
		NumPostsPerScrape: 100,
	}
}

// Scrapes the subreddit's "new" listing. This is a time-ordered (most-recent
// first) list of posts in the subreddit. We limit the number of responses
// according to the given datetime.
func NewContextForNew(subreddit string) Context {
	context := NewDefaultContext(subreddit)
	context.UrlPath = fmt.Sprintf("/r/%s/new", subreddit)
	return context
}

// Scrapes the subreddit's "hot" listing. This is a list of popular posts
// (some new, some old). Reddit's algorithm for ranking these posts is
// opaque, some combo of date + score + number of views perhaps?
// Since the dates mean very little, we set the finishing criteria
// to be a max number of responses.
func NewContextForHot(subreddit string) Context {
	return NewDefaultContext(subreddit)
}

func NewScraper(clientid, clientsecret, username, password string) (scraper *Scraper, err error) {
	cfg := reddit.BotConfig{
		Agent: useragent,
		App: reddit.App{
			ID:       clientid,
			Secret:   clientsecret,
			Username: username,
			Password: password,
		},
	}

	scraper = &Scraper{}
	if scraper.bot, err = reddit.NewBot(cfg); err != nil {
		err = fmt.Errorf("Could not create reddit bot: %v", err)
		return nil, err
	}
	return
}

func NewScraperFromConfig(conf *config.Config) (scraper *Scraper, err error) {
	scraper, err = NewScraper(
		conf.Reddit.Secrets.ClientId,
		conf.Reddit.Secrets.ClientSecret,
		conf.Reddit.Secrets.Username,
		conf.Reddit.Secrets.Password,
	)
	return
}

// GetNextResults scrapes the current "page" of results and returns
// a set of RedditPosts. It handles reddit's slightly odd
// pagination scheme, where each response in a "listing" will have an ID
// and a scrape request contains an "after" field that specifies the post ID
// that the responses should immediately follow (according to whatever
// ordering scheme is implicit in the scrape request. E.g. /new will be by
// date, and /hot (the default) is an opaque combination of date + score.)
func (this *Scraper) GetNextResults(context *Context) (posts []types.RedditPost, err error) {
	// Get listing (~100 posts from that subreddit)
	harvest, err := this.bot.ListingWithParams(
		context.UrlPath,
		map[string]string{
			"limit": fmt.Sprintf("%d", context.NumPostsPerScrape),
			"after": context.After,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch listing for subreddit '%s': %v", context.Subreddit, err)
	}

	for _, botpost := range harvest.Posts {
		redditPost := newRedditPostFromBotPost(botpost)
		if redditPost.IsSticky {
			// Skip Sticky posts because they tend to be non-useful posts like rules or announcements.
			continue
		}
		posts = append(posts, redditPost)
		context.After = redditPost.Name
	}
	return posts, nil
}

// Create a new RedditPost object from the scraper client's format
func newRedditPostFromBotPost(bp *reddit.Post) (p types.RedditPost) {
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
	p.SubredditName = strings.ToLower(bp.Subreddit)
	p.SubredditId = bp.SubredditID
	return
}
