package providers

type Downloadable struct {
	URL string
	Filename string
}

type ModProvider interface {
	Fetch(slug, mcVersion, loader string) (mod *Downloadable, isModpack bool, err error)
	FetchModpack(pack *Downloadable) (mods []*Downloadable, err error)
}
