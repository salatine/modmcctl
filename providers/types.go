package providers

type ModDownload struct {
	URL string
	Filename string
}

type ModProvider interface {
	Fetch(slug, mcVersion, loader string) (mod *ModDownload, isModpack bool, err error)
	FetchModpack(pack *ModDownload) (mods []*ModDownload, err error)
}
