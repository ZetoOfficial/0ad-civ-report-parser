package events

import (
	"encoding/json"
	"testing"
)

func TestDecodeTrain(t *testing.T) {
	raw := json.RawMessage(`{"type":"train","template":"units/spart/inf_spearman_b","count":5,"entities":[2481],"queued":false}`)
	ev := Decode(2, 1200, raw)
	if ev.Type != "train" || ev.Player != 2 || ev.TMs != 1200 {
		t.Fatalf("ev = %+v", ev)
	}
	d := ev.Data.(TrainData)
	if d.Template != "units/spart/inf_spearman_b" || d.Count != 5 || d.Source != 2481 {
		t.Errorf("TrainData = %+v", d)
	}
}

func TestDecodeResearchPhase(t *testing.T) {
	raw := json.RawMessage(`{"type":"research","entity":2417,"template":"phase_town_generic","pushFront":false}`)
	ev := Decode(2, 295000, raw)
	d := ev.Data.(ResearchData)
	if d.Template != "phase_town_generic" || d.Source != 2417 {
		t.Errorf("ResearchData = %+v", d)
	}
}

func TestDecodeAttack(t *testing.T) {
	raw := json.RawMessage(`{"type":"attack","entities":[1,2,3],"target":99,"allowCapture":true}`)
	ev := Decode(2, 0, raw)
	d := ev.Data.(AttackData)
	if d.Target != 99 || d.Entities != 3 || !d.AllowCapture {
		t.Errorf("AttackData = %+v", d)
	}
}

func TestDecodeUnknownIsRaw(t *testing.T) {
	raw := json.RawMessage(`{"type":"walk","entities":[1],"x":10,"z":20}`)
	ev := Decode(2, 0, raw)
	if ev.Type != TypeRaw {
		t.Fatalf("type = %q, want raw", ev.Type)
	}
	d := ev.Data.(RawData)
	if d.Kind != "walk" {
		t.Errorf("Kind = %q", d.Kind)
	}
}

func TestDecodeResign(t *testing.T) {
	raw := json.RawMessage(`{"type":"resign"}`)
	ev := Decode(2, 1700000, raw)
	if ev.Type != "resign" {
		t.Errorf("type = %q", ev.Type)
	}
	if _, ok := ev.Data.(ResignData); !ok {
		t.Errorf("Data = %T", ev.Data)
	}
}
