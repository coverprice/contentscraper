package drivers

type FeedName string
type RssFeed string

type IDriver interface {
	Harvest() error
	// MarkForPublishing() error
	// GetFeed(feedname FeedName) (RssFeed, error)
}

// --------------------------------------

// A string that uniquely identifies a SourceConfig. Used to store
// its "LastRun" data. E.g. "reddit/r/some_subreddit" or "twitter"
type SourceConfigId string

// ISourceConfig describes a data source to the backend, which may have extra details
// about a sub-stream, e.g. a specific sub-reddit.
type ISourceConfig interface {
	GetSourceConfigId() SourceConfigId
}

/*


// --------------------------------------

// FilterCriteria is a DO that contains the parameters used for filtering persisted data.
// It will typically be wrapped + extended to add information about (for example)
// a specific subreddit, a twitter account + "is RT vs regular tweet"
type FilterCriteria struct {
	Percentile   float64
	MaxPostCount int
}

// Again, trivial interface+implementation to allow type checking when the above is included
// anonymously in implementations.
type IFilterCriteria interface {
	GetPercentile() float64
}

func (fc *FilterCriteria) GetPercentile() float64 {
	return fc.Percentile
}

// --------------------------------------

type FeedName string

type Feed struct {
	Name           FeedName
	Description    string
	driver         IDriver
	SourceConfigs  []ISourceConfig
	SourceLastRun  map[SourceConfigId]SourceLastRun
	FilterCriteria []IFilterCriteria
}

func (this *Feed) ScrapePosts(config ISourceConfig, startDate, endDate uint64) (err error) {
	return this.driver.ScrapePosts(config, startDate, endDate)
}

func (this *Feed) FilterByCriteria() (err error) {
	for _, criteria := range this.FilterCriteria {
		this.driver.FilterByCriteria(criteria)
	}
}

*/
