package server

import (
	"fmt"
	"github.com/coverprice/contentscraper/config"
	persist "github.com/coverprice/contentscraper/drivers/reddit/persistence"
	"github.com/coverprice/contentscraper/server/htmlutil"
	// log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	NUM_ITEMS_PER_PAGE = 10
)

// Verify that HtmlViewerRequestHandler implements IRequestHandler interface
var _ IRequestHandler = &HtmlViewerRequestHandler{}

// The HtmlViewerRequestHandler handles a single request for a page within a specific
// feed. It requests the posts from a separate class, and converts them to an HTML response.
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
    {{define "js"}}
    <script src="/static/imagesloaded.pkgd.min.js"></script>
    <script>
    let globals = {
        numPages: {{.NumPages}},
        currentPageNum: {{.PageNum}},
        previousPageLink: '{{.PreviousPagelink.Link}}',
        nextPageLink: '{{.NextPagelink.Link}}',
    };
    </script>
    <script src="/static/viewer.js"></script>
    {{end}}
    {{define "pagination"}}
    <nav>
        <ul class="pagination">
        {{range .Pagelinks}}
            <li class="page-item{{if not .IsEnabled}} disabled{{end}} {{if .IsHighlighted}} active{{end}}">
                <a class="page-link" href="{{.Link}}" {{if not .IsEnabled}} tabindex="-1"{{end}}>{{.Text}}</a>
            </li>
        {{end}}
        </ul>
    </nav>
    {{end}}
    {{define "content"}}
    <h4>
        Reddit Feed: {{.Title}}
        <small class="text-muted">{{.Description}}</small>
    </h4>

    {{template "pagination" .}}

    <div class="container-fluid">
        {{range $itemIndex, $post := .Posts}}
        <div class="row feeditem">
            <div class="col">
                <div class="container-fluid">
                    <div class="row">
                        <div class="col alert alert-info">
                            <a href="https://www.reddit.com{{.Permalink}}">{{.Title}}</a>
                            <small>Score: {{.Score}}</small>
                            <small>Days old: {{.AgeInDays}}</small>
                            <small class="text-muted">{{.SubredditName}}</small>
                        </div>
                    </div>
                    {{if .MediaLink}}
                    <div class="row">
                        <div class="col">
                            <a href="{{.Url}}">
                                {{if not (eq .MediaLink.Embed "")}}
                                    {{.MediaLink.Embed}}
                                {{else if hasSuffix .MediaLink.Url ".mp4"}}
                                    <video playsinline autoplay loop controls class="videocontainer">
                                        <source src="{{.MediaLink.Url}}" type="video/mp4" />
                                    </video>
                                {{else if hasSuffix .MediaLink.Url ".webm"}}
                                    <video playsinline autoplay loop controls class="videocontainer">
                                        <source src="{{.MediaLink.Url}}" type="video/webm" />
                                    </video>
                                {{else if not (eq .MediaLink.Url "")}}
                                    <img src="{{.MediaLink.Url}}">
                                {{else}}
                                    {{.MediaLink.Url}}
                                    <small>[No preview available]</small>
                                {{end}}
                            </a>
                        </div>
                    </div>
                    {{end}}
                </div>
            </div>
        </div>
        {{end}}
    </div>

    {{template "pagination" .}}

    {{end}}
`
var htmlImageTempl = htmlutil.ParseTemplate(htmlImageTemplateStr)

type pagelink struct {
	Text          string
	Link          string
	IsEnabled     bool
	IsHighlighted bool
}

func (this *HtmlViewerRequestHandler) HandleFeed(
	feed *config.RedditFeed,
	pageNum int,
	w http.ResponseWriter,
) {
	if pageNum == 0 {
		pageNum = 1
	}

	posts, err := this.getPosts(feed)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal error retrieving posts for feed: %s %v", feed.Name, err), 500)
		return
	}

	itemsPerPage := NUM_ITEMS_PER_PAGE
	numPages := (len(posts) + NUM_ITEMS_PER_PAGE - 1) / NUM_ITEMS_PER_PAGE
	startIdx := itemsPerPage * (pageNum - 1)

	if startIdx >= len(posts) || startIdx+itemsPerPage >= len(posts) {
		// Out of bounds.
		posts = []annotatedPost{}
	} else {
		posts = posts[startIdx : startIdx+itemsPerPage]
	}

	pagelinks := getPagelinks(feed.Name, pageNum, numPages)
	data := struct {
		Title       string
		Description string
		htmlutil.Breadcrumbs
		Posts            []annotatedPost
		Pagelinks        []pagelink
		PreviousPagelink pagelink
		NextPagelink     pagelink
		NumPages         int
		PageNum          int
	}{
		Title:       feed.Name,
		Description: feed.Description,
		Breadcrumbs: []htmlutil.Breadcrumb{
			htmlutil.NewBreadcrumb("Home", "/"),
			htmlutil.NewBreadcrumb(feed.Name, "/"),
		},
		Posts:            posts,
		Pagelinks:        pagelinks,
		PreviousPagelink: pagelinks[0], // Clunky, but necessary since arithmetic isn't possible in templates.
		NextPagelink:     pagelinks[len(pagelinks)-1],
		NumPages:         numPages,
		PageNum:          pageNum,
	}
	htmlutil.RenderTemplate(w, htmlImageTempl, data)
}

func getPagelinks(feedname string, pageNum, numPages int) (links []pagelink) {
	link := pagelink{
		Text:          "Previous",
		Link:          constructUrl(&feedname, pageNum-1),
		IsEnabled:     true,
		IsHighlighted: false,
	}
	if pageNum == 1 {
		link.IsEnabled = false
	}
	links = append(links, link)

	for pn := 1; pn <= numPages; pn++ {
		link = pagelink{
			Text:          fmt.Sprintf("%d", pn),
			Link:          constructUrl(&feedname, pn),
			IsEnabled:     true,
			IsHighlighted: (pageNum == pn),
		}
		links = append(links, link)
	}
	link = pagelink{
		Text:          "Next",
		Link:          constructUrl(&feedname, pageNum+1),
		IsEnabled:     true,
		IsHighlighted: false,
	}
	links = append(links, link)
	return
}
