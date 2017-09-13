package redditbot

import (
	"github.com/coverprice/contentscraper/backingstore"
	"github.com/coverprice/contentscraper/scrapers/runner"
)

type DataStore struct {
	dbconn *backingstore.DbConn
}

func NewDataStore(dbconn *backingstore.DbConn) (datastore DataStore, err error) {
	datastore.dbconn = dbconn
	err = datastore.initTables()
	return
}

func (s *DataStore) initTables() (err error) {
	err = s.dbconn.ExecSql(`
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

func (s *DataStore) StorePost(rpost *runner.IPost) (err error) {
	post := (*rpost).(RedditPost)
	err = s.dbconn.ExecSql(`
        INSERT OR REPLACE INTO redditpost
            ( id
            , rawid
            , permalink
            , time_created
            , time_updated
            , is_active
            , is_sticky
            , score
            , title
            , url
            , subreddit_name
            , subreddit_id
        ) VALUES
            ( $a
            , $b
            , $c
            , $d
            , $e
            , $f
            , $g
            , $h
            , $i
            , $j
            , $k
            , $l
        )`,
		post.Id,
		post.RawId,
		post.Permalink,
		int64(post.TimeCreated),
		int64(post.TimeUpdated),
		post.IsActive,
		post.IsSticky,
		post.Score,
		post.Title,
		post.Url,
		post.SubredditName,
		post.SubredditId,
	)
	return err
}

/*
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
*/
