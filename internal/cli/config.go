package cli

import (
	"flag"
	"strings"
	"github.com/salatine/modmcctl/internal/profiles"
)

type Config struct {
	ProfileName        string
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

	profileUse, _ := Flag("profile-use", "", "use an existing profile", nil)
	profileSave, _ := Flag("profile-save", "", "create or update a profile", nil)

	flag.Parse()

	pf, err := profiles.Load()
	if err != nil {
		panic(err)
	}
	
	if *profileUse != "" {
		for _, p := range pf.Profiles {
			if p.Name == *profileUse {
				pf.Current = p.Name
				profiles.Save(pf)

				return fromProfile(p)
			}
		}

		panic("profile not found: " + *profileUse)
	}

	Validate()

	cfg := &Config{
		Mode:      *mode,
		ClientDir: GetClientDir(*clientDirFlag),
		ServerDir: GetServerDir(*serverDirFlag),
		Version:   *version,
		Loader:    *loader,
		Provider:  *providerName,
		Mods:      strings.Split(*modsFlag, ","),
	}

	if *profileSave != "" {
		found := false
		cfg.ProfileName = *profileSave

		for i, _ := range pf.Profiles {
			if pf.Profiles[i].Name == *profileSave {
				pf.Profiles[i] = profiles.Profile{
					Name: *profileSave,
					Mode: cfg.Mode,
					ClientDir: cfg.ClientDir,
					ServerDir: cfg.ServerDir,
					Version: cfg.Version,
					Loader: cfg.Loader,
					Provider: cfg.Provider,
					Mods: cfg.Mods,
				}

				found = true
				break
			}
		}

		if !found {
			pf.Profiles = append(pf.Profiles, profiles.Profile{
					Name: *profileSave,
					Mode: cfg.Mode,
					ClientDir: cfg.ClientDir,
					ServerDir: cfg.ServerDir,
					Version: cfg.Version,
					Loader: cfg.Loader,
					Provider: cfg.Provider,
					Mods: cfg.Mods,
			})
		}

		pf.Current = *profileSave
		profiles.Save(pf)
	}

	if len(cfg.Mods) == 0 && pf.Current != "" {
		for _, p := range pf.Profiles {
			if p.Name == pf.Current {
				return fromProfile(p)
			}
		}
	}

	return cfg
}

func fromProfile(p profiles.Profile) *Config {
	return &Config{
		ProfileName: p.Name,
		Mode: p.Mode,
		ClientDir: p.ClientDir,
		ServerDir: p.ServerDir,
		Version: p.Version,
		Loader: p.Loader,
		Provider: p.Provider,
		Mods: p.Mods,
	}
}

