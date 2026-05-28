// Package commands streams lines from a replay commands.txt file.
package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type LineKind int

const (
	KindStart LineKind = iota
	KindTurn
	KindCmd
	KindHash
	KindHashQuick
	KindEnd
)

type Line struct {
	Kind      LineKind
	StartJSON json.RawMessage // KindStart
	TurnN     int             // KindTurn
	TickMs    int             // KindTurn
	Player    int             // KindCmd (1-based)
	CmdJSON   json.RawMessage // KindCmd
	Hash      string          // KindHash, KindHashQuick
}

type Reader struct {
	sc *bufio.Scanner
}

func New(r io.Reader) *Reader {
	sc := bufio.NewScanner(r)
	// real replays have very long start lines (>64KB default buffer)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	return &Reader{sc: sc}
}

// Next reads the next line. Returns (Line, true, nil) on success,
// (Line{}, false, nil) on EOF, (Line{}, false, err) on parse error.
func (r *Reader) Next() (Line, bool, error) {
	if !r.sc.Scan() {
		if err := r.sc.Err(); err != nil {
			return Line{}, false, err
		}
		return Line{}, false, nil
	}
	raw := r.sc.Text()
	if raw == "" {
		return r.Next()
	}
	switch {
	case raw == "end":
		return Line{Kind: KindEnd}, true, nil
	case strings.HasPrefix(raw, "start "):
		return Line{Kind: KindStart, StartJSON: json.RawMessage(raw[len("start "):])}, true, nil
	case strings.HasPrefix(raw, "turn "):
		var n, ms int
		if _, err := fmt.Sscanf(raw, "turn %d %d", &n, &ms); err != nil {
			return Line{}, false, fmt.Errorf("commands: bad turn line %q: %w", raw, err)
		}
		return Line{Kind: KindTurn, TurnN: n, TickMs: ms}, true, nil
	case strings.HasPrefix(raw, "cmd "):
		sp1 := strings.IndexByte(raw[4:], ' ')
		if sp1 < 0 {
			return Line{}, false, fmt.Errorf("commands: bad cmd line %q", raw)
		}
		p, err := strconv.Atoi(raw[4 : 4+sp1])
		if err != nil {
			return Line{}, false, fmt.Errorf("commands: bad cmd player %q: %w", raw, err)
		}
		return Line{Kind: KindCmd, Player: p, CmdJSON: json.RawMessage(raw[4+sp1+1:])}, true, nil
	case strings.HasPrefix(raw, "hash-quick "):
		return Line{Kind: KindHashQuick, Hash: raw[len("hash-quick "):]}, true, nil
	case strings.HasPrefix(raw, "hash "):
		return Line{Kind: KindHash, Hash: raw[len("hash "):]}, true, nil
	default:
		return Line{}, false, fmt.Errorf("commands: unknown line %q", raw)
	}
}
