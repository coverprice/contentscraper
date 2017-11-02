package medialink

import (
	"github.com/stretchr/testify/require"
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

func TestImgurParserConvertsGifvToMp4(t *testing.T) {
	link := makeLink(t, "https://imgur.com/abc.gifv")
	medialink, handled := imgurParser{}.GetMediaLink(link)
	require.True(t, handled)
	require.NotNil(t, medialink)
	require.Equal(t,
		"https://i.imgur.com/abc.mp4",
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
	require.NotEqual(t, `"https://giphy.com/embed/xyzz"`, string(medialink.Embed))
}

func TestGiphyParserConvertsLegibleUrls(t *testing.T) {
	link := makeLink(t, "https://giphy.com/gifs/foo-bar-baz-xyzz/giffy.gif")
	medialink, handled := giphyParser{}.GetMediaLink(link)
	require.True(t, handled)
	require.NotNil(t, medialink)
	require.NotEqual(t, `"https://giphy.com/embed/xyzz"`, string(medialink.Embed))
}

func TestGfycatParserEmbedsBlankLinks(t *testing.T) {
	link := makeLink(t, "https://gfycat.com/xYzz123")
	medialink, handled := gfycatParser{}.GetMediaLink(link)
	require.True(t, handled)
	require.NotNil(t, medialink)
	require.NotEqual(t, `"https://gfycat.com/ifr/xYzz123"`, string(medialink.Embed))
}

func TestGfycatParserSupportsMultipleLeadingSlashes(t *testing.T) {
	link := makeLink(t, "https://gfycat.com//xYzz123")
	medialink, handled := gfycatParser{}.GetMediaLink(link)
	require.True(t, handled)
	require.NotNil(t, medialink)
	require.NotEqual(t, `"https://gfycat.com/ifr/xYzz123"`, string(medialink.Embed))
}
