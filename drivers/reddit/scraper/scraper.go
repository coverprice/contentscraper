package reddit

import (
	"fmt"
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	"github.com/turnage/graw/reddit"
	"time"
)

const (
	useragent = "Fedora:github.com/coverprice/contentscraper:0.1.0 (by /u/jayzefrashe)"
)

type Scraper struct {
	bot        reddit.Bot
	params     scraperParams
	ResultChan chan *types.RedditPost // RedditPosts emitted to client here
	ErrChan    chan error             // Scraper errors emitted here
	KillChan   chan bool              // Client sends data here to terminate scrape.
}

// Parameters for a scrape request
type scraperParams struct {
	Subreddit string // Name of the subreddit to scrape
	UrlPath   string // listing URL path, e.g. "/r/somereddit/new"
	IsNew     bool   // True if /new. Used to determine when to stop scraping.
	FromTime  uint64 // Only return posts made after this Epoch time
	MaxPosts  int    // Only return the maximum of this many posts
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

// Scrapes the subreddit's "new" listing. This is a time-ordered (most-recent
// first) list of posts in the subreddit. We limit the number of responses
// according to the given datetime.
func NewParamsForNew(
	subreddit string,
	fromTime uint64, // Epoch seconds.
	maxPosts int,
) scraperParams {
	return scraperParams{
		Subreddit: subreddit,
		UrlPath:   fmt.Sprintf("/r/%s/new", subreddit),
		IsNew:     true,
		FromTime:  fromTime,
		MaxPosts:  maxPosts,
	}
}

// Scrapes the subreddit's "hot" listing. This is a list of popular posts
// (some new, some old). Reddit's algorithm for ranking these posts is
// opaque, some combo of date + score + number of views perhaps?
// Since the dates mean very little, we set the finishing criteria
// to be a max number of responses.
func NewParamsForHot(
	subreddit string,
	maxPosts int,
) scraperParams {
	return scraperParams{
		Subreddit: subreddit,
		UrlPath:   fmt.Sprintf("/r/%s", subreddit),
		IsNew:     false,
		FromTime:  uint64(time.Now().Unix()) - 7*24*60*60,
		MaxPosts:  maxPosts,
	}
}

// Starts the scraping process, where results are emitted on the
// scrapers ResultChan, errors to the ErrChan, and it can receive
// a Kill signal to halt a scraper.
func (this *Scraper) Start(params scraperParams) {
	this.ResultChan = make(chan *types.RedditPost, 20)
	this.ErrChan = make(chan error)
	this.KillChan = make(chan bool, 1)
	go this.doScrape(params)
}

// goroutine that does the scraping. Calls out to the bot to get the
// posts for a particular page and emits them to the Result channel.
// Continues to the "next page" until some criteria is met (e.g. age
// of posts, number of responses).
//
// It handles reddit's slightly odd
// pagination scheme, where each response in a "listing" will have an ID
// and a scrape request contains an "after" field that specifies the post ID
// that the responses should immediately follow (according to whatever
// ordering scheme is implicit in the scrape request. E.g. /new will be by
// date, and /hot (the default) is an opaque combination of date + score.)
func (this *Scraper) doScrape(params scraperParams) {
	defer close(this.ErrChan)
	defer close(this.ResultChan)

	var after = "" // Get postings after this subreddit post reference.
	var numResponsesSent int
	for {
		// Exit the loop if sent a signal from the calling process
		select {
		case <-this.KillChan:
			return
		default:
		}

		// Get listing (~100 posts from that subreddit)
		harvest, err := this.bot.Listing(params.UrlPath, after)
		if err != nil {
			this.ErrChan <- fmt.Errorf("Failed to fetch listing for subreddit '%s': %v", params.Subreddit, err)
			return
		}

		var lastPost *types.RedditPost = nil
		var numHarvestPosts = 0
		for _, botpost := range harvest.Posts {
			redditPost := types.NewRedditPostFromBotPost(botpost)
			if redditPost.IsSticky {
				// Skip Sticky posts because they tend to be non-useful posts like rules or announcements.
				continue
			}

			// We are presuming that posts are returned from the redditbot
			// in the response's order, so that the last post in the response
			// is the one we should use in the "after" field for the next
			// scrape request.
			lastPost = &redditPost
			if redditPost.TimeCreated < params.FromTime {
				// Filter out posts created before the FromTime.
				continue
			}

			numHarvestPosts++
			numResponsesSent++
			// Emit the post so that it can be persisted
			this.ResultChan <- &redditPost
		}

		// Do we need to scrape the next page?
		if numHarvestPosts == 0 || // Nothing in the response, so we quit.
			numResponsesSent >= params.MaxPosts ||
			(params.IsNew && lastPost.TimeCreated <= params.FromTime) {
			return
		}

		// In the next loop iteration, ask Reddit to give us posts after the given id.
		after = lastPost.Name
	}
}
