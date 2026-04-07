package providers

type ModProvider interface {
	FetchMod(slug, mcVersion, loader string) (url, filename string, err error)
}
