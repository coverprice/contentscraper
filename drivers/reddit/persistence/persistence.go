package persistence

import (
	"database/sql"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	log "github.com/sirupsen/logrus"
)

type Persistence struct {
	dbconn          *sql.DB
	searchPostByPk  *sql.Stmt
	searchPostByUrl *sql.Stmt
}

func NewPersistence(dbconn *sql.DB) (persistence *Persistence, err error) {
	persistence = &Persistence{
		dbconn: dbconn,
	}
	err = persistence.initTables()

	persistence.searchPostByPk, err = persistence.dbconn.Prepare(`
        SELECT EXISTS(
            SELECT 1
            FROM redditpost
            WHERE id = $a
              AND subreddit_id = $b
            LIMIT 1
        )`)
	if err != nil {
		return
	}
	persistence.searchPostByUrl, err = persistence.dbconn.Prepare(`
        SELECT EXISTS(
            SELECT 1
            FROM redditpost
            WHERE url = $a
            LIMIT 1
        )`)
	if err != nil {
		return
	}
	return
}

// TODO: Investigate whether this code could be replaced by an ORM framework

func (this *Persistence) initTables() (err error) {
	_, err = this.dbconn.Exec(`
        CREATE TABLE IF NOT EXISTS redditpost
            ( id TEXT
            , name TEXT NOT NULL
            , permalink TEXT NOT NULL
            , time_created INTEGER NOT NULL
            , time_stored INTEGER NOT NULL
            , is_active INTEGER NOT NULL
            , is_sticky INTEGER NOT NULL
            , score INTEGER NOT NULL
            , title TEXT NOT NULL
            , url TEXT
            , subreddit_name TEXT NOT NULL
            , subreddit_id TEXT NOT NULL
            , is_published INTEGER DEFAULT 0
            , PRIMARY KEY (id, subreddit_id)
        ) WITHOUT ROWID
        ;
        CREATE INDEX IF NOT EXISTS
            reddit_subreddit_name ON redditpost(subreddit_name)
        ;
        CREATE INDEX IF NOT EXISTS
            reddit_url ON redditpost(url)
        ;
        CREATE INDEX IF NOT EXISTS
            reddit_time_created ON redditpost(time_created)
        ;
        CREATE INDEX IF NOT EXISTS
            reddit_time_stored ON redditpost(time_stored)
        ;
        CREATE INDEX IF NOT EXISTS
            reddit_id ON redditpost(id)
    `)
	return
}

// Stores/Updates a RedditPost and returns whether it was a store or an
// update.

type StoreResult int

const (
	STORERESULT_NEW = iota
	STORERESULT_UPDATED
	STORERESULT_SKIPPED
)

func (this *Persistence) StorePost(
	post *types.RedditPost,
) (
	result StoreResult,
	err error,
) {
	var postExists int
	log.Debugf("Checking if post exists")
	if err = this.searchPostByPk.QueryRow(post.Id, post.SubredditId).Scan(&postExists); err != nil {
		return
	}
	if postExists == 0 {
		log.Debugf("Post does not exist")
		// Does not exist when searching by Primary Key.

		// If this post has an image URL, verify that it doesn't already
		// exist elsewhere.
		if post.Url != "" {
			log.Debugf("Searching by URL")
			if err = this.searchPostByUrl.QueryRow(post.Url).Scan(&postExists); err != nil {
				return
			}
			if postExists != 0 {
				log.Debugf("Post URL already exists")
				return STORERESULT_SKIPPED, nil
			}
		}
		// We need to insert this post
		log.Debugf("Inserting post")
		if err = this.insertPost(post); err != nil {
			return
		}
		return STORERESULT_NEW, nil
	}
	// Exists, update
	log.Debugf("Updating post")
	if err = this.updatePost(post); err != nil {
		return
	}
	return STORERESULT_UPDATED, nil
}

func (this *Persistence) insertPost(post *types.RedditPost) (err error) {
	_, err = this.dbconn.Exec(`
        INSERT INTO redditpost
            ( id
            , name
            , permalink
            , time_created
            , time_stored
            , is_active
            , is_sticky
            , score
            , title
            , url
            , subreddit_name
            , subreddit_id
            , is_published
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
		int64(post.TimeStored),
		post.IsActive,
		post.IsSticky,
		post.Score,
		post.Title,
		post.Url,
		post.SubredditName,
		post.SubredditId,
		post.IsPublished,
	)
	return
}

func (this *Persistence) updatePost(post *types.RedditPost) (err error) {
	_, err = this.dbconn.Exec(`
        UPDATE redditpost SET
             name = $a
            , permalink = $b
            , is_active = $c
            , is_sticky = $d
            , score = $e
            , title = $f
            , url = $g
            , is_published = $h
        WHERE id = $i
          AND subreddit_id = $j
        `,
		post.Name,
		post.Permalink,
		post.IsActive,
		post.IsSticky,
		post.Score,
		post.Title,
		post.Url,
		post.IsPublished,

		post.Id,
		post.SubredditId,
	)
	return
}

func (this *Persistence) GetPosts(
	where_clause string,
	params ...interface{},
) (posts []types.RedditPost, err error) {
	var rows *sql.Rows
	var sql = `
        SELECT 
            id
            , name
            , permalink
            , time_created
            , time_stored
            , is_active
            , is_sticky
            , score
            , title
            , url
            , subreddit_name
            , subreddit_id
            , is_published
        FROM redditpost
        ` + where_clause
	if rows, err = this.dbconn.Query(sql, params...); err != nil {
		log.Debugf("Error calling SQL %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var redditPost types.RedditPost

		err = rows.Scan(
			&redditPost.Id,
			&redditPost.Name,
			&redditPost.Permalink,
			&redditPost.TimeCreated,
			&redditPost.TimeStored,
			&redditPost.IsActive,
			&redditPost.IsSticky,
			&redditPost.Score,
			&redditPost.Title,
			&redditPost.Url,
			&redditPost.SubredditName,
			&redditPost.SubredditId,
			&redditPost.IsPublished,
		)
		if err != nil {
			return
		}
		/*
			        Maybe this could make a comeback, if I could figure out how...
			        Possibly, Scan the row into a map?
			        if err = mapstructure.Decode(row, &redditPost); err != nil {
					return nil, fmt.Errorf("Could not decode reddit post ID: '%s' subreddit: '%s' %v", row["id"], row["subreddit"], err)
					}
		*/
		posts = append(posts, redditPost)
	}
	log.Debugf("Retrieved %d posts", len(posts))
	return posts, nil
}
