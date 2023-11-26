package urlfrontier

type URLFrontier struct {
	urls []string
}

func New(urls []string) *URLFrontier {
	// todo,  do some logic validation
	return &URLFrontier{urls: urls}
}
