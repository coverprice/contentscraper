package server

import (
	"fmt"
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFilterByMaxDailyPost(t *testing.T) {
	id := 0
	fakePost := func(subreddit string, ageInDays, score int64) annotatedPost {
		id++
		return annotatedPost{
			RedditPost: types.RedditPost{
				Id:            fmt.Sprintf("id_%d", id),
				SubredditName: subreddit,
				Score:         score,
			},
			AgeInDays: ageInDays,
		}
	}
	posts := []annotatedPost{
		fakePost("zzz", 0, 10),
		fakePost("zzz", 0, 20),
		fakePost("zzz", 0, 30),

		fakePost("bbb", 0, 10),
		fakePost("bbb", 0, 20),
		fakePost("bbb", 0, 30),
		fakePost("bbb", 1, 20),
		fakePost("bbb", 2, 30),
		fakePost("bbb", 2, 10),
		fakePost("bbb", 2, 20),
		fakePost("bbb", 2, 30),

		fakePost("aaa", 0, 20),
		fakePost("aaa", 0, 10),
		fakePost("aaa", 0, 30),
	}

	feed := config.RedditFeed{
		Subreddits: []config.Subreddit{
			config.Subreddit{
				Name:          "aaa",
				MaxDailyPosts: 5,
			},
			config.Subreddit{
				Name:          "zzz",
				MaxDailyPosts: 2,
			},
			config.Subreddit{
				Name:          "bbb",
				MaxDailyPosts: 2,
			},
		},
	}

	results := filterByMaxDailyPosts(posts, &feed)

	type expected struct {
		SubredditName string
		AgeInDays     int64
		Score         int64
		ExpectedCnt   int
		SeenCnt       int
	}
	// we expect to see each of these once
	expectedResults := []expected{
		expected{"aaa", 0, 30, 1, 0},
		expected{"aaa", 0, 20, 1, 0},
		expected{"aaa", 0, 10, 1, 0},

		expected{"bbb", 0, 30, 1, 0},
		expected{"bbb", 0, 20, 1, 0},

		expected{"bbb", 1, 20, 1, 0},

		expected{"bbb", 2, 30, 2, 0}, // this we expected to see twice

		expected{"zzz", 0, 30, 1, 0},
		expected{"zzz", 0, 20, 1, 0},
	}

	// The +1 is because we expect to see bbb/2/30 twice.
	require.Equal(t, len(expectedResults)+1, len(results))

	for i, er := range expectedResults {
		found := false
		for _, r := range results {
			if r.SubredditName == er.SubredditName && r.AgeInDays == er.AgeInDays && r.Score == er.Score {
				expectedResults[i].SeenCnt++
				found = true
			}
		}
		require.True(t, found, "Expected to find: %v", er)
	}
	for _, er := range expectedResults {
		require.Equal(t, er.ExpectedCnt, er.SeenCnt, "Expected to see subreddit %s with age %d", er.SubredditName, er.AgeInDays)
	}
}
