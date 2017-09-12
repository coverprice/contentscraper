package redditbot

import (
        "errors"
        "fmt"
        "github.com/coverprice/contentscraper/config"
        "github.com/turnage/graw/reddit"
)

const (
    useragent = "Fedora:github.com/coverprice/contentscraper:0.1.0 (by /u/jayzefrashe)"
)

func MakeBot(conf *config.Config) (bot *reddit.Bot, err error) {
    cfg := reddit.BotConfig{
        Agent: useragent,
        App: reddit.App{
            ID:     conf.Reddit_secrets.ClientId,
            Secret: conf.Reddit_secrets.ClientSecret,
            Username: conf.Reddit_secrets.Username,
            Password: conf.Reddit_secrets.Password,
        },
    }

    var new_bot reddit.Bot
    if new_bot, err = reddit.NewBot(cfg); err != nil {
        return nil, errors.New(fmt.Sprintf("Could not create bot: %v", err))
    }
    return &new_bot, nil
}
