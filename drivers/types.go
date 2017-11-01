package drivers

import (
	"net/http"
)

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

// --------------------------------------

// Feed is a data object returned by a Driver to describe a feed.
// It's used by the HTTP server to construct links to a specific feed.
type Feed struct {
	Name              string
	Description       string
	Status            FeedStatus
	TimeLastHarvested int64
}

type FeedStatus int

const (
	FEEDSTATUS_IDLE FeedStatus = 0
	FEEDSTATUS_HARVESTING
	FEEDSTATUS_ERROR
)
