package persistence

import (
	"github.com/coverprice/contentscraper/database"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCanCreateAndRetrieveRedditPost(t *testing.T) {
	testDb, err := database.NewTestDatabase()
	if err != nil {
		t.Fatal("Could not init database", err)
	}
	defer testDb.Cleanup()

	sut, err := NewPersistence(testDb.DbConn)

	var post = &types.RedditPost{
		Id:            "some_id",
		Name:          "jkjkjk",
		TimeCreated:   1234,
		TimeStored:    1235,
		Permalink:     "/r/funny/xyz123",
		IsActive:      true,
		IsSticky:      false,
		Score:         5678,
		Title:         "A fake post",
		Url:           "https://imgur.com/foo.jpg",
		SubredditName: "funny",
		SubredditId:   "ppp9999",
	}

	var result StoreResult
	result, err = sut.StorePost(post)
	require.Nil(t, err, "Could not store 1st post")
	require.Equal(t, StoreResult(STORERESULT_NEW), result, "Unexpected StoreResult")

	// Same post again, should be updated
	post.Score = 9999
	result, err = sut.StorePost(post)
	require.Nil(t, err, "Could not update 1st post")
	require.Equal(t, StoreResult(STORERESULT_UPDATED), result, "Unexpected StoreResult")

	// Different post ID, should be skipped because the URL already exists
	post.Id = "another_id"
	post.SubredditId = "another_subreddit_id"
	result, err = sut.StorePost(post)
	require.Nil(t, err, "Could not duplicate 1st post")
	require.Equal(t, StoreResult(STORERESULT_SKIPPED), result, "Unexpected StoreResult")

	// Different post ID, should be new because the URL is different
	post.TimeCreated = 5555
	post.Url = "https://tinypic.com/bar.jpg"
	result, err = sut.StorePost(post)
	require.Nil(t, err, "Could not create 2nd post")
	require.Equal(t, StoreResult(STORERESULT_NEW), result, "Unexpected StoreResult")

	var posts []types.RedditPost
	if posts, err = sut.GetPosts(
		"WHERE lower(subreddit_name)=$a ORDER BY time_created", "funny",
	); err != nil {
		t.Error("Could not retrieve posts:", err)
	}

	require.Equal(t, 2, len(posts), "Expected length of results")
	require.Equal(t, "some_id", posts[0].Id, "Incorrect 1st post ID")
	require.Equal(t, "another_id", posts[1].Id, "Incorrect 2nd post ID")
}
