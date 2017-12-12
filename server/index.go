package server

import (
	"fmt"
	"github.com/coverprice/contentscraper/drivers"
	"github.com/coverprice/contentscraper/server/htmlutil"
	log "github.com/sirupsen/logrus"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"
)

var indexTemplateStr = `
    {{define "title"}}Home{{end}}
    {{define "js"}}
    <script>
    let menuIdx = 0;
    let numItems = 0;
    let menuUrl = '/';
    $(document).ready(function() {
	numItems = $('.mainmenu').length;
	moveToMenuIdx(0);
    });

    function moveToMenuIdx(idx) {
      $('.mainmenu.table-primary').removeClass('table-primary');
      let item = $('.mainmenu').eq(idx);
      item.addClass('table-primary');
      menuIdx = idx;
      menuUrl = item.find('a')[0].href;
    }

    $(document).keypress(function(event) {
        let key = String.fromCharCode(event.which);
        if (key == "k" && menuIdx > 0) {
	    moveToMenuIdx(menuIdx - 1);
	} else if (key == "j" && menuIdx < numItems-1) {
	    moveToMenuIdx(menuIdx + 1);
	} else if (key == "l") {
            window.location = menuUrl;
        } else {
            return;
        }
        event.preventDefault();
    });

    </script>
    {{end}}
    {{define "style"}}
    <style>
    .active a {
        color: white !important;
    }
    </style>
    {{end}}
    {{define "content"}}
    <div class="container">
    <table class="table">
    <tbody>
        {{range .DriverFeeds}}
            <tr class="mainmenu">
                <td><a href="{{.BaseUrl}}/?feed={{.Feed.Name}}">{{.Feed.Name}}</td>
                <td><a href="{{.BaseUrl}}/?feed={{.Feed.Name}}">{{.Feed.Description}}</td>
                <td><small class="text-muted">{{.StatusText}}</small></td>
            </tr>
        {{end}}
    </tbody>
    </table>
    </div>
    {{end}}
`

var indexTempl = htmlutil.ParseTemplate(indexTemplateStr)

// Verify that indexHandler implements http.Handler interface
var _ http.Handler = &indexHandler{}

type indexHandler struct {
	server *Server
}

type driverFeed struct {
	BaseUrl    string
	Feed       drivers.Feed
	StatusText string
}

func (this indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var allfeeds []driverFeed
	for _, driver := range this.server.Drivers {
		var baseUrl = strings.TrimRight(driver.GetBaseUrlPath(), "/")
		for _, feed := range driver.GetFeeds() {
			allfeeds = append(allfeeds, driverFeed{
				BaseUrl:    baseUrl,
				Feed:       feed,
				StatusText: getStatusText(feed),
			})
		}
	}
	sort.Sort(ByFeedName(allfeeds))

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

type ByFeedName []driverFeed

func (a ByFeedName) Len() int      { return len(a) }
func (a ByFeedName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByFeedName) Less(i, j int) bool {
	return a[i].Feed.Name < a[j].Feed.Name
}

func getStatusText(feed drivers.Feed) string {
	switch feed.Status {
	case drivers.FEEDSTATUS_IDLE:
		if feed.TimeLastHarvested == 0 {
			return "Not harvested yet"
		}
		duration := time.Now().Sub(time.Unix(feed.TimeLastHarvested, 0))
		return fmt.Sprintf("%02d:%02d ago", int(math.Floor(duration.Hours())), int(math.Floor(duration.Minutes()))%60)
	case drivers.FEEDSTATUS_ERROR:
		return "Error"
	case drivers.FEEDSTATUS_HARVESTING:
		return "Harvesting"
	default:
		log.Errorf("Unsupported status: %v", feed.Status)
		return "Error"
	}
}
