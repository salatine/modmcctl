package installers

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"github.com/salatine/modmcctl/internal/cli"
)


func getNeoForgeVersionForMC(mcVersion string) (string, error) {
	resp, err := http.Get("https://maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var meta MavenMetadata
	if err := xml.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return "", err
	}

	prefix := ""
	parts := strings.Split(mcVersion, ".")
	if len(parts) >= 3 {
		prefix = parts[1] + "." + parts[2] + "."
	} else if len(parts) == 2 {
		prefix = parts[1] + ".0."
	}

	for i := len(meta.Versioning.Versions.Version) - 1; i >= 0; i-- {
		v := meta.Versioning.Versions.Version[i]
		if strings.HasPrefix(v, prefix) {
			return v, nil
		}
	}
	return meta.Versioning.Release, nil
}

func installNeoForge(mode, clientDir, serverDir, mcVersion string) error {
	version, err := getNeoForgeVersionForMC(mcVersion)
	if err != nil {
		return err
	}

	install := func(dir string, installType string) error {
		versionPath := filepath.Join(dir, "versions", "neoforge-"+version)
		if _, err := os.Stat(versionPath); err == nil {
			fmt.Printf("NeoForge %s already installed in %s, skipping...\n", version, installType)
			return nil
		}

		fmt.Printf("Installing NeoForge %s (%s)... ", version, installType)
		cli.EnsureDir(dir)
		installer := filepath.Join(dir, "neoforge-installer.jar")
		url := fmt.Sprintf("https://maven.neoforged.net/releases/net/neoforged/neoforge/%s/neoforge-%s-installer.jar", version, version)

		if err := cli.DownloadFile(url, installer); err != nil {
			return err
		}
		defer os.Remove(installer)

		args := []string{"-jar", installer}
		if installType == "client" {
			args = append(args, "--installClient")
		} else if installType == "server" {
			args = append(args, "--installServer")
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

