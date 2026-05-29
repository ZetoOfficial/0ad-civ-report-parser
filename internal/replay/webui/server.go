package webui

import (
	"net/http"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/techlib"
)

// NewServer returns the JSON REST API for replays. Static frontend assets are
// served separately (vite dev server in dev; vite preview / nginx / static
// host in prod) — the backend never touches HTML.
// lib may be nil; when non-nil, research events are resolved to human-readable names.
func NewServer(repRoot string, lib *techlib.Lib) http.Handler {
	h := &handlers{repRoot: repRoot, lib: lib}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/replays", h.listReplays)
	mux.HandleFunc("/api/replays/", h.getReplay)
	return mux
}
