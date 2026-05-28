package webui

import (
	"net/http"
)

func NewServer(repRoot string) http.Handler {
	h := &handlers{repRoot: repRoot}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			h.index(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/replay/", h.replay)
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))
	return mux
}
