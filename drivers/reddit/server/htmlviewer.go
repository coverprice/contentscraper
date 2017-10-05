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
	"net/url"
	"regexp"
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
                    <div class="row">
                        <div class="col">
                            <a href="{{.Url}}">
                                {{if hasSuffix .ImageUrl ".mp4"}}
                                    <video playsinline autoplay loop controls>
                                        <source id="mp4Source" src="{{.ImageUrl}}" type="video/mp4" />
                                    </video>
                                {{else if hasSuffix .ImageUrl ".webm"}}
                                    <video playsinline autoplay loop controls>
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

type annotatedPost struct {
	types.RedditPost
	ImageUrl string
}

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

	var annotatedPosts []annotatedPost
	for _, post := range posts {
		annotatedPosts = append(annotatedPosts, annotatePost(post))
	}

	data := struct {
		Title       string
		Description string
		htmlutil.Breadcrumbs
		Posts []annotatedPost
	}{
		Title:       feed.Name,
		Description: feed.Description,
		Breadcrumbs: []htmlutil.Breadcrumb{
			htmlutil.NewBreadcrumb("Home", "/"),
			htmlutil.NewBreadcrumb(feed.Name, "/"),
		},
		Posts: annotatedPosts,
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
          AND is_active = 1
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

func annotatePost(p types.RedditPost) annotatedPost {
	var a annotatedPost
	a = annotatedPost{RedditPost: p}

	if a.Url != "" {
		a.ImageUrl = makeImageViewable(a.Url)
	}
	return a
}

// Note: "(?i)" means set the "i" flag (case-insensitive)
// "(?:" begins a non-capturing group)
var standardGraphicsSuffixesRe = regexp.MustCompile(`(?i)\.(?:gif|png|jpe?g|mp4|webm)$`)

// isDomain returns true if the candidate hostname == or is sub-ordinate to the given domain,
// e.g. if domain is "foo.com" then "foo.com" and "sub.foo.com" will match, but "barfoo.com" won't.
func isDomain(domain, candidate string) bool {
	return isMatch(`(?i)(?:^|\.)`+domain+`$`, candidate)
}
func isMatch(pattern, candidate string) bool {
	result, err := regexp.MatchString(pattern, candidate)
	if err != nil {
		log.Fatal("Invalid pattern: %s", pattern)
	}
	return result
}

func makeImageViewable(rawurl string) string {
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Warningf("Could not parse the URL: '%s' %v", rawurl, err)
		return ""
	}

	// Let's make pattern matching a littler easier:
	u.Scheme = strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Host)
	if strings.Index(host, ":") != -1 {
		host = host[:strings.Index(host, ":")]
	}
	path := strings.ToLower(u.Path)

	if !(u.Scheme == "http" || u.Scheme == "https") {
		log.Warningf("Non HTTP Scheme in URL: '%s", rawurl)
		return ""
	}

	if standardGraphicsSuffixesRe.MatchString(path) {
		// This is a standard-looking image URL, return verbatim.
		return rawurl
	}

	// Now to handle all the crazy special cases!

	// http://i.imgur.com/foooo.gifv --> http://i.imgur.com/foooo.mp4
	if isDomain("imgur.com", host) && isMatch(`\.gifv$`, path) {
		u.Host = "i.imgur.com"
		u.Path = strings.TrimSuffix(u.Path, ".gifv") + ".mp4"
		return u.String()
	}

	// http://gfycat.com/SomeId --> https://fat.gfycat.com/SomeId.webm or https://giant.gfycat.com/SomeId.mp4
	if isDomain("gfycat.com", host) {
		u.Host = "fat.gfycat.com"
		u.Path = u.Path + ".webm"
		return u.String()
	}

	// http://imgur.com/foooo --> http://i.imgur.com/2iiK88I.jpg
	if isDomain("imgur.com", host) && isMatch(`^/[[:alnum:]]+$`, u.Path) {
		u.Host = "i.imgur.com"
		u.Path = u.Path + ".jpg"
		return u.String()
	}

	log.Infof("Couldn't work out the graphics link for this URL: '%s'", rawurl)
	return ""
}
