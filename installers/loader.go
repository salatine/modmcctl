package installers

import "strings"

func InstallLoader(loader, mode, clientDir, serverDir, mcVersion string) error {
	if strings.ToLower(loader) == "neoforge" {
		return installNeoForge(mode, clientDir, serverDir, mcVersion)
	} else {
		return installFabric(mode, clientDir, serverDir, mcVersion)
	}
}

