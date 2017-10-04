package server

import (
	"github.com/coverprice/contentscraper/config"
	"net/http"
)

// The server is split into various layers. From the top:
// - HttpServer: (the standard net/http package) listens on the port and
//   receives the request, and delegates certain URL path prefixes to a HttpHandler.
// - HttpHandler: understands the HTTP protocol. Parses the URL path to get the
//   request parameters, and delegates handling the request to a IRequestHandler.
// - IRequestHandler: understands how to construct a response for a given
//   presentation protocol (e.g. HTML, XML (for RSS feeds, etc), and may delegate
//   the information retrieval to a lower layer.

type IRequestHandler interface {
	HandleFeed(
		feed *config.RedditFeed,
		pagenum int,
		w http.ResponseWriter,
	)
}
