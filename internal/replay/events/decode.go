package events

import (
	"encoding/json"
)

// Decode turns a single `cmd <P> {json}` line into a typed Event.
// player is 1-based; tMs is the cumulative ms from game start.
func Decode(player int, tMs int64, raw json.RawMessage) Event {
	var head struct {
		Type string `json:"type"`
	}
	_ = json.Unmarshal(raw, &head) // malformed JSON would have failed earlier
	ev := Event{TMs: tMs, Player: player, Type: head.Type}

	switch head.Type {
	case "train":
		var d struct {
			Template string `json:"template"`
			Count    int    `json:"count"`
			Entities []int  `json:"entities"`
			Queued   bool   `json:"queued"`
		}
		_ = json.Unmarshal(raw, &d)
		src := 0
		if len(d.Entities) > 0 {
			src = d.Entities[0]
		}
		ev.Data = TrainData{Template: d.Template, Count: d.Count, Source: src, Queued: d.Queued}
	case "construct":
		var d struct {
			Template string  `json:"template"`
			X        float64 `json:"x"`
			Z        float64 `json:"z"`
			Entities []int   `json:"entities"`
		}
		_ = json.Unmarshal(raw, &d)
		ev.Data = ConstructData{Template: d.Template, X: d.X, Z: d.Z, Builders: len(d.Entities)}
	case "research":
		var d struct {
			Template string `json:"template"`
			Entity   int    `json:"entity"`
		}
		_ = json.Unmarshal(raw, &d)
		ev.Data = ResearchData{Template: d.Template, Source: d.Entity}
	case "attack":
		var d struct {
			Target       int   `json:"target"`
			Entities     []int `json:"entities"`
			AllowCapture bool  `json:"allowCapture"`
		}
		_ = json.Unmarshal(raw, &d)
		ev.Data = AttackData{Target: d.Target, Entities: len(d.Entities), AllowCapture: d.AllowCapture}
	case "garrison":
		var d struct {
			Target   int   `json:"target"`
			Entities []int `json:"entities"`
		}
		_ = json.Unmarshal(raw, &d)
		ev.Data = GarrisonData{Target: d.Target, Entities: len(d.Entities)}
	case "tribute":
		var d struct {
			Player  int            `json:"player"`
			Amounts map[string]int `json:"amounts"`
		}
		_ = json.Unmarshal(raw, &d)
		ev.Data = TributeData{
			To:    d.Player,
			Food:  d.Amounts["food"],
			Wood:  d.Amounts["wood"],
			Stone: d.Amounts["stone"],
			Metal: d.Amounts["metal"],
		}
	case "resign":
		ev.Data = ResignData{}
	case "chat":
		var d struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(raw, &d)
		ev.Data = ChatData{Text: d.Message}
	default:
		ev.Type = TypeRaw
		ev.Data = RawData{Kind: head.Type, JSON: raw}
	}
	return ev
}
