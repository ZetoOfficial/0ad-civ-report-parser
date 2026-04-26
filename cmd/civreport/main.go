package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/render"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

func main() {
	var (
		gamedataFlag string
		outFlag      string
		all          bool
		check        bool
	)
	flag.StringVar(&gamedataFlag, "gamedata", "", "path to 0 A.D. mods/public root (overrides OAD_GAMEDATA_ROOT)")
	flag.StringVar(&outFlag, "out", "", "output file (default: <civ>_buildings_report.md in CWD)")
	flag.BoolVar(&all, "all", false, "generate reports for all 15 civilizations")
	flag.BoolVar(&check, "check", false, "smoke-check: parse all civs without writing files")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: civreport [flags] <civ>\n\n")
		fmt.Fprintf(os.Stderr, "Generate a Russian-language buildings/units/technologies report\n")
		fmt.Fprintf(os.Stderr, "for one or more 0 A.D. civilizations.\n\n")
		fmt.Fprintf(os.Stderr, "Civilization codes: athen, brit, cart, gaul, germ, han, iber, kush,\n")
		fmt.Fprintf(os.Stderr, "                    mace, maur, pers, ptol, rome, sele, spart\n")
		fmt.Fprintf(os.Stderr, "Russian aliases also supported (спарт, афин, германцы, ...)\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	root := paths.ResolveRoot(gamedataFlag)
	layout := paths.Layout{Root: root}
	if _, err := os.Stat(layout.Templates()); err != nil {
		fail("gamedata templates not found at %s (use --gamedata or set %s): %v",
			layout.Templates(), paths.EnvGameDataRoot, err)
	}

	idx, err := tmpl.NewIndex(layout.Templates())
	if err != nil {
		fail("template index: %v", err)
	}
	resolver := tmpl.NewResolver(idx)
	gen := render.NewGenerator(layout, resolver)

	switch {
	case check:
		runCheck(gen)
	case all:
		runAll(gen, outFlag)
	default:
		args := flag.Args()
		if len(args) != 1 {
			flag.Usage()
			os.Exit(2)
		}
		runOne(gen, args[0], outFlag)
	}
}

func runOne(gen *render.Generator, input, outFlag string) {
	info, ok := civdata.ResolveCivInput(input)
	if !ok {
		fail("could not resolve civilization %q. Try one of: athen, spart, germ, ...", input)
	}
	out, err := gen.Generate(info)
	if err != nil {
		fail("generate %s: %v", info.Code, err)
	}
	body := out.Overview + "\n" + out.Structree
	outPath := outFlag
	if outPath == "" {
		outPath = info.OutputFile
	}
	if err := os.WriteFile(outPath, []byte(body), 0o600); err != nil {
		fail("write %s: %v", outPath, err)
	}
	abs, _ := filepath.Abs(outPath)
	lines := strings.Count(body, "\n") + 1
	fmt.Printf("OK %s → %s (%d lines)\n", info.Code, abs, lines)
}

func runAll(gen *render.Generator, outFlag string) {
	for _, civInfo := range civdata.Civilizations {
		out, err := gen.Generate(civInfo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", civInfo.Code, err)
			continue
		}
		body := out.Overview + "\n" + out.Structree
		outPath := civInfo.OutputFile
		if outFlag != "" {
			outPath = filepath.Join(outFlag, civInfo.OutputFile)
		}
		if err := os.WriteFile(outPath, []byte(body), 0o600); err != nil {
			fmt.Fprintf(os.Stderr, "WRITE %s: %v\n", civInfo.Code, err)
			continue
		}
		lines := strings.Count(body, "\n") + 1
		fmt.Printf("OK %s → %s (%d lines)\n", civInfo.Code, outPath, lines)
	}
}

func runCheck(gen *render.Generator) {
	failed := 0
	for _, civInfo := range civdata.Civilizations {
		out, err := gen.Generate(civInfo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", civInfo.Code, err)
			failed++
			continue
		}
		body := out.Overview + "\n" + out.Structree
		lines := strings.Count(body, "\n") + 1
		ok := lines >= 100
		mark := "OK"
		if !ok {
			mark = "WARN"
			failed++
		}
		fmt.Printf("%s %s (%d lines)\n", mark, civInfo.Code, lines)
	}
	if failed > 0 {
		os.Exit(1)
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "civreport: "+format+"\n", args...)
	os.Exit(1)
}
