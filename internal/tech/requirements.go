package tech

import (
	"fmt"
	"sort"
	"strings"
)

// DescribeRequirements renders a Requirements tree as human-readable
// Russian text. Used as the fallback when a tech does not provide its
// own requirementsTooltip. Handles entity/all/any/notciv/civ/tech.
//
// Examples:
//
//	{"tech":"phase_town"}
//	  → "технология: phase_town"
//	{"entity":{"class":"Village","number":5}}
//	  → "5+ зданий класса Village"
//	{"all":[{"tech":"phase_town"},{"entity":{"class":"Village","number":5}}]}
//	  → "(технология: phase_town И 5+ зданий класса Village)"
func DescribeRequirements(req Requirements) string {
	if req == nil {
		return ""
	}
	parts := []string{}
	for _, key := range sortedKeys(req) {
		switch key {
		case "civ":
			if v, ok := req["civ"].(string); ok {
				parts = append(parts, "цивилизация: "+v)
			}
		case "notciv":
			if list := stringOrList(req["notciv"]); len(list) > 0 {
				parts = append(parts, "не для цивилизаций: "+strings.Join(list, ", "))
			}
		case "tech":
			if v, ok := req["tech"].(string); ok {
				parts = append(parts, "технология: "+v)
			}
		case "entity":
			parts = append(parts, describeEntity(req["entity"]))
		case "all":
			parts = append(parts, describeJoin(req["all"], " И "))
		case "any":
			parts = append(parts, describeJoin(req["any"], " ИЛИ "))
		}
	}
	out := []string{}
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, "; ")
}

func describeEntity(v any) string {
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	class, _ := m["class"].(string)
	number := 0
	switch n := m["number"].(type) {
	case float64:
		number = int(n)
	case int:
		number = n
	}
	switch {
	case class != "" && number > 0:
		return fmt.Sprintf("%d+ зданий класса %s", number, class)
	case class != "":
		return "класс: " + class
	}
	return ""
}

func describeJoin(v any, sep string) string {
	list, ok := v.([]any)
	if !ok {
		return ""
	}
	parts := []string{}
	for _, item := range list {
		if m, ok := item.(map[string]any); ok {
			if s := DescribeRequirements(Requirements(m)); s != "" {
				parts = append(parts, s)
			}
		}
	}
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return "(" + strings.Join(parts, sep) + ")"
}

func stringOrList(v any) []string {
	switch x := v.(type) {
	case string:
		if x == "" {
			return nil
		}
		return []string{x}
	case []any:
		out := []string{}
		for _, e := range x {
			if s, ok := e.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func sortedKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
