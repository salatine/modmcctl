package providers

import (
	"os"
	"fmt"
	"net/http"
	"encoding/json"
	"strings"
	"io"
	"archive/zip"
)

const CURSEFORGE_MANIFEST_FILE = "manifest.json"

func isCurseForgeModpack(classID float64) bool {
	return int(classID) == 4471
}

func downloadFile(path, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

type curseForgeFile struct {
	ProjectID int `json:"projectID"`
	FileID int `json:"fileID"`
}

type curseForgeManifest struct {
	Files []curseForgeFile `json:"files"`
}

type CurseForgeProvider struct{}

func (p *CurseForgeProvider) Fetch(slug, mcVersion, loader string) (mod *ModDownload, isModpack bool, err error) {
	apiKey := os.Getenv("CURSEFORGE_API_KEY")
	if apiKey == "" {
		return nil, false, fmt.Errorf("CURSEFORGE_API_KEY not set")
	}

	client := &http.Client{}
	url := fmt.Sprintf("https://api.curseforge.com/v1/mods/search?gameId=432&slug=%s", slug)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("x-api-key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	var search map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&search); err != nil {
		return nil, false, err
	}

	dataRaw, ok := search["data"].([]interface{})
	if !ok || len(dataRaw) == 0 {
		return nil, false, fmt.Errorf("mod not found: %s", slug)
	}
	
	var modData map[string]interface{}
	var classID float64
	for _, m := range dataRaw {
		mMap := m.(map[string]interface{})
		if classID, ok = mMap["classId"].(float64); ok {
			modData = mMap
			break
		}
	}

	if classID != 6 && classID != 4471 {
		return nil, false, fmt.Errorf("unsupported CurseForge class ID: %d", classID)
	}

	targetLoaderID := 0
	switch strings.ToLower(loader) {
	case "fabric":
		targetLoaderID = 4
	case "neoforge":
		targetLoaderID = 6
	}

	indexes, ok := modData["latestFilesIndexes"].([]interface{})
	if ok {
		for _, idx := range indexes {
			entry := idx.(map[string]interface{})
			gv, _ := entry["gameVersion"].(string)
			lID, _ := entry["modLoader"].(float64)

			if gv == mcVersion && (targetLoaderID == 0 || int(lID) == targetLoaderID) {
				fileID := int(entry["fileId"].(float64))
				filename := entry["filename"].(string)
				mod = &ModDownload{
					URL: p.buildCurseforgeURL(fileID, filename),
					Filename: filename,
				}

				return mod, isCurseForgeModpack(classID), nil
			}
		}
	}

	return nil, false, fmt.Errorf("no compatible version for %s", slug)
}

func (p *CurseForgeProvider) buildCurseforgeURL(fileID int, fileName string) string {
	idStr := fmt.Sprintf("%d", fileID)
	if len(idStr) < 4 {
		return ""
	}
	return fmt.Sprintf("https://edge.forgecdn.net/files/%s/%s/%s", idStr[:len(idStr)-3], idStr[len(idStr)-3:], fileName)
}

func (p *CurseForgeProvider) FetchModpack(pack *ModDownload) ([]*ModDownload, error) {
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
		if f.Name == CURSEFORGE_MANIFEST_FILE {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			var manifest curseForgeManifest
			if err := json.NewDecoder(rc).Decode(&manifest); err != nil {
				return nil, err
			}

			var mods []*ModDownload

			for _, file := range manifest.Files {
				url, filename, err := p.resolveCurseForgeFile(file.ProjectID, file.FileID)
				if err != nil {
					return nil, err
				}

				mods = append(mods, &ModDownload{
					URL:      url,
					Filename: filename,
				})
			}

			return mods, nil
		}
	}

	return nil, fmt.Errorf("%s not found", CURSEFORGE_MANIFEST_FILE)
}

func (p *CurseForgeProvider) resolveCurseForgeFile(projectID, fileID int) (string, string, error) {
	apiKey := os.Getenv("CURSEFORGE_API_KEY")
	if apiKey == "" {
		return "", "", fmt.Errorf("CURSEFORGE_API_KEY not set")
	}

	url := fmt.Sprintf("https://api.curseforge.com/v1/mods/%d/files/%d", projectID, fileID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("x-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var data struct {
		Data struct {
			FileName    string `json:"fileName"`
			DownloadURL string `json:"downloadUrl"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", err
	}

	if data.Data.DownloadURL == "" {
		url := p.buildCurseforgeURL(fileID, data.Data.FileName)
		if url == "" {
			return "", "", fmt.Errorf("no download URL")
		}

		return url, data.Data.FileName, nil
	}

	return data.Data.DownloadURL, data.Data.FileName, nil
}
