package server

import (
	"fmt"
	"github.com/coverprice/contentscraper/config"
	persist "github.com/coverprice/contentscraper/drivers/reddit/persistence"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	"github.com/coverprice/contentscraper/server/htmlutil"
	"github.com/coverprice/contentscraper/toolbox"
	// log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
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
    {{define "js"}}
	<script>
    // Returns the scale factor required to get the image dimensions to fit into the given window
    // dimensions.
	function getScaleFactor(img_w, img_h, window_w, window_h) {
	    let scale_factor = window_w / img_w;
        return (scale_factor * img_h > window_h) ? window_h / img_h : scale_factor;
    }
  		// Scale down images that are wider than the page so they fit on the page.
	$(document).ready(function() {
	    $('.postimage').each(function(idx, el) {
            let el_w = el.naturalWidth || el.videoWidth;
            let el_h = el.naturalHeight || el.videoHeight;
            let max_w = window.innerWidth - el.x - 50;
            let max_h = window.innerHeight - 100;
            let scale_factor = getScaleFactor(el_w, el_h, max_w, max_h)
            el.style.width = Math.floor(scale_factor * el_w) + "px";
            el.style.height = Math.floor(scale_factor * el_h) + "px";
		})
	});
	</script>
    {{end}}
    {{define "content"}}
    <h4>
        Reddit Feed: {{.Title}}
        <small class="text-muted">{{.Description}}</small>
    </h4>

    <nav>
        <ul class="pagination">
        {{range .Pagelinks}}
            <li class="page-item{{if not .IsEnabled}} disabled{{end}} {{if .IsHighlighted}} active{{end}}">
                <a class="page-link" href="{{.Link}}" {{if not .IsEnabled}} tabindex="-1"{{end}}>{{.Text}}</a>
            </li>
        {{end}}
        </ul>
    </nav>

    <div class="container-fluid">
        {{range .Posts}}
        <div class="row">
            <div class="col">
                <div class="container-fluid">
                    <div class="row">
                        <div class="col alert alert-info">
                            <a href="https://www.reddit.com{{.Permalink}}">{{.Title}}</a>
                            <small>Score: {{.Score}}</small>
                            <small class="text-muted">{{.SubredditName}}</small>
                        </div>
                    </div>
                    {{if not (eq .Url "")}}
                    <div class="row">
                        <div class="col">
                            <a href="{{.Url}}">
                                {{if not (eq .VideoHtml "")}}
                                    <div style="position:relative;padding-bottom:54%"><iframe src="https://gfycat.com/ifr{{.VideoHtml}}" frameborder="0" scrolling="no" width="100%" height="100%" style="position:absolute;top:0;left:0" allowfullscreen></iframe></div>
                                {{else if hasSuffix .ImageUrl ".mp4"}}
                                    <video playsinline autoplay loop controls class="postimage">
                                        <source id="mp4Source" src="{{.ImageUrl}}" type="video/mp4" />
                                    </video>
                                {{else if hasSuffix .ImageUrl ".webm"}}
                                    <video playsinline autoplay loop controls class="postimage">
                                        <source id="webmSource" src="{{.ImageUrl}}" type="video/webm" />
                                    </video>
                                {{else if not (eq .ImageUrl "")}}
                                    <img src="{{.ImageUrl}}" class="img-fluid postimage">
                                {{else}}
                                    {{.Url}}
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
    {{end}}
`
var htmlImageTempl = htmlutil.ParseTemplate(htmlImageTemplateStr)

type pagelink struct {
	Text          string
	Link          string
	IsEnabled     bool
	IsHighlighted bool
}

type annotatedPost struct {
	types.RedditPost
	ImageUrl  string
	VideoHtml string
}

func (this *HtmlViewerRequestHandler) HandleFeed(
	feed *config.RedditFeed,
	pagenum int,
	w http.ResponseWriter,
) {
	if pagenum == 0 {
		pagenum = 1
	}

	posts, err := this.getPosts(feed)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal error retrieving posts for feed: %s %v", feed.Name, err), 500)
		return
	}

	itemsPerPage := 20
	startIdx := itemsPerPage * (pagenum - 1)

	if startIdx >= len(posts) || startIdx+itemsPerPage >= len(posts) {
		// Out of bounds.
		posts = []types.RedditPost{}
	} else {
		posts = posts[startIdx : startIdx+itemsPerPage]
	}

	var annotatedPosts []annotatedPost
	for _, post := range posts {
		a := annotatedPost{RedditPost: post}
		if a.Url != "" {
			a.ImageUrl = htmlutil.RealImageUrl(a.Url)
			if a.ImageUrl == "" {
				u, err := url.Parse(a.Url)
				if err == nil {
					// http://gfycat.com/SomeId --> https://fat.gfycat.com/SomeId.webm or https://giant.gfycat.com/SomeId.mp4
					if toolbox.InDomain("gfycat.com", strings.ToLower(u.Host)) {
						a.VideoHtml = u.Path
					}
				}
			}
		}

		annotatedPosts = append(annotatedPosts, a)
	}

	data := struct {
		Title       string
		Description string
		htmlutil.Breadcrumbs
		Posts     []annotatedPost
		Pagelinks []pagelink
	}{
		Title:       feed.Name,
		Description: feed.Description,
		Breadcrumbs: []htmlutil.Breadcrumb{
			htmlutil.NewBreadcrumb("Home", "/"),
			htmlutil.NewBreadcrumb(feed.Name, "/"),
		},
		Posts:     annotatedPosts,
		Pagelinks: getPagelinks(feed.Name, pagenum),
	}
	htmlutil.RenderTemplate(w, htmlImageTempl, data)
}

func getPagelinks(feedname string, pagenum int) (links []pagelink) {
	link := pagelink{
		Text:          "Previous",
		Link:          ConstructUrl(&feedname, pagenum-1),
		IsEnabled:     true,
		IsHighlighted: false,
	}
	if pagenum == 1 {
		link.IsEnabled = false
	}
	links = append(links, link)

	for i := -2; i < 3; i++ {
		pn := pagenum + i
		if pn <= 0 {
			continue
		}
		link = pagelink{
			Text:          fmt.Sprintf("%d", pn),
			Link:          ConstructUrl(&feedname, pn),
			IsEnabled:     true,
			IsHighlighted: (pagenum == pn),
		}
		links = append(links, link)
	}
	link = pagelink{
		Text:          "Next",
		Link:          ConstructUrl(&feedname, pagenum+1),
		IsEnabled:     true,
		IsHighlighted: false,
	}
	links = append(links, link)
	return
}
