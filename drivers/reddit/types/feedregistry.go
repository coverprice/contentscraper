package types

import (
	"fmt"
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/drivers"
)

type FeedRegistryItem struct {
	config.RedditFeed
	Status            drivers.FeedHarvestStatus
	TimeLastHarvested int64
}

type TFeedRegistry map[string]*FeedRegistryItem

// The FeedRegistry is a simple map (with some functions to assist with adding/retrieving)
// that maps a Feed name to the Reddit configuration.
var FeedRegistry = make(TFeedRegistry)

func (this *TFeedRegistry) AddItem(feed *config.RedditFeed) {
	fri := FeedRegistryItem{
		RedditFeed:        *feed,
		Status:            drivers.FEEDHARVESTSTATUS_IDLE,
		TimeLastHarvested: 0,
	}

	(*this)[fri.RedditFeed.Name] = &fri
}

func (this *TFeedRegistry) GetItemByName(feedname string) (*FeedRegistryItem, error) {
	item, ok := (*this)[feedname]
	if !ok {
		return nil, fmt.Errorf("Unknown feed name %s", feedname)
	}
	return item, nil
}

func (this *TFeedRegistry) GetAllItems() (ret []*FeedRegistryItem) {
	for _, val := range *this {
		ret = append(ret, val)
	}
	return ret
}
