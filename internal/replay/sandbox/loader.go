// Package sandbox reads per-player time-series sequences from a 0 A.D.
// metadata.json. Originally written to read from a headless-replay sandbox
// (`pyrogenesis -replay= -userdir=`), but in practice the user's own
// metadata.json already contains sequences whenever the summary screen was
// viewed at end-of-game — so LoadFromMeta on the source replay's own
// metadata.json is the primary path. The sandbox path stays available as a
// fallback for replays the user closed before the summary screen rendered.
package sandbox

import (
	"encoding/json"
	"os"
	"path/filepath"

	api "github.com/ZetoOfficial/0ad-civ-report-parser/internal/api/gen"
)

// EnvRoot lets callers point Load at a custom sandbox via env var.
const EnvRoot = "OAD_REPLAY_SANDBOX"

// fallbackRoot is the path used when EnvRoot is unset. Relative to CWD;
// replayreport is normally invoked from the repo root so this lands in
// `<repo>/tmp/0ad-replay-sandbox` (gitignored).
const fallbackRoot = "tmp/0ad-replay-sandbox"

// DefaultRoot returns the sandbox root: $OAD_REPLAY_SANDBOX if set, else
// `tmp/0ad-replay-sandbox` resolved against the current working directory.
func DefaultRoot() string {
	if v := os.Getenv(EnvRoot); v != "" {
		return v
	}
	return fallbackRoot
}

// MetadataPath returns the absolute path where the sandbox stores the
// regenerated metadata.json for a given replay basename.
func MetadataPath(root, replayBasename string) string {
	return filepath.Join(root, "replays", "0.28.0", replayBasename, "metadata.json")
}

// Load reads sequences for each player from the sandbox metadata.json that
// corresponds to replayBasename (e.g. "2026-05-28_0001"). Returns map keyed by
// 1-based player ID. Returns nil (no error) if the file is missing or empty —
// callers treat this as "no sequences yet for this replay".
//
// Note: schema produced by pyrogenesis 0.28 uses camelCase field names; we
// unmarshal into a private mirror type and convert to our snake_case output
// types.
func Load(replayBasename string) (map[int]*api.Sequences, error) {
	return LoadFromRoot(DefaultRoot(), replayBasename)
}

// LoadFromRoot is Load with a custom sandbox root (test seam).
func LoadFromRoot(root, replayBasename string) (map[int]*api.Sequences, error) {
	return LoadFromMeta(MetadataPath(root, replayBasename))
}

// LoadFromMeta parses sequences from any metadata.json path. Missing or empty
// file → nil, nil (caller treats as "no sequences").
func LoadFromMeta(path string) (map[int]*api.Sequences, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}
	return Parse(raw)
}

// Parse decodes the metadata.json bytes and returns per-player sequences.
// Skips player 0 (gaia) and any player whose sequences key is missing.
func Parse(raw []byte) (map[int]*api.Sequences, error) {
	var doc struct {
		PlayerStates []rawPlayer `json:"playerStates"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	out := map[int]*api.Sequences{}
	for i, ps := range doc.PlayerStates {
		if i == 0 {
			continue // gaia
		}
		if ps.Sequences == nil {
			continue
		}
		s := convert(ps.Sequences)
		if s == nil {
			continue
		}
		out[i] = s
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

type rawPlayer struct {
	Sequences *rawSequences `json:"sequences"`
}

type rawSequences struct {
	Time                      []float64           `json:"time"`
	PopulationCount           []int               `json:"populationCount"`
	PercentMapExplored        []int               `json:"percentMapExplored"`
	PercentMapControlled      []int               `json:"percentMapControlled"`
	UnitsTrained              map[string][]int    `json:"unitsTrained"`
	UnitsLost                 map[string][]int    `json:"unitsLost"`
	EnemyUnitsKilled          map[string][]int    `json:"enemyUnitsKilled"`
	UnitsCaptured             map[string][]int    `json:"unitsCaptured"`
	BuildingsConstructed      map[string][]int    `json:"buildingsConstructed"`
	BuildingsLost             map[string][]int    `json:"buildingsLost"`
	EnemyBuildingsDestroyed   map[string][]int    `json:"enemyBuildingsDestroyed"`
	BuildingsCaptured         map[string][]int    `json:"buildingsCaptured"`
	UnitsLostValue            []float64           `json:"unitsLostValue"`
	EnemyUnitsKilledValue     []float64           `json:"enemyUnitsKilledValue"`
	UnitsCapturedValue        []float64           `json:"unitsCapturedValue"`
	BuildingsLostValue        []float64           `json:"buildingsLostValue"`
	EnemyBuildingsDestroyVal  []float64           `json:"enemyBuildingsDestroyedValue"`
	BuildingsCapturedValue    []float64           `json:"buildingsCapturedValue"`
	TradeIncome               []float64           `json:"tradeIncome"`
	TributesSent              []float64           `json:"tributesSent"`
	TributesReceived          []float64           `json:"tributesReceived"`
	LootCollected             []float64           `json:"lootCollected"`
	TreasuresCollected        []float64           `json:"treasuresCollected"`
	SuccessfulBribes          []int               `json:"successfulBribes"`
	FailedBribes              []int               `json:"failedBribes"`
	ResourcesCount            map[string][]float64 `json:"resourcesCount"`
	ResourcesGathered         map[string][]float64 `json:"resourcesGathered"`
	ResourcesUsed             map[string][]float64 `json:"resourcesUsed"`
	ResourcesBought           map[string][]float64 `json:"resourcesBought"`
	ResourcesSold             map[string][]float64 `json:"resourcesSold"`
}

func convert(r *rawSequences) *api.Sequences {
	if len(r.Time) == 0 {
		return nil
	}
	return &api.Sequences{
		Time:                         r.Time,
		PopulationCount:              r.PopulationCount,
		PercentMapExplored:           r.PercentMapExplored,
		PercentMapControlled:         r.PercentMapControlled,
		UnitsTrained:                 api.ClassSeriesMap(r.UnitsTrained),
		UnitsLost:                    api.ClassSeriesMap(r.UnitsLost),
		EnemyUnitsKilled:             api.ClassSeriesMap(r.EnemyUnitsKilled),
		UnitsCaptured:                api.ClassSeriesMap(r.UnitsCaptured),
		BuildingsConstructed:         api.ClassSeriesMap(r.BuildingsConstructed),
		BuildingsLost:                api.ClassSeriesMap(r.BuildingsLost),
		EnemyBuildingsDestroyed:      api.ClassSeriesMap(r.EnemyBuildingsDestroyed),
		BuildingsCaptured:            api.ClassSeriesMap(r.BuildingsCaptured),
		UnitsLostValue:               r.UnitsLostValue,
		EnemyUnitsKilledValue:        r.EnemyUnitsKilledValue,
		UnitsCapturedValue:           r.UnitsCapturedValue,
		BuildingsLostValue:           r.BuildingsLostValue,
		EnemyBuildingsDestroyedValue: r.EnemyBuildingsDestroyVal,
		BuildingsCapturedValue:       r.BuildingsCapturedValue,
		TradeIncome:                  r.TradeIncome,
		TributesSent:                 r.TributesSent,
		TributesReceived:             r.TributesReceived,
		LootCollected:                r.LootCollected,
		TreasuresCollected:           r.TreasuresCollected,
		SuccessfulBribes:             r.SuccessfulBribes,
		FailedBribes:                 r.FailedBribes,
		ResourcesCount:               api.ResourceSeriesMap(r.ResourcesCount),
		ResourcesGathered:            api.ResourceSeriesMap(r.ResourcesGathered),
		ResourcesUsed:                api.ResourceSeriesMap(r.ResourcesUsed),
		ResourcesBought:              api.ResourceSeriesMap(r.ResourcesBought),
		ResourcesSold:                api.ResourceSeriesMap(r.ResourcesSold),
	}
}
