package web

import (
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// NewHandler serves API traffic via apiHandler and everything else via the SPA filesystem.
func NewHandler(apiHandler http.Handler, publicFS fs.FS) http.Handler {
	static := http.FileServer(http.FS(publicFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api" {
			apiHandler.ServeHTTP(w, r)
			return
		}

		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		filePath := normalizePath(r.URL.Path)
		if fileExists(publicFS, filePath) {
			static.ServeHTTP(w, r)
			return
		}

		if !fileExists(publicFS, "index.html") {
			http.NotFound(w, r)
			return
		}

		indexReq := r.Clone(r.Context())
		indexReq.URL = cloneURLWithPath(r.URL, "/index.html")
		static.ServeHTTP(w, indexReq)
	})
}

func normalizePath(requestPath string) string {
	cleaned := strings.TrimPrefix(path.Clean("/"+requestPath), "/")
	if cleaned == "" || cleaned == "." {
		return "index.html"
	}
	return cleaned
}

func fileExists(filesystem fs.FS, name string) bool {
	f, err := filesystem.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return false
	}

	return !stat.IsDir()
}

func cloneURLWithPath(original *url.URL, newPath string) *url.URL {
	if original == nil {
		return &url.URL{Path: newPath}
	}

	clone := *original
	clone.Path = newPath
	clone.RawPath = ""
	return &clone
}
