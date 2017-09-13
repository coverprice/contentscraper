package redditbot

import (
	"github.com/coverprice/contentscraper/backingstore"
	"github.com/coverprice/contentscraper/config"
)

func makeDbConn(conf *config.Config) (dbconn *backingstore.DbConn, err error) {
	if dbconn, err = backingstore.GetConnection(conf.BackendStorePath); err != nil {
		return
	}

	err = dbconn.ExecSql(`
        CREATE TABLE IF NOT EXISTS redditpost
            ( id TEXT PRIMARY KEY
            , rawid TEXT
            , permalink TEXT
            , time_created INTEGER
            , time_updated INTEGER
            , is_active INTEGER
            , is_sticky INTEGER
            , score INTEGER
            , title TEXT
            , url TEXT
            , subreddit_name TEXT
            , subreddit_id TEXT
        )
        ;
        CREATE INDEX IF NOT EXISTS
            red_subreddit_name ON redditpost(subreddit_name)
        ;
        CREATE INDEX IF NOT EXISTS
            red_title ON redditpost(title)
        ;
        CREATE INDEX IF NOT EXISTS
            red_time_updated ON redditpost(time_updated)
    `)
	return
}
