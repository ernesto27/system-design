package seedurl

type SeedURL struct {
	Urls []string
}

func New() *SeedURL {
	// hardcode for now, this should be get for database
	urls := []string{"https://news.ycombinator.com/", "https://www.infobae.com/", "https://thenewstack.io/"}
	//urls := []string{"https://news.ycombinator.com/"}
	return &SeedURL{Urls: urls}
}
