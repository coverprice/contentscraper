package server

import (
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	"github.com/coverprice/contentscraper/server/htmlutil"
	log "github.com/sirupsen/logrus"
	"sort"
	"time"
)

type annotatedPost struct {
	types.RedditPost
	AgeInDays int64 // how many days old this post is.
	MediaLink *htmlutil.MediaLink
}

type cachedPosts struct {
	Posts       []annotatedPost
	TimeCreated int64
}

var postCache = make(map[string]cachedPosts)

// getPosts retrieves all the posts for the given feed, and sorts them in
// display order.
// (The result may be large, and is expected to be cached).
func (this *HtmlViewerRequestHandler) getPosts(
	feed *config.RedditFeed,
) (posts []annotatedPost, err error) {
	now := int64(time.Now().Unix())

	cache, ok := postCache[feed.Name]
	if !ok || (cache.TimeCreated+6*60*60 < now) {
		posts, err = this.getPostsImpl(now, feed)
		if err != nil {
			return
		}
		cache = cachedPosts{
			Posts:       posts,
			TimeCreated: now,
		}
		postCache[feed.Name] = cache
	}
	return cache.Posts, nil
}

func (this *HtmlViewerRequestHandler) getPostsImpl(
	now int64,
	feed *config.RedditFeed,
) (posts []annotatedPost, err error) {

	minTime := int64(now - 7*24*60*60)
	if posts, err = this.getPostsFilteredByPercentile(minTime, feed); err != nil {
		return
	}

	// Decorate the posts with the age in days.
	decoratePostAge(now, posts)

	// Sort posts by feed, age_in_days, score DESC
	sortPostsByFeedAgeScore(posts)

	// Filter out posts that exceed the max_daily_posts criteria
	posts = filterByMaxDailyPosts(posts, feed)

	// Convert image links into embedded links
	decoratePostsWithMediaLinks(posts)

	// Sort into display order
	sortPostsIntoDisplayOrder(posts)
	return
}

func decoratePostAge(now int64, posts []annotatedPost) {
	// Determine when the next "3am" from now is.
	t := time.Unix(now, 0)
	day := t.Day()
	if t.Hour() >= 3 {
		day++
	}
	// The unix time considered to be the start of "0 days old".
	timeBoundary := time.Date(t.Year(), t.Month(), day, 3, 0, 0, 0, t.Location()).Unix()

	for i, _ := range posts {
		posts[i].AgeInDays = (timeBoundary - posts[i].TimeStored) / (24 * 60 * 60)
	}
}

func decoratePostsWithMediaLinks(posts []annotatedPost) {
	var err error
	for i, _ := range posts {
		if posts[i].Url != "" {
			if posts[i].MediaLink, err = htmlutil.UrlToEmbedUrl(posts[i].Url); err != nil {
				log.Error("Error trying to convert post URL to MediaLink", err)
			}
		}
	}
}

type ByFeedAgeScore []annotatedPost

func (a ByFeedAgeScore) Len() int      { return len(a) }
func (a ByFeedAgeScore) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByFeedAgeScore) Less(i, j int) bool {
	if a[i].SubredditName == a[j].SubredditName {
		if a[i].AgeInDays == a[j].AgeInDays {
			// Score DESC
			return a[i].Score >= a[j].Score
		} else {
			// AgeInDays ASC
			return a[i].AgeInDays < a[j].AgeInDays
		}
	} else {
		// SubredditName ASC
		return a[i].SubredditName < a[j].SubredditName
	}
}
func sortPostsByFeedAgeScore(posts []annotatedPost) {
	sort.Sort(ByFeedAgeScore(posts))
}

type ByAgeTimeStoredId []annotatedPost

func (a ByAgeTimeStoredId) Len() int      { return len(a) }
func (a ByAgeTimeStoredId) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByAgeTimeStoredId) Less(i, j int) bool {
	if a[i].AgeInDays == a[j].AgeInDays {
		if a[i].TimeStored == a[j].TimeStored {
			// Id ASC
			return a[i].Id < a[j].Id
		} else {
			// TimeStored DESC
			return a[i].TimeStored > a[j].TimeStored
		}
	} else {
		// AgeInDays ASC
		return a[i].AgeInDays < a[j].AgeInDays
	}
}
func sortPostsIntoDisplayOrder(posts []annotatedPost) {
	sort.Sort(ByAgeTimeStoredId(posts))
}

func filterByMaxDailyPosts(posts []annotatedPost, feed *config.RedditFeed) (results []annotatedPost) {
	var subredditToMaxDailyPosts = make(map[string]int) // Subreddit name -> Max daily posts
	for _, subreddit := range feed.Subreddits {
		subredditToMaxDailyPosts[subreddit.Name] = subreddit.MaxDailyPosts
	}

	// posts are expected to be sorted by Subreddit / AgeInDays / Score (DESC)
	dailyPostCnt := 0
	currentSubreddit := ""
	currentAgeInDays := int64(-1)
	for _, post := range posts {
		if currentSubreddit != post.SubredditName || currentAgeInDays != post.AgeInDays {
			currentSubreddit = post.SubredditName
			currentAgeInDays = post.AgeInDays
			dailyPostCnt = 0
		}
		dailyPostCnt++
		if dailyPostCnt < subredditToMaxDailyPosts[currentSubreddit] {
			results = append(results, post)
		}
	}
	return
}

// Retrieves posts for the feed, applying per-subreddit percentile filters.
// Returned posts are NOT sorted.
func (this *HtmlViewerRequestHandler) getPostsFilteredByPercentile(
	minTime int64,
	feed *config.RedditFeed,
) (posts []annotatedPost, err error) {
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

	var redditPosts []types.RedditPost
	if redditPosts, err = this.persistence.GetPostsForSubredditScores(minTime, subredditMinScore); err != nil {
		return
	}
	for _, redditPost := range redditPosts {
		posts = append(posts, annotatedPost{RedditPost: redditPost})
	}
	return
}
