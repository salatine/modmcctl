package core

import (
	"os"
	"fmt"
	"sync"
	"strings"
	"path/filepath"
	"github.com/salatine/modmcctl/internal/cli"
	"github.com/salatine/modmcctl/loaders"
	"github.com/salatine/modmcctl/providers"
)

func Run(cfg *cli.Config) error {
	modsDirs := resolveModsDirs(cfg)

	if err := loaders.InstallLoader(
		cfg.Loader,
		cfg.Mode,
		cfg.ClientDir,
		cfg.ServerDir,
		cfg.Version,
	); err != nil {
		return err
	}

	provider := resolveProvider(cfg.Provider)

	return downloadMods(provider, cfg, modsDirs)
}

func resolveModsDirs(cfg *cli.Config) []string {
	var dirs []string

	if cfg.Mode == "client" || cfg.Mode == "both" {
		dirs = append(dirs, filepath.Join(cfg.ClientDir, "mods"))
	}
	if cfg.Mode == "server" || cfg.Mode == "both" {
		dirs = append(dirs, filepath.Join(cfg.ServerDir, "mods"))
	}

	return dirs
}

func resolveProvider(name string) providers.ModProvider {
	if name == "curseforge" {
		return &providers.CurseForgeProvider{}
	}
	return &providers.ModrinthProvider{}
}

func downloadMods(p providers.ModProvider, cfg *cli.Config, modsDirs []string) error {
	var mcDirs []string
	for _, modsDir := range modsDirs {
		mcDirs = append(mcDirs, filepath.Dir(modsDir))
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

	for _, slug := range cfg.Mods {
		slug = strings.TrimSpace(slug)
		if slug == "" {
			continue
		}

		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			fileDownload, isModpack, err := p.Fetch(s, cfg.Version, cfg.Loader)
			if err != nil {
				fmt.Println("error:", err)
				return
			}

			if isModpack {
				var mods []*providers.Downloadable
				for _, mcDir := range mcDirs {
					if mods, err = p.FetchModpack(fileDownload, mcDir); err != nil {
						fmt.Println("error:", err)
						return
					}
				}

				for _, mod := range mods {
					download(mod, modsDirs)
				}
			} else {
				download(fileDownload, modsDirs)
			}


		}(slug)
	}

	wg.Wait()
	return nil
}

func download(downloadable *providers.Downloadable, dirs []string) {
	for _, dir := range dirs {
		cli.EnsureDir(dir)
		path := filepath.Join(dir, downloadable.Filename)

		if _, err := os.Stat(path); err == nil {
			fmt.Println("skip:", downloadable.Filename)
			continue
		}

		fmt.Println("downloading:", downloadable.Filename)
		cli.DownloadFile(downloadable.URL, path)
	}
}
