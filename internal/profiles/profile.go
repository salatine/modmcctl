package profiles

import (
	"os"
	"encoding/json"
	"path/filepath"
)

const CONFIG_DIR = "modmcctl"
const CONFIG_FILE = "profiles.json"

type Profile struct {
	Name string `json:"name"`
	Mode string `json:"mode"`
	ClientDir string `json:"client_dir"`
	ServerDir string `json:"server_dir"`
	Version string `json:"version"`
	Loader string `json:"loader"`
	Provider string `json:"provider"`
	Mods []string `json:"mods"`
}

type ProfilesFile struct {
	Profiles []Profile `json:"profiles"`
	Current string `json:"current"`
}


func configPath() string {
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, CONFIG_DIR, CONFIG_FILE)
}

func Load() (*ProfilesFile, error) {
	path := configPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &ProfilesFile{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var file ProfilesFile
	err = json.Unmarshal(data, &file)
	return &file, err
}

func Save(file *ProfilesFile) error {
	path := configPath()
	os.MkdirAll(filepath.Dir(path), 0755)

	data, err := json.MarshalIndent(file, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
