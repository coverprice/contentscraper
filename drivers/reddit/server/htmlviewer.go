package server

import (
	"fmt"
	"github.com/coverprice/contentscraper/config"
	persist "github.com/coverprice/contentscraper/drivers/reddit/persistence"
	"github.com/coverprice/contentscraper/drivers/reddit/types"
	"github.com/coverprice/contentscraper/server/htmlutil"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	NUM_ITEMS_PER_PAGE = 20
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

    let itemOffset = 0;
    let numItemsPerPage = {{.NumItemsPerPage}};
    let pageNum = {{.PageNum}};

    function scrollToItem() {
        window.scrollTo(0, $('#item' + itemOffset).offset().top);
    }
    $(document).keypress(function(event) {
        let key = String.fromCharCode(event.which);
        if (key == "k") {               // Up
            if (itemOffset > 0) {
                itemOffset--;
                scrollToItem();
            }
        } else if (key == "j") {        // Down
            if (itemOffset < numItemsPerPage - 1) {
                itemOffset++;
                scrollToItem();
            }
        } else if (key == "h") {        // Previous
            if (pageNum > 1) {
                window.location = '{{.PreviousPagelink.Link}}';
            }
        } else if (key == "l") {        // Next
            window.location = '{{.NextPagelink.Link}}';
        } else if (key == "i") {        // Home
            window.location = '/';
        } else {
            return;
        }
        event.preventDefault();
    });
	</script>
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
        <div class="row" id="item{{$itemIndex}}">
            <div class="col">
                <div class="container-fluid">
                    <div class="row">
                        <div class="col alert alert-info">
                            <a href="https://www.reddit.com{{.Permalink}}">{{.Title}}</a>
                            <small>Score: {{.Score}}</small>
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
                                    <video playsinline autoplay loop controls class="postimage">
                                        <source src="{{.MediaLink.Url}}" type="video/mp4" />
                                    </video>
                                {{else if hasSuffix .MediaLink.Url ".webm"}}
                                    <video playsinline autoplay loop controls class="postimage">
                                        <source src="{{.MediaLink.Url}}" type="video/webm" />
                                    </video>
                                {{else if not (eq .MediaLink.Url "")}}
                                    <img src="{{.MediaLink.Url}}" class="img-fluid postimage">
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

type annotatedPost struct {
	types.RedditPost
	MediaLink *htmlutil.MediaLink
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

	itemsPerPage := NUM_ITEMS_PER_PAGE
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
			if a.MediaLink, err = htmlutil.UrlToEmbedUrl(a.Url); err != nil {
				log.Error("Error trying to convert post URL to MediaLink", err)
			}
		}
		annotatedPosts = append(annotatedPosts, a)
	}

	pagelinks := getPagelinks(feed.Name, pagenum)
	data := struct {
		Title       string
		Description string
		htmlutil.Breadcrumbs
		Posts            []annotatedPost
		Pagelinks        []pagelink
		PreviousPagelink pagelink
		NextPagelink     pagelink
		NumItemsPerPage  int
		PageNum          int
	}{
		Title:       feed.Name,
		Description: feed.Description,
		Breadcrumbs: []htmlutil.Breadcrumb{
			htmlutil.NewBreadcrumb("Home", "/"),
			htmlutil.NewBreadcrumb(feed.Name, "/"),
		},
		Posts:            annotatedPosts,
		Pagelinks:        pagelinks,
		PreviousPagelink: pagelinks[0], // Clunky, but necessary since arithmetic isn't possible in templates.
		NextPagelink:     pagelinks[len(pagelinks)-1],
		NumItemsPerPage:  itemsPerPage,
		PageNum:          pagenum,
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
