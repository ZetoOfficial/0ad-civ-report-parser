package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var spaFS embed.FS

func NewServer(repRoot string) http.Handler {
	h := &handlers{repRoot: repRoot}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/replays", h.listReplays)
	mux.HandleFunc("/api/replays/", h.getReplay)

	dist, err := fs.Sub(spaFS, "dist")
	if err != nil {
		panic("webui: embed dist subtree: " + err.Error())
	}
	staticFS := http.FileServer(http.FS(dist))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/" {
			serveIndex(w, dist)
			return
		}
		clean := strings.TrimPrefix(r.URL.Path, "/")
		if _, err := fs.Stat(dist, clean); err == nil {
			staticFS.ServeHTTP(w, r)
			return
		}
		serveIndex(w, dist)
	})

	return mux
}

func serveIndex(w http.ResponseWriter, dist fs.FS) {
	raw, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		http.Error(w, "frontend not built — run `make web-build`", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(raw)
}
