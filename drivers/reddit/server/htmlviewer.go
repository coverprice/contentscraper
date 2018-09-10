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
    <script src="https://unpkg.com/imagesloaded@4/imagesloaded.pkgd.min.js"></script>
    <script>
    // Returns the scale factor required to get the image dimensions to fit into the given window
    // dimensions.
    function getScaleFactor(img_w, img_h, window_w, window_h) {
        let scale_factor = window_w / img_w;
        return (scale_factor * img_h > window_h) ? window_h / img_h : scale_factor;
    }
    // Scale down images that are wider than the page so they fit on the page.
    // Waits for naturalHeight/Width to become available by using a jQuery plugin.
    $(document).imagesLoaded().progress(function(instance, image) {
        let el = image.img;
        if (!image.isLoaded) {
            console.log("Failed to load image: " + el.src);
	    $(el).closest('div.feeditem').remove();
            return;
        }
        let el_w = el.naturalWidth || el.videoWidth || el.width;
        let el_h = el.naturalHeight || el.videoHeight || el.height;
        if (!el_w || !el_h) {
            console.log("Error getting width/height for image: " + el.src);
            return;
        }
        let max_w = window.innerWidth - el.x - 50;
        let max_h = window.innerHeight - 100;
        let scale_factor = getScaleFactor(el_w, el_h, max_w, max_h)
        let new_w = Math.floor(scale_factor * el_w);
        let new_h = Math.floor(scale_factor * el_h);
        /*
            console.log(
              "Processing image: "+
              "  NW/NH: " + el_w + "," + el_h + " (" + (el_w/el_h).toFixed(4) + ")" +
              "  Window W/H: " + max_w + "," + max_h + " (" + (max_w/max_h).toFixed(4) + ")" +
              "  Scalefactor: " + scale_factor.toFixed(4) +
              "  New W/H: " + new_w + "," + new_h + " (" + (new_w/new_h).toFixed(4) + ")" +
              "  "+el.src
            );
        */
        el.style.width = new_w + "px";
        el.style.height = new_h + "px";
    });
    $(document).ready(function() {
        let max_height = window.innerHeight - 100;
	$('.videocontainer').each(function(idx, el) {
		el.style.maxHeight = max_height + "px";
	});
    });

    let numPages = {{.NumPages}};
    let pageNum = {{.PageNum}};

    function scrollToNextItem(is_up) {
       let window_top_y = $(window).scrollTop();
       let new_top_y = 0;

       let items = [0, window_top_y];
       $(".feeditem").each(function(idx, el) {
           items.push(Math.floor($(el).offset().top));
       })
       if (is_up) {
          items.sort((a, b) => b-a);	// Descending order
          new_top_y = items.find((val) => val < window_top_y) || 0;
       } else {
          items.sort((a, b) => a-b);	// Ascending order
          new_top_y = items.find((val) => val > window_top_y) || items.pop();
       }
       window.scrollTo(0, new_top_y);
    }
    $(document).keypress(function(event) {
        let key = String.fromCharCode(event.which);
        if (key == "k" || key == "j") {               // Up/Down
            scrollToNextItem(key == "k")

        } else if (key == "h" && pageNum > 1) {        // Previous page
            window.location = '{{.PreviousPagelink.Link}}';

        } else if (key == "l" && pageNum < numPages) {        // Next page
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
