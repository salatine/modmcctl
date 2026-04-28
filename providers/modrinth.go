package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"archive/zip"
	"strings"
	"path"
)

const MODRINTH_MANIFEST_FILE = "modrinth.index.json"

func isModrinthModpack(projectType string) bool {
	return projectType == "modpack"
}

type modrinthFile struct {
	Path string `json:"path"`
	Downloads []string `json:"downloads"`
}

type modrinthManifest struct {
	Files []modrinthFile `json:"files"`
}

type Project struct {
	ID string `json:"id"`
	ProjectType string `json:"project_type"`
}

type ModVersion struct {
    VersionNumber string   `json:"version_number"`
    GameVersions  []string `json:"game_versions"`
    Loaders       []string `json:"loaders"`
    Files         []struct {
        URL string `json:"url"`
		Filename string `json:"filename"`
    } `json:"files"`
    DatePublished string `json:"date_published"`
}

type ModrinthProvider struct{}

func (p *ModrinthProvider) Fetch(slug, mcVersion, loader string) (mod *Downloadable, isModpack bool, err error) {
	searchURL := fmt.Sprintf("https://api.modrinth.com/v2/project/%s", slug)
	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("modrinth project not found: %s", slug)
	}

	var project Project
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, false, err
	}

	if project.ProjectType != "mod" && project.ProjectType != "modpack" {
		return nil, false, fmt.Errorf("unsupported Modrinth project type: %s", project.ProjectType)
	}

	vURL := fmt.Sprintf("https://api.modrinth.com/v2/project/%s/version", project.ID)
	respV, err := http.Get(vURL)
	if err != nil {
		return nil, false, err
	}
	defer respV.Body.Close()

	if respV.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("failed to fetch versions for project: %s", slug)
	}

	var versions []ModVersion

	if err := json.NewDecoder(respV.Body).Decode(&versions); err != nil {
		return nil, false, err
	}

	var latestVersion *ModVersion

	for i := range versions {
		v := &versions[i]

		mcOk := false
		for _, gv := range v.GameVersions {
			if gv == mcVersion {
				mcOk = true
				break
			}
		}
		if !mcOk {
			continue
		}

		loaderOk := false
		for _, l := range v.Loaders {
			if l == loader {
				loaderOk = true
				break
			}
		}
		if !loaderOk {
			continue
		}

		if latestVersion == nil || v.DatePublished > latestVersion.DatePublished {
			latestVersion = v
		}
	}

	if latestVersion == nil || len(latestVersion.Files) == 0 {
		return nil, false, fmt.Errorf("no compatible version found for loader %s and mc version %s", loader, mcVersion)
	}

	mod = &Downloadable{
		URL: latestVersion.Files[0].URL,
		Filename: latestVersion.Files[0].Filename,
	}

	return mod, isModrinthModpack(project.ProjectType), nil
}

func (p *ModrinthProvider) FetchModpack(pack *Downloadable) ([]*Downloadable, error) {
	if err := downloadFile(pack.Filename, pack.URL); err != nil {
		return nil, err
	}
	defer os.Remove(pack.Filename)

	r, err := zip.OpenReader(pack.Filename)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == MODRINTH_MANIFEST_FILE {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			var manifest modrinthManifest
			if err := json.NewDecoder(rc).Decode(&manifest); err != nil {
				return nil, err
			}

			var mods []*Downloadable
			for _, file := range manifest.Files {
				if len(file.Downloads) == 0 {
					continue
				}

				if !strings.HasPrefix(file.Path, "mods/") {
					continue
				}

				filename := path.Base(file.Path)

				if filename == "." || filename == "/" {
					continue
				}

				mods = append(mods, &Downloadable{
					URL:      file.Downloads[0],
					Filename: filename,
				})
			}

			return mods, nil
		}
	}

	return nil, fmt.Errorf("%s not found", MODRINTH_MANIFEST_FILE)
}
