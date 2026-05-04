package cli

import (
	"flag"
	"strings"
)

type Config struct {
	Mode        string
	ClientDir   string
	ServerDir   string
	Version     string
	Loader      string
	Provider    string
	Mods        []string
}

func ParseConfig() *Config {
	mode, _ := Flag("mode", "client", "client|server|both", Contains("not supported mode", "client", "server", "both"))
	clientDirFlag, _ := Flag("client-dir", "", ".minecraft client absolute path", nil)
	serverDirFlag, _ := Flag("server-dir", "", ".minecraft server absolute path", nil)
	version := Required(Flag("version", "", "minecraft version", nil))
	loader := Required(Flag("loader", "", "neoforge|fabric", Contains("not supported loader", "neoforge", "fabric")))
	providerName, _ := Flag("provider", "modrinth", "modrinth|curseforge", Contains("not supported provider", "modrinth", "curseforge"))
	modsFlag := Required(Flag("mods", "", "list of mod/modpack slugs or names, separated by comma", nil))

	flag.Parse()
	Validate()

	return &Config{
		Mode:      *mode,
		ClientDir: GetClientDir(*clientDirFlag),
		ServerDir: GetServerDir(*serverDirFlag),
		Version:   *version,
		Loader:    *loader,
		Provider:  *providerName,
		Mods:      strings.Split(*modsFlag, ","),
	}
}
