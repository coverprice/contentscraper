package types

import (
	"fmt"
	"github.com/coverprice/contentscraper/config"
	"github.com/coverprice/contentscraper/drivers"
)

type FeedRegistryItem struct {
	config.RedditFeed
	Status            drivers.FeedStatus
	TimeLastHarvested int64
}

type TFeedRegistry map[string]*FeedRegistryItem

var FeedRegistry = make(TFeedRegistry)

func (this *TFeedRegistry) AddItem(feed *config.RedditFeed) {
	fri := FeedRegistryItem{
		RedditFeed:        *feed,
		Status:            drivers.FEEDSTATUS_IDLE,
		TimeLastHarvested: 0,
	}

	(*this)[fri.RedditFeed.Name] = &fri
}

// TODO: Remove this if not used anywhere
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
