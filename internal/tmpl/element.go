package tmpl

import (
	"strconv"
	"strings"
)

type Element struct {
	Name     string
	Attrs    map[string]string
	Text     string
	Children []*Element
}

func (e *Element) Attr(name string) string {
	if e.Attrs == nil {
		return ""
	}
	return e.Attrs[name]
}

func (e *Element) HasAttr(name string) bool {
	if e.Attrs == nil {
		return false
	}
	_, ok := e.Attrs[name]
	return ok
}

func (e *Element) Child(name string) *Element {
	for _, c := range e.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (e *Element) Get(path string) *Element {
	cur := e
	for _, seg := range strings.Split(path, "/") {
		if cur == nil {
			return nil
		}
		cur = cur.Child(seg)
	}
	return cur
}

func (e *Element) GetText(path string) string {
	c := e.Get(path)
	if c == nil {
		return ""
	}
	return strings.TrimSpace(c.Text)
}

func (e *Element) GetFloat(path string) (float64, bool) {
	t := e.GetText(path)
	if t == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func (e *Element) GetInt(path string) (int, bool) {
	v, ok := e.GetFloat(path)
	if !ok {
		return 0, false
	}
	return int(v), true
}

func (e *Element) GetTokens(path string) []string {
	c := e.Get(path)
	if c == nil {
		return nil
	}
	return splitTokens(c.Text)
}

func splitTokens(text string) []string {
	fs := strings.Fields(text)
	out := make([]string, 0, len(fs))
	for _, f := range fs {
		f = strings.TrimSpace(f)
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

func (e *Element) Clone() *Element {
	c := &Element{
		Name: e.Name,
		Text: e.Text,
	}
	if e.Attrs != nil {
		c.Attrs = make(map[string]string, len(e.Attrs))
		for k, v := range e.Attrs {
			c.Attrs[k] = v
		}
	}
	for _, ch := range e.Children {
		c.Children = append(c.Children, ch.Clone())
	}
	return c
}
