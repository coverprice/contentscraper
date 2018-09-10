package server

import (
	"fmt"
	"net/url"
)

const (
	BaseUrlPath = "/reddit/"
)

func constructUrl(feedname *string, pagenum int) string {
	v := url.Values{}
	if feedname != nil {
		v.Set("feed", *feedname)
	}
	if pagenum != 0 {
		v.Add("page", fmt.Sprintf("%d", pagenum))
	}
	u := url.URL{
		Path:     BaseUrlPath,
		RawQuery: v.Encode(),
	}
	return u.String()
}
