package runner

type ParamBag map[string]interface{}
type IPost interface{}

type IScraper interface {
	Scrape(paramBag ParamBag) ([]IPost, error)
}

type IDataStore interface {
	StorePost(post *IPost) error
}

type ScraperRunner struct {
	Scraper   IScraper
	Datastore IDataStore
}

func (sr *ScraperRunner) Run(paramBag ParamBag) (err error) {
	var posts []IPost
	if posts, err = sr.Scraper.Scrape(paramBag); err != nil {
		return
	}
	for _, post := range posts {
		if err = sr.Datastore.StorePost(&post); err != nil {
			return
		}
	}
	return nil
}
