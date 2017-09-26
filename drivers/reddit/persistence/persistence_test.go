package persistence

import (
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	"github.com/coverprice/contentscraper/toolbox"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCanCreateAndRetrieveRedditPost(t *testing.T) {
	var testDb = toolbox.NewTestDatabase(t)
	defer testDb.Cleanup()

	var p = NewPersistence(testDb.DbConn)

	var post = &types.RedditPost{
		Id:            "some_id",
		Name:          "jkjkjk",
		IsPublished:   false,
		TimeCreated:   1234,
		Permalink:     "/r/funny/xyz123",
		IsActive:      true,
		IsSticky:      false,
		Score:         5678,
		Title:         "A fake post",
		Url:           "/r/funny/xyz1235555",
		SubredditName: "funny",
		SubredditId:   "ppp9999",
	}

	var err error
	if err = p.StorePost(&post); err != nil {
		t.Error("Could not store 1st post", err)
	}

	post.Id = "another_id"
	post.TimeCreated = 9999
	if err := p.StorePost(&post); err != nil {
		t.Error("Could not store 2nd post", err)
	}

	var posts []types.RedditPost
	if posts, err = p.GetPosts(
		"WHERE subreddit=$a ORDER BY time_created", "funny",
	); err != nil {
		t.Error("Could not retrieve posts", err)
	}

	require.Equal(t, 2, len(posts), "Expected length of results")
	require.Equal(t, "some_id", posts[0].Id, "Incorrect 1st post ID")
	require.Equal(t, "another_id", posts[1].Id, "Incorrect 2nd post ID")
}
