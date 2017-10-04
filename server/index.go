package server

import (
	"github.com/coverprice/contentscraper/drivers"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
)

var indexTemplateStr = `
    <!DOCTYPE html>
    <html>
        <head>
            <meta charset="UTF-8">
            <title>{{.Title}}</title>
        </head>
        <body>
            <h1>{{.Title}}</h1>
            <ul>
                {{range .DriverFeeds}}
                    <li>
                        <a href="{{.BaseUrl}}/?feed={{.Feed.Name}}">{{.Feed.Name}} - {{.Feed.Description}}</a>
                    </li>
                {{end}}
            </ul>
        </body>
    </html>
`

var indexTempl = template.Must(template.New("index").Parse(indexTemplateStr))

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
		var baseUrl = driver.GetBaseUrlPath()
		for _, feed := range driver.GetFeeds() {
			allfeeds = append(allfeeds, driverFeed{
				BaseUrl: baseUrl,
				Feed:    feed,
			})
		}
	}

	data := struct {
		Title       string
		DriverFeeds []driverFeed
	}{
		Title:       "Content Scraper",
		DriverFeeds: allfeeds,
	}

	err := indexTempl.Execute(w, data)
	if err != nil {
		log.Fatal("Failed to execute index.html template:", err)
	}
}
