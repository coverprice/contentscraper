package drivers

import (
	"net/http"
)

// Feed is a data object returned by a Driver to describe a stream of posts,
// that (depending on the Driver) may be from multiple sources. E.g. a "amusing"
// Feed might include posts from /r/funny and /r/gifs.
// It's used by the UI to allow the user to select a feed and explore the posts within.
type Feed struct {
	Name              string
	Description       string
	Status            FeedHarvestStatus
	// The epoch time (in seconds) that this Feed was last harvested for content.
	// 0 means it was never harvested.
	TimeLastHarvested int64
}

type FeedHarvestStatus int

const (
	FEEDHARVESTSTATUS_IDLE       FeedHarvestStatus = 0
	FEEDHARVESTSTATUS_HARVESTING FeedHarvestStatus = 1
	FEEDHARVESTSTATUS_ERROR      FeedHarvestStatus = 2
)

// --------------------------------------

// The interface that all content source drivers must implement.
type IDriver interface {
	// Scrapes and persists posts from the website. (Run periodically by the mainloop)
	Harvest() error

	// Return the path that the driver's publishing handler will handle, e.g. "/reddit/"
	GetBaseUrlPath() string
	GetFeeds() []Feed
	GetHttpHandler() http.Handler
}

// --------------------------------------

// A string that uniquely identifies a SourceConfig. Used to store
// its "LastRun" data. E.g. "reddit/r/some_subreddit" or "twitter"
type SourceConfigId string

// ISourceConfig describes a data source to the backend, which may have extra details
// about a sub-stream, e.g. a specific sub-reddit.
type ISourceConfig interface {
	GetSourceConfigId() SourceConfigId
}
