package server

import (
	"fmt"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTimeBoundaryIsNextFollowing3am(t *testing.T) {
	oneHour := int64(60 * 60)

	// date -d '2017-10-17 02:00:00 PDT' +%s   => 1508230800
	twoAm := int64(1508230800)
	require.Equal(t, getTimeBoundary(twoAm), twoAm+oneHour) // 3am the same day.

	// date -d '2017-10-17 21:00:00 PDT' +%s   => 1508299200
	ninePm := int64(1508299200)
	require.Equal(t, getTimeBoundary(ninePm), ninePm+6*oneHour) // 3am the following day.
}

func TestAgeInDaysIsCorrectlyDecorated(t *testing.T) {
	timeBoundary := int64(1508234400)
	oneDayInterval := int64(24 * 60 * 60)

	// Next timeBoundary should be 3am tomorrow morning, so this is the list of AgeInDays for each
	// of the above relative to that time.
	type fixture struct {
		TimeStored        int64
		ExpectedAgeInDays int64
	}
	fixtures := []fixture{
		fixture{timeBoundary, 0},
		fixture{timeBoundary - 1, 0},  // 1 second ago
		fixture{timeBoundary + 1, -1}, // 1 second in the future
		fixture{timeBoundary - oneDayInterval, 1},
		fixture{timeBoundary - (oneDayInterval * 2), 2},
		fixture{timeBoundary - (oneDayInterval * 3), 3},
		fixture{timeBoundary + 1 + (oneDayInterval), -2},
		fixture{timeBoundary + 1 + (oneDayInterval * 2), -3},
	}
	for i, fix := range fixtures {
		posts := []annotatedPost{
			annotatedPost{
				RedditPost: types.RedditPost{
					TimeStored: fix.TimeStored,
				},
			},
		}
		decoratePostAge(timeBoundary, posts)
		require.Equal(t, fix.ExpectedAgeInDays, posts[0].AgeInDays, fmt.Sprintf("9pm Post index %d error.", i))
	}
}
