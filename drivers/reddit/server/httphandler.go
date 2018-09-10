package server

import (
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strconv"
)

// Verify that HttpHandler implements http.Handler interface
var _ http.Handler = &HttpHandler{}

// HttpHandler is attached to the standard "http" server, bound to a base URL.
// It knows the base URL and understands HTTP. It parses the URL and delegates
// the response to the HtmlViewerRequestHandler.
type HttpHandler struct {
	requestHandler IRequestHandler
}

func NewHttpHandler(requestHandler IRequestHandler) *HttpHandler {
	handler := HttpHandler{
		requestHandler: requestHandler,
	}
	return &handler
}

func (this HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, "Invalid request. Cannot parse URL query", 500)
		return
	}
	feedname := values.Get("feed")
	feed, err := types.FeedRegistry.GetItemByName(feedname)
	if err != nil {
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
	this.requestHandler.HandleFeed(&feed.RedditFeed, pagenum, w)
}
