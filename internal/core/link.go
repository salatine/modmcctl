package core

import (
	"os"
	"path/filepath"
	"runtime"
	"io"
	"github.com/salatine/modmcctl/internal/cli"
)

func linkProfileDirs(cfg *cli.Config, profileName string) error {
	base, err := getProfilesBaseDir()
	if err != nil {
		return err
	}

	base = filepath.Join(base, "profiles", profileName)

	if cfg.Mode == "client" || cfg.Mode == "both" {
		src := filepath.Join(base, "client", "mods")
		dst := filepath.Join(cfg.ClientDir, "mods")

		if err := relink(src, dst); err != nil {
			return err
		}
	}

	if cfg.Mode == "server" || cfg.Mode == "both" {
		src := filepath.Join(base, "server", "mods")
		dst := filepath.Join(cfg.ServerDir, "mods")

		if err := relink(src, dst); err != nil {
			return err
		}
	}

	return nil
}

func getProfilesBaseDir() (string, error) {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", os.ErrNotExist
		}
		return filepath.Join(appData, "modmcctl"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "share", "modmcctl"), nil
}

func relink(src, dst string) error {
	if err := os.MkdirAll(src, 0755); err != nil {
		return err
	}

	os.RemoveAll(dst)

	err := os.Symlink(src, dst)
	if err == nil {
		return nil
	}

	return copyDir(src, dst)
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())

		if e.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
