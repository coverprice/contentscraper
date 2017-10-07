package htmlutil

import (
	"fmt"
	"github.com/coverprice/contentscraper/toolbox"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/url"
	"regexp"
	"strings"
)

// MediaLink contains either a Url to an image/video, or raw HTML that will embed such an image/video.
// It is used as the result from RealMediaUrl, which attempts to analyze a Url to a site like imgur.com
// or gfycat.com, and return a way of rendering that content.
type MediaLink struct {
	Url   string        // Direct link (must construct <img> or <video> link yourself)
	Embed template.HTML // raw HTML that will embed the image.
}

// Note: "(?i)" means set the "i" flag (case-insensitive)
// "(?:" begins a non-capturing group)
var standardGraphicsSuffixesRe = regexp.MustCompile(`(?i)\.(?:gif|png|jpe?g|mp4|webm)$`)

// Attempts to convert an image/video URL to to a MediaLink, which can be used
// to directly display the image/video within the page. Direct URLs are preferred
// so as to give the most control to the template (for sizing, etc), but as a last
// resort it will return raw HTML that will embed the media.
// E.g.URLs like "http://example.com/foo.jpg" will be returned unmodified,
// "http://imgur.com/foo" will be converted to "http://i.imgur.com/foo.jpg",
// and "gfycat.com/foo" will return HTML to embed it in the page (since it's
// difficult to reliably convert that URL into a direct video link)
func UrlToEmbedUrl(rawurl string) (link *MediaLink, err error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Warningf("Could not parse the URL: '%s' %v", rawurl, err)
		return
	}

	// Let's make pattern matching a littler easier:
	u.Scheme = strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Host)
	if strings.Index(host, ":") != -1 {
		host = host[:strings.Index(host, ":")]
	}
	path := strings.ToLower(u.Path)

	if !(u.Scheme == "http" || u.Scheme == "https") {
		err = fmt.Errorf("Non HTTP Scheme in URL: '%s", rawurl)
		return
	}

	if standardGraphicsSuffixesRe.MatchString(path) {
		// This is a standard-looking image URL, return verbatim.
		return &MediaLink{
			Url: rawurl,
		}, nil
	}

	// http://i.imgur.com/foooo.gifv --> http://i.imgur.com/foooo.mp4
	if toolbox.InDomain("imgur.com", host) && toolbox.MatchString(`\.gifv$`, path) {
		u.Host = "i.imgur.com"
		u.Path = strings.TrimSuffix(u.Path, ".gifv") + ".mp4"
		return &MediaLink{
			Url: u.String(),
		}, nil
	}

	// http://imgur.com/foooo --> http://i.imgur.com/2iiK88I.jpg
	if toolbox.InDomain("imgur.com", host) && toolbox.MatchString(`^/[[:alnum:]]+$`, u.Path) {
		u.Host = "i.imgur.com"
		u.Path = u.Path + ".jpg"
		return &MediaLink{
			Url: u.String(),
		}, nil
	}

	// http://gfycat.com/SomeId --> embedded HTML
	if toolbox.InDomain("gfycat.com", strings.ToLower(u.Host)) {
		// This is from gfycat's website.
		html := `<div style="position:relative;padding-bottom:54%">` +
			`<iframe src="https://gfycat.com/ifr%s" frameborder="0" scrolling="no" width="100%" height="100%" style="position:absolute;top:0;left:0" allowfullscreen>` +
			`</iframe>` +
			`</div>`
		html = fmt.Sprintf(html, u.Path)
		return &MediaLink{
			Embed: template.HTML(html), // template.HTML() means "don't escape this when rendered"
		}, nil
	}

	log.Debugf("Could not embed URL: '%s'", rawurl)
	return nil, nil
}
