package persistence

import (
	"fmt"
	"github.com/coverprice/contentscraper/database"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	"github.com/mitchellh/mapstructure"
)

type Persistence struct {
	dbconn *database.DbConn
}

func NewPersistence(dbconn *database.DbConn) (persistence *Persistence, err error) {
	persistence = &Persistence{
		dbconn: dbconn,
	}
	err = persistence.initTables()
	return
}

// TODO: Investigate whether this code could be replaced by an ORM framework

func (this *Persistence) initTables() (err error) {
	err = this.dbconn.ExecSql(`
        CREATE TABLE IF NOT EXISTS redditpost
            ( id TEXT
            , name TEXT
            , rawid TEXT
            , permalink TEXT
            , time_created INTEGER
            , is_active INTEGER
            , is_sticky INTEGER
            , score INTEGER
            , title TEXT
            , url TEXT
            , subreddit_name TEXT
            , subreddit_id TEXT
            , is_published INTEGER DEFAULT 0
            , PRIMARY KEY (id, subreddit_id)
        )
        ;
        CREATE INDEX IF NOT EXISTS
            reddit_subreddit_name ON redditpost(subreddit_name)
        ;
        CREATE INDEX IF NOT EXISTS
            reddit_title ON redditpost(title)
        ;
        CREATE INDEX IF NOT EXISTS
            reddit_time_created ON redditpost(time_created)
    `)
	return
}

func (this *Persistence) StorePost(post *types.RedditPost) (err error) {
	err = this.dbconn.ExecSql(`
        INSERT OR REPLACE INTO redditpost
            ( id
            , name
            , permalink
            , time_created
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
            , $m
        )`,
		post.Id,
		post.Name,
		post.Permalink,
		int64(post.TimeCreated),
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

func (this *Persistence) GetPosts(
	where_clause string,
	params ...interface{},
) (posts []types.RedditPost, err error) {
	var rows database.MultiRowResult
	var sql = "SELECT * FROM redditpost " + where_clause
	if rows, err = this.dbconn.GetAllRows(sql, params...); err != nil {
		return nil, err
	}
	for _, row := range rows {
		var reddit_post types.RedditPost
		err = mapstructure.Decode(row, &reddit_post)
		if err != nil {
			return nil, fmt.Errorf("Could not decode reddit post ID: '%s' subreddit: '%s' %v", row["id"], row["subreddit"], err)
		}
		posts = append(posts, reddit_post)
	}
	return posts, nil
}
