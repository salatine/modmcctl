package providers

import (
	"fmt"
	"path/filepath"
	"io"
	"os"
	"strings"
	"archive/zip"
)

type Downloadable struct {
	URL string
	Filename string
}

type ModProvider interface {
	Fetch(slug, mcVersion, loader string) (mod *Downloadable, isModpack bool, err error)
	FetchModpack(pack *Downloadable, destDir string) (mods []*Downloadable, err error)
}

func safeJoin(base, target string) (string, error) {
	destPath := filepath.Join(base, target)

	if !strings.HasPrefix(destPath, filepath.Clean(base)+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid path: %s", target)
	}

	return destPath, nil
}

func extractOverridesFromZip(r *zip.ReadCloser, destDir string) error {
	const prefix = "overrides/"

	for _, f := range r.File {
		if !strings.HasPrefix(f.Name, prefix) {
			continue
		}

		relPath := strings.TrimPrefix(f.Name, prefix)
		if relPath == "" {
			continue
		}

		destPath, err := safeJoin(destDir, relPath)
		if err != nil {
			return err
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		src, err := f.Open()
		if err != nil {
			return err
		}

		dst, err := os.Create(destPath)
		if err != nil {
			src.Close()
			return err
		}
		
		fmt.Printf("overrides: %s\n", destPath)
		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			dst.Close()
			return err
		}

		src.Close()
		dst.Close()
	}

	return nil
}
