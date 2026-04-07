package cli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func NormalizeMCVersion(v string) string {
	parts := strings.Split(v, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return v
}

func GetClientDir(custom string) string {
	if custom != "" {
		return custom
	}
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		return filepath.Join(home, "AppData", "Roaming", ".minecraft")
	}
	return filepath.Join(home, ".minecraft")
}

func GetServerDir(custom string) string {
	if custom != "" {
		return custom
	}
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		return filepath.Join(home, "AppData", "Roaming", ".minecraftserver")
	}
	return filepath.Join(home, ".minecraftserver")
}

func EnsureDir(path string) {
	os.MkdirAll(path, os.ModePerm)
}

func DownloadFile(url, path string) error {
	var lastErr error
	for i := 0; i < 3; i++ {
		resp, err := http.Get(url)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("status: %s", resp.Status)
			continue
		}

		out, err := os.Create(path)
		if err != nil {
			return err
		}
		defer out.Close()

		if _, err = io.Copy(out, resp.Body); err == nil {
			return nil
		}
		lastErr = err
		os.Remove(path)
	}
	return lastErr
}
