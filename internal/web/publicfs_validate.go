package web

import (
	"fmt"
	"io/fs"
)

func requireIndexFile(filesystem fs.FS) error {
	f, err := filesystem.Open("index.html")
	if err != nil {
		return fmt.Errorf("index.html not found in web assets: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat index.html: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("index.html is a directory")
	}

	return nil
}
