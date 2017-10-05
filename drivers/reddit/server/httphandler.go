package server

import (
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/drivers"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strconv"
)

// Verify that HttpHandler implements http.Handler interface
var _ http.Handler = &HttpHandler{}

// HttpHandler is attached to the standard "http" server, bound to a base URL.
// It knows the base URL and understands Http. It parses the URL and delegates
// the response to the demuxer.
type HttpHandler struct {
	feeds          map[string]config.RedditFeed
	requestHandler IRequestHandler
}

func NewHttpHandler(requestHandler IRequestHandler) *HttpHandler {
	handler := HttpHandler{
		feeds:          make(map[string]config.RedditFeed),
		requestHandler: requestHandler,
	}
	return &handler
}

func (this *HttpHandler) AddFeed(feed config.RedditFeed) {
	this.feeds[feed.Name] = feed
}

func (this *HttpHandler) GetFeedsForServer() []drivers.Feed {
	var ret []drivers.Feed
	for _, redditfeed := range this.feeds {
		ret = append(ret, drivers.Feed{
			Name:        redditfeed.Name,
			Description: redditfeed.Description,
		})
	}
	return ret
}

func (this HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "Invalid request. Cannot parse URL query", 500)
		return
	}
	feedname := values.Get("feed")
	feed, ok := this.feeds[feedname]
	if !ok {
		log.Errorf("Unknown reddit feed name: '%s'", feedname)
		http.NotFound(w, r)
		return
	}

	var pagenum int
	pagenumStr := values.Get("page")
	if pagenumStr != "" {
		pagenum, err = strconv.Atoi(pagenumStr)
	}
	if pagenum < 0 || pagenum > 5000 {
		http.Error(w, "Invalid request. Cannot parse page number", 500)
		return
	}
	log.Debugf("Received request for feed '%s', page %d", feedname, pagenum)
	this.requestHandler.HandleFeed(&feed, pagenum, w)
}
