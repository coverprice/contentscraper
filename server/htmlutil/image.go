package htmlutil

import (
	"github.com/coverprice/contentscraper/toolbox"
	log "github.com/sirupsen/logrus"
	"net/url"
	"regexp"
	"strings"
)

// Note: "(?i)" means set the "i" flag (case-insensitive)
// "(?:" begins a non-capturing group)
var standardGraphicsSuffixesRe = regexp.MustCompile(`(?i)\.(?:gif|png|jpe?g|mp4|webm)$`)

// Converts a URL to (possibly) an image into a direct link to that image.
// For URLs like "http://example.com/foo.jpg", the result will be unmodified.
// For other URLs such as gfycat or imgur links, the URL may be interpreted
// as a link to a container page, so the resulting URL is a direct link to
// the content (to display in either a <video> or <img> tag).
// If the URL can't be interpreted, an empty string is returned.
func RealImageUrl(rawurl string) string {
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

	// http://i.imgur.com/foooo.gifv --> http://i.imgur.com/foooo.mp4
	if toolbox.InDomain("imgur.com", host) && toolbox.MatchString(`\.gifv$`, path) {
		u.Host = "i.imgur.com"
		u.Path = strings.TrimSuffix(u.Path, ".gifv") + ".mp4"
		return u.String()
	}

	// http://imgur.com/foooo --> http://i.imgur.com/2iiK88I.jpg
	if toolbox.InDomain("imgur.com", host) && toolbox.MatchString(`^/[[:alnum:]]+$`, u.Path) {
		u.Host = "i.imgur.com"
		u.Path = u.Path + ".jpg"
		return u.String()
	}

	// http://gfycat.com/SomeId --> https://fat.gfycat.com/SomeId.webm or https://giant.gfycat.com/SomeId.mp4
	if toolbox.InDomain("gfycat.com", host) {
		u.Host = "fat.gfycat.com"
		u.Path = u.Path + ".webm"
		return u.String()
	}

	log.Debugf("Couldn't work out the graphics link for this URL: '%s'", rawurl)
	return ""
}
