// Command replayreport parses 0ad replays and writes analysis.json next to each.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/techlib"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/webui"
)

const defaultReplayDir = "/Users/zeto/Library/Application Support/0ad/replays/0.28.0"

func main() {
	var (
		all      bool
		check    bool
		repDir   string
		addr     string
		gamedata string
	)
	flag.BoolVar(&all, "all", false, "process every replay subdir under replay root")
	flag.BoolVar(&check, "check", false, "validate replays; exit with non-zero on any failure (no http)")
	flag.StringVar(&repDir, "replays", defaultReplayDir, "replay root (used when no positional arg)")
	flag.StringVar(&addr, "addr", ":8080", "HTTP listen address")

	// Resolution order: CLI flag > env var > compiled default.
	gamedataDefault := paths.DefaultGameDataRoot
	if env := os.Getenv(paths.EnvGameDataRoot); env != "" {
		gamedataDefault = env
	}
	flag.StringVar(&gamedata, "gamedata", gamedataDefault, "path to 0ad mods/public data root")
	flag.Parse()

	// Load tech library. A missing/unreadable gamedata root is non-fatal; the
	// pipeline falls back to recording only the raw template name.
	lib, err := techlib.Load(gamedata)
	if err != nil {
		fmt.Fprintf(os.Stderr, "techlib: %v; continuing without tech metadata\n", err)
		lib = nil
	}

	if flag.NArg() == 1 && !all {
		if err := runOne(flag.Arg(0), lib); err != nil {
			fmt.Fprintln(os.Stderr, "ERROR:", err)
			os.Exit(1)
		}
		return
	}

	root := repDir
	if flag.NArg() == 1 && all {
		root = flag.Arg(0)
	}
	if check {
		runScan(root, true, lib)
		return
	}

	fmt.Printf("scanning %s …\n", root)
	runScan(root, false, lib)
	if err := http.ListenAndServe(addr, webui.NewServer(root, lib)); err != nil {
		fmt.Fprintln(os.Stderr, "http:", err)
		os.Exit(1)
	}
}

func runOne(dir string, lib *techlib.Lib) error {
	a, err := replay.Run(dir, lib)
	if err != nil {
		return err
	}
	fmt.Printf("OK %s — %s (%s, %d events)\n", a.Game.MatchID, a.Game.Map, dir, len(a.Events))
	return nil
}

func runScan(root string, strict bool, lib *techlib.Lib) {
	entries, err := os.ReadDir(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scan:", err)
		os.Exit(1)
	}
	ok, skipped, failed := 0, 0, 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, e.Name())
		if _, err := os.Stat(filepath.Join(dir, "metadata.json")); err != nil {
			skipped++
			continue
		}
		if _, err := replay.Run(dir, lib); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", e.Name(), err)
			failed++
			continue
		}
		ok++
	}
	fmt.Printf("scan: %d ok, %d skipped (no metadata), %d failed\n", ok, skipped, failed)
	if strict && failed > 0 {
		os.Exit(2)
	}
}
