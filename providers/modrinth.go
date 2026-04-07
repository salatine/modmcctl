package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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

func (p *ModrinthProvider) FetchMod(slug, mcVersion, loader string) (string, string, error) {
	searchURL := fmt.Sprintf("https://api.modrinth.com/v2/project/%s", slug)
	resp, err := http.Get(searchURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("modrinth project not found: %s", slug)
	}

	var project struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return "", "", err
	}

	vURL := fmt.Sprintf("https://api.modrinth.com/v2/project/%s/version", project.ID)
	respV, err := http.Get(vURL)
	if err != nil {
		return "", "", err
	}
	defer respV.Body.Close()

	if respV.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to fetch versions for project: %s", slug)
	}

	var versions []ModVersion

	if err := json.NewDecoder(respV.Body).Decode(&versions); err != nil {
		return "", "", err
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
		return "", "", fmt.Errorf("no compatible version found for loader %s and mc version %s", loader, mcVersion)
	}

	return latestVersion.Files[0].URL, latestVersion.Files[0].Filename, nil
}
