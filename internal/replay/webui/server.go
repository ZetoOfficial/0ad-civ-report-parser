package webui

import "net/http"

// NewServer returns the JSON REST API for replays. Static frontend assets are
// served separately (vite dev server in dev; vite preview / nginx / static
// host in prod) — the backend never touches HTML.
func NewServer(repRoot string) http.Handler {
	h := &handlers{repRoot: repRoot}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/replays", h.listReplays)
	mux.HandleFunc("/api/replays/", h.getReplay)
	return mux
}
