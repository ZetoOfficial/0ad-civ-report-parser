package tmpl

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func ParseFile(path string) (*Element, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	return Parse(f, path)
}

func Parse(r io.Reader, source string) (*Element, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	var root *Element
	stack := []*Element{}
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("xml %s: %w", source, err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			el := &Element{Name: t.Name.Local}
			if len(t.Attr) > 0 {
				el.Attrs = make(map[string]string, len(t.Attr))
				for _, a := range t.Attr {
					el.Attrs[a.Name.Local] = a.Value
				}
			}
			if len(stack) == 0 {
				root = el
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, el)
			}
			stack = append(stack, el)
		case xml.EndElement:
			if len(stack) == 0 {
				return nil, fmt.Errorf("xml %s: unbalanced end element %s", source, t.Name.Local)
			}
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if len(top.Children) > 0 {
				top.Text = strings.TrimSpace(top.Text)
			}
		case xml.CharData:
			if len(stack) > 0 {
				top := stack[len(stack)-1]
				top.Text += string(t)
			}
		}
	}
	if root == nil {
		return nil, fmt.Errorf("xml %s: no root element", source)
	}
	return root, nil
}
