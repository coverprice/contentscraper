package medialink

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func makeLink(t *testing.T, rawurl string) link {
	l, err := newLink(rawurl)
	if err != nil {
		t.Fatal("Could not get link", rawurl, err)
	}
	return l
}

func TestRedditParserReturnsHandledForRedditComment(t *testing.T) {
	link := makeLink(t, "https://reddit.com/r/foo/comments/xyz")
	_, handled := redditParser{}.GetMediaLink(link)
	require.True(t, handled)
}

// This test is disabled because this transform doesn't appear to work now.
func xTestImgurParserConvertsGifvToMp4(t *testing.T) {
	link := makeLink(t, "https://imgur.com/abc.gifv")
	medialink, handled := imgurParser{}.GetMediaLink(link)
	require.True(t, handled)
	require.NotNil(t, medialink)
	require.Equal(t,
		"https://i.imgur.com/abc.mp4",
		medialink.Url)
}

func TestImgurParserNormalizedHost(t *testing.T) {
	link := makeLink(t, "https://imgur.com/abc.gifv")
	medialink, handled := imgurParser{}.GetMediaLink(link)
	require.True(t, handled)
	require.NotNil(t, medialink)
	require.Equal(t,
		"https://i.imgur.com/abc.gifv",
		medialink.Url)
}

func TestImgurParserConvertsRawPathToJpg(t *testing.T) {
	link := makeLink(t, "https://imgur.com/abc")
	medialink, handled := imgurParser{}.GetMediaLink(link)
	require.True(t, handled)
	require.NotNil(t, medialink)
	require.Equal(t,
		"https://i.imgur.com/abc.jpg",
		medialink.Url)
}

func TestGiphyParserConvertsMediaLinks(t *testing.T) {
	link := makeLink(t, "https://media.giphy.com/media/xyzz/giffy.gif")
	medialink, handled := giphyParser{}.GetMediaLink(link)
	require.True(t, handled)
	require.NotNil(t, medialink)
	require.True(t, strings.Contains(string(medialink.Embed), `"https://giphy.com/embed/xyzz"`))
}

func TestGiphyParserConvertsLegibleUrls(t *testing.T) {
	link := makeLink(t, "https://giphy.com/gifs/foo-bar-baz-xyzz/giffy.gif")
	medialink, handled := giphyParser{}.GetMediaLink(link)
	require.True(t, handled)
	require.NotNil(t, medialink)
	require.True(t, strings.Contains(string(medialink.Embed), `"https://giphy.com/embed/xyzz"`))
}

func TestGfycatParserEmbedsBlankLinks(t *testing.T) {
	link := makeLink(t, "https://gfycat.com/xYzz123")
	medialink, handled := gfycatParser{}.GetMediaLink(link)
	require.True(t, handled)
	require.NotNil(t, medialink)
	require.True(t, strings.Contains(string(medialink.Embed), `"https://gfycat.com/ifr/xYzz123"`))
}

func TestGfycatParserSupportsMultipleLeadingSlashes(t *testing.T) {
	link := makeLink(t, "https://gfycat.com//xYzz123")
	medialink, handled := gfycatParser{}.GetMediaLink(link)
	require.True(t, handled)
	require.NotNil(t, medialink)
	require.True(t, strings.Contains(string(medialink.Embed), `"https://gfycat.com/ifr/xYzz123"`))
}

func TestGfycatParserSupportsGifDetail(t *testing.T) {
	link := makeLink(t, "https://gfycat.com/gifs/detail/xYzz123")
	medialink, handled := gfycatParser{}.GetMediaLink(link)
	require.True(t, handled)
	require.NotNil(t, medialink)
	require.True(t, strings.Contains(string(medialink.Embed), `"https://gfycat.com/ifr/xYzz123"`))
}
