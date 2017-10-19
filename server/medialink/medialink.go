package medialink

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/url"
	"strings"
)

// MediaLink contains either a Url to an image/video, or raw HTML that will embed such an image/video.
// It is used as the result from RealMediaUrl, which attempts to analyze a Url to a site like imgur.com
// or gfycat.com, and return a way of rendering that content.
type MediaLink struct {
	Url   string        // Direct link (must construct <img> or <video> link yourself)
	Embed template.HTML // raw HTML that will embed the image.
}

type link struct {
	RawUrl string
	Url    *url.URL
	Host   string
	Path   string
}

func newLink(rawurl string) (l link, err error) {
	l.RawUrl = rawurl
	l.Url, err = url.Parse(rawurl)
	if err != nil {
		log.Warningf("Could not parse the URL: '%s' %v", rawurl, err)
		return
	}

	l.Url.Scheme = strings.ToLower(l.Url.Scheme)
	if !(l.Url.Scheme == "http" || l.Url.Scheme == "https") {
		err = fmt.Errorf("Non HTTP Scheme in URL: '%s", rawurl)
		return
	}

	// Let's make pattern matching a littler easier:
	l.Host = strings.ToLower(l.Url.Host)
	if strings.Index(l.Host, ":") != -1 {
		l.Host = l.Host[:strings.Index(l.Host, ":")]
	}
	l.Path = strings.ToLower(l.Url.Path)
	return
}

var parsers = []iSiteUrlParser{
	&redditParser{},
	&imgurParser{},
	&giphyParser{},
	&gfycatParser{},
	&otherParser{},
}

// Attempts to convert an image/video URL to to a MediaLink, which can be used
// to directly display the image/video within the page. Direct URLs are preferred
// so as to give the most control to the template (for sizing, etc), but as a last
// resort it will return raw HTML that will embed the media.
// E.g.URLs like "http://example.com/foo.jpg" will be returned unmodified,
// "http://imgur.com/foo" will be converted to "http://i.imgur.com/foo.jpg",
// and "gfycat.com/foo" will return HTML to embed it in the page (since it's
// difficult to reliably convert that URL into a direct video link)
func UrlToMediaLink(rawurl string) (link *MediaLink, err error) {
	l, err := newLink(rawurl)
	if err != nil {
		log.Warningf("Could not parse the URL: '%s' %v", rawurl, err)
		return
	}

	for _, parser := range parsers {
		var handled bool
		if link, handled = parser.GetMediaLink(l); handled == true {
			return
		}
	}
	log.Debugf("Could not embed URL: '%s'", rawurl)
	return nil, nil
}
