package main

import (
        "fmt"
        "log"
        "github.com/coverprice/contentscraper/config"
        "github.com/coverprice/contentscraper/redditbot"
        "flag"
)

func main() {
    flag.Parse()
    conf, err := config.Parse()
    if err != nil {
        log.Fatal("Could not load/parse config file", err)
    }

    bot, err := redditbot.MakeBot(conf)
    if err != nil {
        log.Fatal("Could not create Reddit bot.", err)
    }

    harvest, err := (*bot).Listing("/r/golang", "")
    if err != nil {
        log.Fatal("Failed to fetch /r/golang: ", err)
    }

    for _, post := range harvest.Posts[:5] {
        fmt.Printf("[%s] posted [%s]\n", post.Author, post.Title)
    }
}
