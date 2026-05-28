package output

const SchemaVersion = 1

type Analysis struct {
	SchemaVersion int        `json:"schema_version"`
	Game          GameInfo   `json:"game"`
	Players       []Player   `json:"players"`
	Events        []Event    `json:"events"`
	Snapshots     []Snapshot `json:"snapshots"` // empty in v1; reserved for forward compat
	FinalState    FinalState `json:"final_state"`
	Metrics       Metrics    `json:"metrics"`
}

type GameInfo struct {
	MatchID       string   `json:"match_id"`
	Map           string   `json:"map"`
	MapType       string   `json:"map_type"`
	Timestamp     int64    `json:"timestamp"`          // unix sec
	DurationMs    int64    `json:"duration_ms"`
	EngineVersion string   `json:"engine_version"`
	VictoryConds  []string `json:"victory_conditions"`
}

type Player struct {
	ID     int    `json:"id"`              // 1-based; matches cmd <P>
	Name   string `json:"name"`
	Civ    string `json:"civ"`
	Team   int    `json:"team"`
	IsAI   bool   `json:"is_ai"`
	AIDiff int    `json:"ai_diff,omitempty"`
	Color  Color  `json:"color"`
}

type Color struct{ R, G, B int }

type Event struct {
	T      int64  `json:"t"`               // ms from game start
	Player int    `json:"player"`          // 1-based
	Type   string `json:"type"`
	Data   any    `json:"data,omitempty"`
}

type Snapshot struct{} // placeholder; not populated in v1

type FinalState struct {
	Players map[int]PlayerFinalState `json:"players"`
}

type PlayerFinalState struct {
	Phase           string    `json:"phase"`
	State           string    `json:"state"`   // "active" | "won" | "defeated"
	Outcome         string    `json:"outcome"` // resolved (resign event takes precedence)
	PopCount        int       `json:"pop_count"`
	PopLimit        int       `json:"pop_limit"`
	PopMax          int       `json:"pop_max"`
	ResourceCounts  Resources `json:"resource_counts"`
	ResearchedTechs []string  `json:"researched_techs"`
}

type Resources struct {
	Food, Wood, Stone, Metal int
}

type Metrics struct {
	Players map[int]PlayerMetrics `json:"players"`
	Density []DensityBin          `json:"action_density"`
}

type PlayerMetrics struct {
	PhaseTimings map[string]int `json:"phase_timings"` // sec
	Engagements  []Engagement   `json:"engagements"`
	Anomalies    []Anomaly      `json:"anomalies"`
}

type Engagement struct {
	TStartSec    int `json:"t_start_sec"`
	TEndSec      int `json:"t_end_sec"`
	Target       int `json:"target"`
	PeakUnits    int `json:"peak_units"`
	CommandCount int `json:"command_count"`
}

type Anomaly struct {
	Type      string         `json:"type"`
	TStartSec int            `json:"t_start_sec"`
	TEndSec   int            `json:"t_end_sec"`
	Severity  string         `json:"severity"`
	Details   map[string]any `json:"details,omitempty"`
}

type DensityBin struct {
	TSec   int            `json:"t_sec"`    // bin start
	Counts map[string]int `json:"counts"`   // category → count (per all players)
}
