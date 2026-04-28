package loaders

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"github.com/salatine/modmcctl/internal/cli"
)

func getFabricLoaderVersionForMC(mcVersion string) (string, error) {
	resp, err := http.Get("https://meta.fabricmc.net/v2/versions/loader/" + mcVersion)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data []struct {
		Loader struct {
			Version     string `json:"version"`
			Stable      bool   `json:"stable"`
			Recommended bool   `json:"recommended"`
		} `json:"loader"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", fmt.Errorf("no Fabric loader found for Minecraft %s", mcVersion)
	}

	for _, d := range data {
		if d.Loader.Recommended {
			return d.Loader.Version, nil
		}
	}

	for _, d := range data {
		if d.Loader.Stable {
			return d.Loader.Version, nil
		}
	}

	return data[0].Loader.Version, nil
}

func installFabric(mode, clientDir, serverDir, mcVersion string) error {
	resp, err := http.Get("https://meta.fabricmc.net/v2/versions/installer")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	type Installer struct {
		Version string `json:"version"`
		URL     string `json:"url"`
	}

	var installers []Installer
	if err := json.NewDecoder(resp.Body).Decode(&installers); err != nil {
		return err
	}
	if len(installers) == 0 {
		return fmt.Errorf("no Fabric installer found")
	}

	var latestInstaller Installer
	for _, inst := range installers {
		if latestInstaller.Version == "" || inst.Version > latestInstaller.Version {
			latestInstaller = inst
		}
	}

	installerURL := latestInstaller.URL

	install := func(dir string, installType string) error {
		fmt.Printf("Installing Fabric (%s)... \n", installType)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		installer := filepath.Join(dir, "fabric-installer.jar")
		if err := cli.DownloadFile(installerURL, installer); err != nil {
			return err
		}
		defer os.Remove(installer)

		loaderVersion, err := getFabricLoaderVersionForMC(mcVersion)
		if err != nil {
			return err
		}

		args := []string{"-jar", installer, "client", "-dir", dir, "-mcversion", mcVersion, "-loader", loaderVersion}
		if installType == "server" {
			args = []string{"-jar", installer, "server", "-dir", dir, "-mcversion", mcVersion, "-loader", loaderVersion, "-downloadMinecraft"}
		}

		cmd := exec.Command("java", args...)
		cmd.Dir = dir
		cmd.Stdout = nil

		if err := cmd.Run(); err != nil {
			return err
		}
		return nil
	}

	switch mode {
	case "client":
		return install(clientDir, "client")
	case "server":
		return install(serverDir, "server")
	case "both":
		if err := install(clientDir, "client"); err != nil {
			return err
		}
		return install(serverDir, "server")
	default:
		return fmt.Errorf("invalid mode: %s", mode)
	}
}

