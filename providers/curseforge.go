package providers

import (
	"os"
	"fmt"
	"net/http"
	"encoding/json"
	"strings"
)

type CurseForgeProvider struct{}

func (p *CurseForgeProvider) FetchMod(slug, mcVersion, loader string) (string, string, error) {
	apiKey := os.Getenv("CURSEFORGE_API_KEY")
	if apiKey == "" {
		return "", "", fmt.Errorf("CURSEFORGE_API_KEY not set")
	}

	client := &http.Client{}
	url := fmt.Sprintf("https://api.curseforge.com/v1/mods/search?gameId=432&slug=%s", slug)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("x-api-key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var search map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&search); err != nil {
		return "", "", err
	}

	dataRaw, ok := search["data"].([]interface{})
	if !ok || len(dataRaw) == 0 {
		return "", "", fmt.Errorf("mod not found: %s", slug)
	}

	var modData map[string]interface{}
	for _, m := range dataRaw {
		mMap := m.(map[string]interface{})
		if classID, ok := mMap["classId"].(float64); ok && classID == 6 {
			modData = mMap
			break
		}
	}
	if modData == nil {
		modData = dataRaw[0].(map[string]interface{})
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
				return buildCurseforgeURL(fileID, filename), filename, nil
			}
		}
	}

	return "", "", fmt.Errorf("no compatible version for %s", slug)
}

func buildCurseforgeURL(fileID int, fileName string) string {
	idStr := fmt.Sprintf("%d", fileID)
	if len(idStr) < 4 {
		return ""
	}
	return fmt.Sprintf("https://edge.forgecdn.net/files/%s/%s/%s", idStr[:len(idStr)-3], idStr[len(idStr)-3:], fileName)
}

