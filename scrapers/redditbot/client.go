package redditbot

import (
	"fmt"
	"github.com/coverprice/contentscraper/config"
	"github.com/turnage/graw/reddit"
)

const (
	useragent = "Fedora:github.com/coverprice/contentscraper:0.1.0 (by /u/jayzefrashe)"
)

func makeBot(conf *config.Config) (bot reddit.Bot, err error) {
	cfg := reddit.BotConfig{
		Agent: useragent,
		App: reddit.App{
			ID:       conf.Reddit_secrets.ClientId,
			Secret:   conf.Reddit_secrets.ClientSecret,
			Username: conf.Reddit_secrets.Username,
			Password: conf.Reddit_secrets.Password,
		},
	}

	if bot, err = reddit.NewBot(cfg); err != nil {
		err = fmt.Errorf("Could not create reddit bot: %v", err)
		return
	}
	return
}
