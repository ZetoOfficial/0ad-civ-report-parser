package output

const SchemaVersion = 4

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
	Timestamp     int64    `json:"timestamp"` // unix sec
	DurationMs    int64    `json:"duration_ms"`
	EngineVersion string   `json:"engine_version"`
	VictoryConds  []string `json:"victory_conditions"`
}

type Player struct {
	ID     int    `json:"id"` // 1-based; matches cmd <P>
	Name   string `json:"name"`
	Civ    string `json:"civ"`
	Team   int    `json:"team"`
	IsAI   bool   `json:"is_ai"`
	AIDiff int    `json:"ai_diff,omitempty"`
	Color  Color  `json:"color"`
}

type Color struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

type Event struct {
	T      int64  `json:"t"`      // ms from game start
	Player int    `json:"player"` // 1-based
	Type   string `json:"type"`
	Data   any    `json:"data,omitempty"`
}

type Snapshot struct{} // placeholder; not populated

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
	Food  int `json:"food"`
	Wood  int `json:"wood"`
	Stone int `json:"stone"`
	Metal int `json:"metal"`
}

type Metrics struct {
	Players map[int]PlayerMetrics `json:"players"`
}

type PlayerMetrics struct {
	PhaseTimings map[string]int `json:"phase_timings"` // sec
	Engagements  []Engagement   `json:"engagements"`
	Anomalies    []Anomaly      `json:"anomalies"`
	Sequences    *Sequences     `json:"sequences,omitempty"` // populated from sandbox-regen metadata.json
}

// Sequences is the time-series data StatisticsTracker writes when a replay is
// replayed headless via `pyrogenesis -replay=`. Arrays are aligned to Time[]
// (seconds from game start, ~30s sampling).
type Sequences struct {
	Time []float64 `json:"time"` // x-axis, seconds

	PopulationCount      []int `json:"population_count"`
	PercentMapExplored   []int `json:"percent_map_explored"`
	PercentMapControlled []int `json:"percent_map_controlled"`

	// Per-class cumulative counts. Class keys: Unit, Worker, Civilian, Cavalry,
	// Champion, Hero, Infantry, Ship, Siege, Trader, Domestic.
	UnitsTrained          ClassSeries `json:"units_trained"`
	UnitsLost             ClassSeries `json:"units_lost"`
	EnemyUnitsKilled      ClassSeries `json:"enemy_units_killed"`
	UnitsCaptured         ClassSeries `json:"units_captured"`
	BuildingsConstructed  ClassSeries `json:"buildings_constructed"`
	BuildingsLost         ClassSeries `json:"buildings_lost"`
	EnemyBuildingsDestroy ClassSeries `json:"enemy_buildings_destroyed"`
	BuildingsCaptured     ClassSeries `json:"buildings_captured"`

	// Resource-cost / income fields are stored as float because 0 A.D.
	// accumulates fractional resources (gather rates).
	UnitsLostValue           []float64 `json:"units_lost_value"`
	EnemyUnitsKilledValue    []float64 `json:"enemy_units_killed_value"`
	UnitsCapturedValue       []float64 `json:"units_captured_value"`
	BuildingsLostValue       []float64 `json:"buildings_lost_value"`
	EnemyBuildingsDestroyVal []float64 `json:"enemy_buildings_destroyed_value"`
	BuildingsCapturedValue   []float64 `json:"buildings_captured_value"`

	TradeIncome        []float64 `json:"trade_income"`
	TributesSent       []float64 `json:"tributes_sent"`
	TributesReceived   []float64 `json:"tributes_received"`
	LootCollected      []float64 `json:"loot_collected"`
	TreasuresCollected []float64 `json:"treasures_collected"`
	SuccessfulBribes   []int     `json:"successful_bribes"`
	FailedBribes       []int     `json:"failed_bribes"`
}

// ClassSeries: per-class cumulative time-series. Each array len == Sequences.Time len.
type ClassSeries map[string][]int

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
