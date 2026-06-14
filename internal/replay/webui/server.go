package webui

import (
	"net/http"

	api "github.com/ZetoOfficial/0ad-civ-report-parser/internal/api/gen"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/techlib"
)

// NewServer returns the JSON REST API for replays. Static frontend assets are
// served separately (vite dev server in dev; vite preview / nginx / static
// host in prod) — the backend never touches HTML.
// lib may be nil; when nil, research events are resolved without human-readable metadata.
func NewServer(repRoot string, lib *techlib.Lib) http.Handler {
	h := &handlers{repRoot: repRoot, lib: lib}
	return api.Handler(h)
}
