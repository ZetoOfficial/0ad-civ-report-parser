// Package events typed-decodes 0ad replay command JSON.
package events

import "encoding/json"

const (
	TypeTrain     = "train"
	TypeConstruct = "construct"
	TypeResearch  = "research"
	TypeAttack    = "attack"
	TypeGarrison  = "garrison"
	TypeTribute   = "tribute"
	TypeResign    = "resign"
	TypeChat      = "chat"
	TypeRaw       = "raw"
)

type Event struct {
	TMs    int64       // ms from game start
	Player int         // 1-based; 0 = gaia (won't occur)
	Type   string      // one of Type* constants
	Data   any // type-specific payload below
}

type TrainData struct {
	Template string `json:"template"`
	Count    int    `json:"count"`
	Source   int    `json:"source,omitempty"` // entity ID training; 0 if from entities[]
	Queued   bool   `json:"queued"`
}

type ConstructData struct {
	Template string  `json:"template"`
	X        float64 `json:"x"`
	Z        float64 `json:"z"`
	Builders int     `json:"builders"`
}

type ResearchData struct {
	Template string `json:"template"`
	Source   int    `json:"source"`
}

type AttackData struct {
	Target       int  `json:"target"`
	Entities     int  `json:"entities"`
	AllowCapture bool `json:"allow_capture"`
}

type GarrisonData struct {
	Target   int `json:"target"`
	Entities int `json:"entities"`
}

type TributeData struct {
	To    int `json:"to"`
	Food  int `json:"food"`
	Wood  int `json:"wood"`
	Stone int `json:"stone"`
	Metal int `json:"metal"`
}

type ResignData struct{}

type ChatData struct {
	Text string `json:"text"`
}

type RawData struct {
	Kind string          `json:"kind"`
	JSON json.RawMessage `json:"json"`
}
