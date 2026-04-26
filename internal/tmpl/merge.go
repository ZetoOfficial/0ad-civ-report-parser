package tmpl

import (
	"strconv"
	"strings"
)

// Merge applies child on top of parent, returning a new merged element.
// Both inputs are left untouched. The result follows 0 A.D. template merge semantics:
//
//   - If child has attribute replace="" → child fully replaces parent's children
//     (parent's text and grandchildren under that element are dropped).
//   - For elements with datatype="tokens" → token list merge: append unprefixed
//     tokens, drop tokens prefixed with '-' from inherited list.
//   - For numeric leaves with op="add" or op="mul" → arithmetic with parent value.
//   - Otherwise: same-named children are merged recursively. Child text overrides
//     parent text (when child has any text). Children appearing only in child are
//     appended; children only in parent are kept.
func Merge(parent, child *Element) *Element {
	if parent == nil {
		return child.Clone()
	}
	if child == nil {
		return parent.Clone()
	}

	if _, ok := child.Attrs["replace"]; ok {
		return child.Clone()
	}

	if isTokensList(parent) || isTokensList(child) {
		return mergeTokens(parent, child)
	}

	if op := child.Attr("op"); op != "" && len(child.Children) == 0 {
		return mergeOp(parent, child, op)
	}

	out := &Element{Name: child.Name}
	out.Attrs = mergeAttrs(parent.Attrs, child.Attrs)
	if strings.TrimSpace(child.Text) != "" {
		out.Text = child.Text
	} else {
		out.Text = parent.Text
	}

	used := make(map[int]bool)
	for _, pchild := range parent.Children {
		idx := findChildIndex(child.Children, pchild.Name, used)
		if idx >= 0 {
			used[idx] = true
			out.Children = append(out.Children, Merge(pchild, child.Children[idx]))
		} else {
			out.Children = append(out.Children, pchild.Clone())
		}
	}
	for i, cchild := range child.Children {
		if !used[i] {
			out.Children = append(out.Children, cchild.Clone())
		}
	}
	return out
}

func findChildIndex(children []*Element, name string, used map[int]bool) int {
	for i, c := range children {
		if used[i] {
			continue
		}
		if c.Name == name {
			return i
		}
	}
	return -1
}

func mergeAttrs(p, c map[string]string) map[string]string {
	if len(p) == 0 && len(c) == 0 {
		return nil
	}
	out := make(map[string]string)
	for k, v := range p {
		out[k] = v
	}
	for k, v := range c {
		out[k] = v
	}
	return out
}

func isTokensList(e *Element) bool {
	if e == nil {
		return false
	}
	return e.Attr("datatype") == "tokens"
}

func mergeTokens(parent, child *Element) *Element {
	out := &Element{Name: child.Name}
	out.Attrs = mergeAttrs(parent.Attrs, child.Attrs)
	if out.Attrs == nil {
		out.Attrs = make(map[string]string)
	}
	out.Attrs["datatype"] = "tokens"

	toks := append([]string{}, splitTokens(parent.Text)...)
	for _, t := range splitTokens(child.Text) {
		if strings.HasPrefix(t, "-") {
			rem := strings.TrimPrefix(t, "-")
			toks = removeToken(toks, rem)
			continue
		}
		if !containsToken(toks, t) {
			toks = append(toks, t)
		}
	}
	out.Text = strings.Join(toks, " ")
	return out
}

func removeToken(toks []string, target string) []string {
	out := toks[:0]
	for _, t := range toks {
		if t != target {
			out = append(out, t)
		}
	}
	return out
}

func containsToken(toks []string, target string) bool {
	for _, t := range toks {
		if t == target {
			return true
		}
	}
	return false
}

func mergeOp(parent, child *Element, op string) *Element {
	pv, pok := parseFloat(parent.Text)
	cv, cok := parseFloat(child.Text)
	if !cok {
		return child.Clone()
	}
	out := child.Clone()
	delete(out.Attrs, "op")
	if !pok {
		return out
	}
	var result float64
	switch op {
	case "mul":
		result = pv * cv
	case "add":
		result = pv + cv
	default:
		return child.Clone()
	}
	out.Text = formatFloat(result)
	return out
}

func parseFloat(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func formatFloat(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}
