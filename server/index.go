package server

import (
	"github.com/coverprice/contentscraper/drivers"
	"github.com/coverprice/contentscraper/server/htmlutil"
	// log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

var indexTemplateStr = `
    {{define "title"}}Home{{end}}
    {{define "content"}}
    <div class="list-group">
        {{range .DriverFeeds}}
            <div class="list-group-item">
                <a href="{{.BaseUrl}}/?feed={{.Feed.Name}}">{{.Feed.Name}} - {{.Feed.Description}}</a>
            </div>
        {{end}}
    </div>
    {{end}}
`

var indexTempl = htmlutil.ParseTemplate(indexTemplateStr)

// Verify that indexHandler implements http.Handler interface
var _ http.Handler = &indexHandler{}

type indexHandler struct {
	server *Server
}

func (this indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	type driverFeed struct {
		BaseUrl string
		Feed    drivers.Feed
	}

	var allfeeds []driverFeed
	for _, driver := range this.server.Drivers {
		var baseUrl = strings.TrimRight(driver.GetBaseUrlPath(), "/")
		for _, feed := range driver.GetFeeds() {
			allfeeds = append(allfeeds, driverFeed{
				BaseUrl: baseUrl,
				Feed:    feed,
			})
		}
	}

	data := struct {
		Title string
		htmlutil.Breadcrumbs
		DriverFeeds []driverFeed
	}{
		Title: "Content Scraper",
		Breadcrumbs: []htmlutil.Breadcrumb{
			htmlutil.NewBreadcrumb("Home", "/"),
		},
		DriverFeeds: allfeeds,
	}

	htmlutil.RenderTemplate(w, indexTempl, data)
}
