package skeleton

import (
	"embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed *.tmpl
var fsys embed.FS

type Data struct {
	CivName        string
	CivCodeUpper   string
	Date           string
	Lang           string
	IncludeHistory bool
	IncludeIcons   bool
	Body           string
}

var templates = map[string]*template.Template{}

func init() {
	for _, name := range []string{"overview", "structree", "common"} {
		raw, err := fsys.ReadFile(name + ".tmpl")
		if err != nil {
			panic(fmt.Sprintf("skeleton: embed read %s.tmpl: %v", name, err))
		}
		tpl, err := template.New(name).Parse(string(raw))
		if err != nil {
			panic(fmt.Sprintf("skeleton: parse %s.tmpl: %v", name, err))
		}
		templates[name] = tpl
	}
}

// Render evaluates the named template (overview/structree/common) with
// the given data and returns the rendered markdown.
func Render(name string, data Data) (string, error) {
	tpl, ok := templates[name]
	if !ok {
		return "", fmt.Errorf("skeleton: unknown template %q", name)
	}
	var sb strings.Builder
	if err := tpl.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("skeleton: execute %s: %w", name, err)
	}
	return sb.String(), nil
}
