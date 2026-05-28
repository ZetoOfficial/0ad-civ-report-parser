// Command replayreport parses 0ad replays and writes analysis.json next to each.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay"
)

const defaultReplayDir = "/Users/zeto/Library/Application Support/0ad/replays/0.28.0"

func main() {
	var (
		all    bool
		check  bool
		repDir string
	)
	flag.BoolVar(&all, "all", false, "process every replay subdir under replay root")
	flag.BoolVar(&check, "check", false, "validate replays; exit with non-zero on any failure (no http)")
	flag.StringVar(&repDir, "replays", defaultReplayDir, "replay root (used when no positional arg)")
	flag.Parse()

	if flag.NArg() == 1 && !all {
		if err := runOne(flag.Arg(0)); err != nil {
			fmt.Fprintln(os.Stderr, "ERROR:", err)
			os.Exit(1)
		}
		return
	}

	root := repDir
	if flag.NArg() == 1 && all {
		root = flag.Arg(0)
	}
	runScan(root, check)
}

func runOne(dir string) error {
	a, err := replay.Run(dir)
	if err != nil {
		return err
	}
	fmt.Printf("OK %s — %s (%s, %d events)\n", a.Game.MatchID, a.Game.Map, dir, len(a.Events))
	return nil
}

func runScan(root string, strict bool) {
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
		if _, err := replay.Run(dir); err != nil {
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
