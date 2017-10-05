package server

import (
	"fmt"
	"github.com/coverprice/contentscraper/config"
	persist "github.com/coverprice/contentscraper/drivers/reddit/persistence"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	"github.com/coverprice/contentscraper/server/htmlutil"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"strings"
)

// Verify that HtmlViewerRequestHandler implements IRequestHandler interface
var _ IRequestHandler = &HtmlViewerRequestHandler{}

type HtmlViewerRequestHandler struct {
	persistence *persist.Persistence
}

func NewHtmlViewerRequestHandler(persistence *persist.Persistence) *HtmlViewerRequestHandler {
	return &HtmlViewerRequestHandler{
		persistence: persistence,
	}
}

var htmlImageTemplateStr = `
    {{define "title"}}Reddit Feed - {{.Title}}{{end}}
    {{define "style"}}
    <style>
    html, body, .mysection {
        width:100%;
        height:100%;
        margin:0;
    }
    .myimg {
        width:auto;
        height:auto;
        max-width:100%;
        max-height:70%;
    }
    </style>
    {{end}}
    {{define "content"}}
    <h4>
        Reddit Feed: {{.Title}}
        <small class="text-muted">{{.Description}}</small>
    </h4>
    <div class="container-fluid">
        {{range .Posts}}
        <div class="row">
            <div class="col-2">
                <div class="list-group">
                    <div class="list-group-item">
                        {{.Title}}
                    </div>
                    <div class="list-group-item">
                        {{.SubredditName}}
                    </div>
                </div>
            </div>
            <div class="col">
                <div class="mysection">
                    <img src="{{.Url}}" class="myimg">
                </div>
            </div>
        </div>
        {{end}}
    </div>
    {{end}}
`
var htmlTextTemplateStr = `
    {{define "title"}}Reddit Feed - {{.Title}}{{end}}
    {{define "content"}}
    <h4>
        Reddit Feed: {{.Title}}
        <small class="text-muted">{{.Description}}</small>
    </h4>
    <div class="container-fluid">
        {{range .Posts}}
            <div class="row">
                <div class="col-2">
                    {{.Score}}
                </div>
                <div class="col">
                    {{.Title}}
                </div>
            </div>
        {{end}}
    </div>
    {{end}}
`
var htmlTextTempl = htmlutil.ParseTemplate(htmlTextTemplateStr)
var htmlImageTempl = htmlutil.ParseTemplate(htmlImageTemplateStr)

func (this *HtmlViewerRequestHandler) HandleFeed(
	feed *config.RedditFeed,
	pagenum int,
	w http.ResponseWriter,
) {
	posts, err := this.getPosts(feed, pagenum)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal error retrieving posts for feed: %s %v", feed.Name, err), 500)
		return
	}

	data := struct {
		Title       string
		Description string
		htmlutil.Breadcrumbs
		Posts []types.RedditPost
	}{
		Title:       feed.Name,
		Description: feed.Description,
		Breadcrumbs: []htmlutil.Breadcrumb{
			htmlutil.NewBreadcrumb("Home", "/"),
			htmlutil.NewBreadcrumb(feed.Name, "/"),
		},
		Posts: posts,
	}
	var t *template.Template
	switch feed.Media {
	case config.MEDIA_TYPE_TEXT:
		t = htmlTextTempl
		log.Debug("Using Text media template")
	case config.MEDIA_TYPE_IMAGE:
		t = htmlImageTempl
		log.Debug("Using Image media template")
	default:
		log.Fatal(fmt.Sprintf("Unsupported Media type: %s", feed.Media))
	}
	htmlutil.RenderTemplate(w, t, data)
}

func (this *HtmlViewerRequestHandler) getPosts(
	feed *config.RedditFeed,
	pagenum int,
) (posts []types.RedditPost, err error) {
	if pagenum == 0 {
		pagenum = 1
	}
	limit := 10
	offset := limit * (pagenum - 1)
	whereClause := `
        WHERE subreddit_name IN ('%s')
        ORDER BY time_stored DESC, id
        LIMIT $a
        OFFSET $b
    `
	var subredditNames []string
	for _, subreddit := range feed.Subreddits {
		subredditNames = append(subredditNames, subreddit.Name)
	}
	whereClause = fmt.Sprintf(whereClause, strings.Join(subredditNames, "', '"))
	log.Debug("Using whereclause: %s      LIMIT %d OFFSET %d", whereClause, limit, offset)
	return this.persistence.GetPosts(whereClause, limit, offset)
}
