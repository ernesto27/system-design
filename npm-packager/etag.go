package main

type Etag struct {
	packages map[string]Dependency
}

func newEtag(packages map[string]Dependency) *Etag {
	return &Etag{packages: packages}
}

func (e *Etag) save() error {
	return nil
}
