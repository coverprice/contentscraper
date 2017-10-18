package persistence

import (
	"fmt"
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
		"WHERE subreddit_name=$a ORDER BY time_created", "funny",
	); err != nil {
		t.Error("Could not retrieve posts:", err)
	}

	require.Equal(t, 2, len(posts), "Expected length of results")
	require.Equal(t, "some_id", posts[0].Id, "Incorrect 1st post ID")
	require.Equal(t, "another_id", posts[1].Id, "Incorrect 2nd post ID")
}

func createFakePosts(t *testing.T, sut *Persistence, subredditName string) {
	// Create some posts
	for i := 1; i <= 100; i++ {
		id := fmt.Sprintf("id_%d", i)
		post := types.RedditPost{
			Id:            id,
			Name:          id,
			TimeCreated:   1234,
			TimeStored:    1235,
			Permalink:     "/r/funny/xyz123",
			IsActive:      true,
			IsSticky:      false,
			Score:         int64(i * 2),
			Title:         "A fake post",
			Url:           "",
			SubredditName: subredditName,
			SubredditId:   subredditName + "ppp9999",
		}

		result, err := sut.StorePost(&post)
		require.Nil(t, err, "Could not store post")
		require.Equal(t, StoreResult(STORERESULT_NEW), result, "Unexpected StoreResult")
	}
}

func TestGetScoreAtPercentile(t *testing.T) {
	testDb, err := database.NewTestDatabase()
	if err != nil {
		t.Fatal("Could not init database", err)
	}
	defer testDb.Cleanup()

	sut, err := NewPersistence(testDb.DbConn)

	minTime := int64(0)
	subredditName := "funny"
	createFakePosts(t, sut, subredditName)

	type fixture struct {
		Percentile    float64
		ExpectedScore int
	}

	// 100 posts scored 2,4,6,...200, so the 70th percentile will be 60,
	// (Remember scores are ranked in descending order)
	fixtures := []fixture{
		fixture{70.0, 60},
		fixture{50.0, 100},
		fixture{30.0, 140},
		fixture{101.0, 0},
	}
	for i, fix := range fixtures {
		score, err := sut.GetScoreAtPercentile(minTime, subredditName, fix.Percentile)

		require.Nil(t, err,
			"Could not retrieve percentile score %f, index %d",
			fix.Percentile, i,
		)
		require.Equal(t, fix.ExpectedScore, score,
			"Unexpected percentile score, index %d",
			i,
		)
	}
}

func TestGetPostsForSubredditScores(t *testing.T) {
	testDb, err := database.NewTestDatabase()
	if err != nil {
		t.Fatal("Could not init database", err)
	}
	defer testDb.Cleanup()

	sut, err := NewPersistence(testDb.DbConn)

	createFakePosts(t, sut, "funny")
	createFakePosts(t, sut, "gifs")
	createFakePosts(t, sut, "pics")

	minTime := int64(0)

	posts, err := sut.GetPostsForSubredditScores(
		minTime,
		map[string]int{
			"funny": 198,
			"gifs":  200,
		})
	require.Nil(t, err, "Could not retrieve posts")
	require.Equal(t, 3, len(posts), "Unexpected number of posts")

	funnyCnt := 0
	gifsCnt := 0
	for _, post := range posts {
		switch post.SubredditName {
		case "funny":
			require.Equal(t, true, post.Score >= 198,
				"Expected score %d to be >= 198", post.Score)
			funnyCnt++
		case "gifs":
			require.Equal(t, int64(200), post.Score)
			gifsCnt++
		default:
			require.Fail(t, "Unexpected subreddit name", post.SubredditName)

		}
	}
	require.Equal(t, funnyCnt, 2)
	require.Equal(t, gifsCnt, 1)
}
