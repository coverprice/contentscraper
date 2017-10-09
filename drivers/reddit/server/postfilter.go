package server

import (
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	// log "github.com/sirupsen/logrus"
	"time"
)

type cachedPosts struct {
	Posts       []types.RedditPost
	TimeCreated int64
}

var postCache = make(map[string]cachedPosts)

func (this *HtmlViewerRequestHandler) getPosts(
	feed *config.RedditFeed,
) (posts []types.RedditPost, err error) {
	now := int64(time.Now().Unix())

	cache, ok := postCache[feed.Name]
	if !ok || (cache.TimeCreated+6*60*60 < now) {
		minTime := int64(now - 7*24*60*60)
		posts, err = this.getPostsFilteredByPercentile(minTime, feed)
		if err != nil {
			return
		}
		cache = cachedPosts{
			Posts:       filterByMaxDailyPosts(now, feed, posts),
			TimeCreated: now,
		}
		postCache[feed.Name] = cache
	}
	return cache.Posts, nil
}

func filterByMaxDailyPosts(
	now int64,
	feed *config.RedditFeed,
	posts []types.RedditPost,
) []types.RedditPost {
	// The database can do the bulk of filtering posts by score for us, but
	// it's difficult to write a simple SQL filter that would limit the
	// number of posts per subreddit per day, so we do that in code.

	var maxDailyPosts = make(map[string]int)
	for _, subreddit := range feed.Subreddits {
		maxDailyPosts[subreddit.Name] = subreddit.MaxDailyPosts
	}

	// We iterate through each post, and calculate which feed/days-ago bucket
	// it would land in. If the bucket isn't full, we append it to results and
	// add it to the bucket count.

	// 2D map which maps subreddit -> days ago -> number of posts on that day
	var dailyPostCnt = make(map[string]map[int64]int)
	var results []types.RedditPost
	for _, post := range posts {
		daysAgo := int64(0)
		if now > post.TimeStored { // This should always be true, but it's just a sanity check.
			daysAgo = (now - post.TimeStored) / int64(24*60*60)
		}
		_, ok := dailyPostCnt[post.SubredditName]
		if !ok {
			dailyPostCnt[post.SubredditName] = make(map[int64]int)
		}
		dailyPostCnt[post.SubredditName][daysAgo]++
		if dailyPostCnt[post.SubredditName][daysAgo] <= maxDailyPosts[post.SubredditName] {
			results = append(results, post)
		}
	}
	return results
}

// Retrieves posts for the feed, applying per-subreddit percentile filters.
// Posts are returned in display order (reverse date).
func (this *HtmlViewerRequestHandler) getPostsFilteredByPercentile(
	minTime int64,
	feed *config.RedditFeed,
) (posts []types.RedditPost, err error) {
	var subredditMinScore = make(map[string]int)
	for _, subreddit := range feed.Subreddits {
		if subreddit.Percentile > 0.0 {
			subredditMinScore[subreddit.Name], err = this.persistence.GetScoreAtPercentile(
				minTime,
				subreddit.Name,
				subreddit.Percentile,
			)
			if err != nil {
				return
			}
		}
	}
	if len(subredditMinScore) == 0 {
		return
	}
	return this.persistence.GetPostsForSubredditScores(minTime, subredditMinScore)
}
