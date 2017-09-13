package redditbot

import (
	"fmt"
	"github.com/coverprice/contentscraper/scrapers/runner"
	"github.com/turnage/graw/reddit"
)

const (
	useragent = "Fedora:github.com/coverprice/contentscraper:0.1.0 (by /u/jayzefrashe)"
)

type Scraper struct {
	bot reddit.Bot
}

func NewScraper(clientid, clientsecret, username, password string) (scraper Scraper, err error) {
	cfg := reddit.BotConfig{
		Agent: useragent,
		App: reddit.App{
			ID:       clientid,
			Secret:   clientsecret,
			Username: username,
			Password: password,
		},
	}

	if scraper.bot, err = reddit.NewBot(cfg); err != nil {
		err = fmt.Errorf("Could not create reddit bot: %v", err)
		return
	}
	return
}

// Implements runner.IScraper
func (s *Scraper) Scrape(paramBag runner.ParamBag) ([]runner.IPost, error) {
	var subreddit_name, ok = paramBag["subreddit"]
	if !ok {
		panic("ParamBag must contain a 'subreddit' parameter")
	}

	var url_path = fmt.Sprintf("/r/%s", subreddit_name)

	var posts = make([]runner.IPost, 0)
	harvest, err := s.bot.Listing(url_path, "")
	if err != nil {
		return posts, fmt.Errorf("Failed to fetch listing for subreddit '%s': %v", subreddit_name, err)
	}

	for _, botpost := range harvest.Posts {
		reddit_post := redditPostFromBotPost(botpost)
		posts = append(posts, reddit_post)
	}
	return posts, nil
}
