package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	mode := flag.String("mode", "client", "client|server|both")
	clientDirFlag := flag.String("client-dir", "", ".minecraft client absolute path")
	serverDirFlag := flag.String("server-dir", "", ".minecraft server absolute path")
	version := flag.String("version", "", "minecraft version")
	loader := flag.String("loader", "", "neoforge|fabric")
	providerName := flag.String("provider", "modrinth", "modrinth|curseforge")
	modsFlag := flag.String("mods", "", "list of mods slugs/names, separated by comma")
	flag.Parse()

	if *version == "" || *loader == "" {
		flag.Usage()
		os.Exit(1)
	}

	clientDir := getClientDir(*clientDirFlag)
	serverDir := getServerDir(*serverDirFlag)
	modsDirs := make([]string, 0)

	if *mode == "client" || *mode == "both" {
		modsDirs = append(modsDirs, filepath.Join(clientDir, "mods"))
	}
	if *mode == "server" || *mode == "both" {
		modsDirs = append(modsDirs, filepath.Join(serverDir, "mods"))
	}


	if err := installLoader(*loader, *mode, clientDir, serverDir, *version); err != nil {
		fmt.Println("error: ", err)
		os.Exit(1)
	}

	var provider ModProvider
	if *providerName == "curseforge" {
		provider = &CurseForgeProvider{}
	} else {
		provider = &ModrinthProvider{}
	}


	if *modsFlag != "" && provider != nil {
		modList := strings.Split(*modsFlag, ",")
		var wg sync.WaitGroup
		sem := make(chan struct{}, 5)

		for _, slug := range modList {
			slug = strings.TrimSpace(slug)
			if slug == "" { continue }

			wg.Add(1)
			go func(s string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				url, filename, err := provider.FetchMod(s, *version, *loader)
				if err != nil {
					fmt.Println("error: ", err)
					return
				}
				for _, modsDir := range modsDirs {
					ensureDir(modsDir)
					path := filepath.Join(modsDir, filename)
					if _, err := os.Stat(path); err == nil {
						fmt.Println("skip: ", filename)
						return
					}

					fmt.Println("downloading mod: ", filename)
					downloadFile(url, path)
				}

			}(slug)
		}
		wg.Wait()
	}
}
