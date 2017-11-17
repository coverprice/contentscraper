package medialink

import (
	"github.com/coverprice/contentscraper/toolbox"
	"html/template"
	"regexp"
	"strings"
)

type iSiteUrlParser interface {
	GetMediaLink(
		l link,
	) (
		ml *MediaLink, // The raw URL or HTML to embed this URL. May be nil.
		// True when this parser has successfully handled the input URL, meaning
		// that the caller should not call any further drivers.
		handled bool,
	)
}

// --------------------------------

type redditParser struct{}

func (this redditParser) GetMediaLink(l link) (ml *MediaLink, handled bool) {
	if !toolbox.InDomain("reddit.com", l.Host) {
		return nil, false
	}
	switch {
	case toolbox.MatchString(`^/r/[^/]+/comments`, l.Path):
		// Links to comments are skipped
		return nil, true
	default:
		return nil, false
	}
}

// --------------------------------

type imgurParser struct{}

func (this imgurParser) GetMediaLink(l link) (ml *MediaLink, handled bool) {
	if !toolbox.InDomain("imgur.com", l.Host) {
		return nil, false
	}
	switch {
	// http://i.imgur.com/foooo.gifv --> http://i.imgur.com/foooo.mp4
	case toolbox.MatchString(`\.gifv$`, l.Path):
		l.Url.Host = "i.imgur.com"
		l.Url.Path = strings.TrimSuffix(l.Url.Path, ".gifv") + ".mp4"
		return &MediaLink{Url: l.Url.String()}, true

	// http://imgur.com/foooo --> http://i.imgur.com/2iiK88I.jpg
	case toolbox.MatchString(`^/[[:alnum:]]+$`, l.Url.Path):
		l.Url.Host = "i.imgur.com"
		l.Url.Path = l.Url.Path + ".jpg"
		return &MediaLink{Url: l.Url.String()}, true

	default:
		return nil, false
	}
}

// --------------------------------

type giphyParser struct{}

// Matches https://giphy.com/media/xxxxxxxxxxx/giphy.gif
var giphyMediaRe = regexp.MustCompile(`^/media/([[:alnum:]]+)`)

// Matches https://giphy.com/gifs/some-url-xxxxxxxxxx
var giphyLegibleUrlRe = regexp.MustCompile(`^/gifs/(?:[[:alnum:]]+-)*([[:alnum:]]+)`)

func (this giphyParser) GetMediaLink(l link) (ml *MediaLink, handled bool) {
	if !toolbox.InDomain("giphy.com", l.Host) {
		return nil, false
	}
	var id string

	// https://media.giphy.com/media/xxxxxxxxxxxxx/giphy.gif
	matches := giphyMediaRe.FindStringSubmatch(l.Url.Path)
	if matches != nil && len(matches) > 1 {
		id = matches[1]
	} else {
		matches = giphyLegibleUrlRe.FindStringSubmatch(l.Url.Path)
		if matches != nil && len(matches) > 1 {
			id = matches[1]
		}
	}
	if id != "" {
		html := `<iframe src="https://giphy.com/embed/` + id + `"` +
			` frameborder="0" scrolling="no" width="100%" height="100%"` +
			` class="giphy-embed" allowFullScreen></iframe>`
		return &MediaLink{Embed: template.HTML(html)}, true
	}
	return nil, false
}

// --------------------------------

type gfycatParser struct{}

// http://gfycat.com/SomeId
var gfycatIdRe = regexp.MustCompile(`^/+(?:gifs/detail/)?([[:alnum:]]+)`)

func (this gfycatParser) GetMediaLink(l link) (ml *MediaLink, handled bool) {
	if !toolbox.InDomain("gfycat.com", l.Host) {
		return nil, false
	}
	matches := gfycatIdRe.FindStringSubmatch(l.Url.Path)
	if matches != nil && len(matches) > 1 {

		// This is from gfycat's website.
		html := `<div style="position:relative;padding-bottom:54%">` +
			`<iframe src="https://gfycat.com/ifr/` + matches[1] + `"` +
			` frameborder="0" scrolling="no" width="100%" height="100%"` +
			` style="position:absolute;top:0;left:0" allowfullscreen>` +
			`</iframe>` +
			`</div>`
			// template.HTML() means "don't escape this when rendered"
		return &MediaLink{Embed: template.HTML(html)}, true
	}
	return nil, false
}

// --------------------------------

type otherParser struct{}

// Note: "(?i)" means set the "i" flag (case-insensitive)
// "(?:" begins a non-capturing group)
var standardGraphicsSuffixesRe = regexp.MustCompile(`(?i)\.(?:gifv?|png|jpe?g|mp4|webm)$`)

func (this otherParser) GetMediaLink(l link) (ml *MediaLink, handled bool) {
	if standardGraphicsSuffixesRe.MatchString(l.Path) {
		// This is a standard-looking image URL, return verbatim.
		return &MediaLink{Url: l.RawUrl}, true
	}
	return nil, false
}
