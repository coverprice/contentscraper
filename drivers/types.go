package drivers

type ParamBag map[string]interface{}
type IPost interface{}

type IScraper interface {
	Scrape(paramBag ParamBag) ([]IPost, error)
}

type IDataStore interface {
	StorePost(post *IPost) error
}
