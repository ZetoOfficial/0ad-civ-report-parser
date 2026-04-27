package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/config"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/render"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/render/skeleton"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

func main() {
	var (
		gamedataFlag   string
		outDirFlag     string
		configFlag     string
		printBaseName  bool
		all            bool
		check          bool
		includeHistory bool
	)
	flag.StringVar(&gamedataFlag, "gamedata", "", "path to 0 A.D. mods/public root (overrides OAD_GAMEDATA_ROOT and config)")
	flag.StringVar(&outDirFlag, "out-dir", "", "output directory for generated files (default: from config or '.')")
	flag.StringVar(&configFlag, "config", "", "path to JSON config file")
	flag.BoolVar(&printBaseName, "print-basename", false, "print BaseName for the given civ and exit (used by Makefile)")
	flag.BoolVar(&all, "all", false, "generate reports for all 15 civilizations")
	flag.BoolVar(&check, "check", false, "smoke-check: parse all civs without writing files")
	flag.BoolVar(&includeHistory, "include-history", false, "include the civ history paragraph in <civ>_overview.md (off by default)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: civreport [flags] <civ>\n\n")
		fmt.Fprintf(os.Stderr, "Generate Russian-language overview + structure-tree reports\n")
		fmt.Fprintf(os.Stderr, "for one or more 0 A.D. civilizations.\n\n")
		fmt.Fprintf(os.Stderr, "Output: <civ>_overview.md, <civ>_structree.md, common.md.\n\n")
		fmt.Fprintf(os.Stderr, "Civ codes: athen, brit, cart, gaul, germ, han, iber, kush,\n")
		fmt.Fprintf(os.Stderr, "           mace, maur, pers, ptol, rome, sele, spart\n")
		fmt.Fprintf(os.Stderr, "Russian aliases also supported (спарт, афин, германцы, ...)\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// TODO(post-epic-1): auto-discover ./config.json next to the binary
	// when --config is absent (per spec). For now, JSON config is opt-in
	// via --config <path>.
	cfg, err := config.Load(configFlag)
	if err != nil {
		fail("config: %v", err)
	}
	// Precedence: CLI flag > JSON config > env var > built-in default.
	// Env var only applies when neither --gamedata nor --config is given.
	if gamedataFlag != "" {
		cfg.Gamedata = gamedataFlag
	} else if configFlag == "" {
		if env := os.Getenv(paths.EnvGameDataRoot); env != "" {
			cfg.Gamedata = env
		}
	}
	if outDirFlag != "" {
		cfg.OutDir = outDirFlag
	}
	if includeHistory {
		cfg.IncludeHistory = true
	}

	if printBaseName {
		args := flag.Args()
		if len(args) != 1 {
			fail("--print-basename requires exactly one civ argument")
		}
		info, ok := civdata.ResolveCivInput(args[0])
		if !ok {
			fail("could not resolve civilization %q", args[0])
		}
		fmt.Println(info.BaseName)
		return
	}

	if _, err := os.Stat(filepath.Join(cfg.Gamedata, "simulation", "templates")); err != nil {
		fail("gamedata templates not found at %s/simulation/templates: %v", cfg.Gamedata, err)
	}

	idx, err := tmpl.NewIndex(filepath.Join(cfg.Gamedata, "simulation", "templates"))
	if err != nil {
		fail("template index: %v", err)
	}
	resolver := tmpl.NewResolver(idx)
	gen := render.NewGenerator(paths.Layout{Root: cfg.Gamedata}, resolver)
	gen.IncludeHistory = cfg.IncludeHistory

	switch {
	case check:
		runCheck(gen, cfg)
	case all:
		runAll(gen, cfg)
	default:
		args := flag.Args()
		if len(args) != 1 {
			flag.Usage()
			os.Exit(2)
		}
		runOne(gen, cfg, args[0])
	}
}

func runOne(gen *render.Generator, cfg *config.Config, input string) {
	info, ok := civdata.ResolveCivInput(input)
	if !ok {
		fail("could not resolve civilization %q. Try one of: athen, spart, germ, ...", input)
	}
	out, err := gen.Generate(info)
	if err != nil {
		fail("generate %s: %v", info.Code, err)
	}
	if err := writeCivFiles(cfg, info, out); err != nil {
		fail("write %s: %v", info.Code, err)
	}
	if err := writeCommon(cfg, gen); err != nil {
		fail("write common.md: %v", err)
	}
	abs, _ := filepath.Abs(cfg.OutDir)
	fmt.Printf("OK %s → %s/{%s,%s} + common.md\n", info.Code, abs, info.OverviewFile(), info.StructreeFile())
}

func runAll(gen *render.Generator, cfg *config.Config) {
	for _, info := range civdata.Civilizations {
		out, err := gen.Generate(info)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", info.Code, err)
			continue
		}
		if err := writeCivFiles(cfg, info, out); err != nil {
			fmt.Fprintf(os.Stderr, "WRITE %s: %v\n", info.Code, err)
			continue
		}
		fmt.Printf("OK %s → %s, %s\n", info.Code, info.OverviewFile(), info.StructreeFile())
	}
	if err := writeCommon(cfg, gen); err != nil {
		fmt.Fprintf(os.Stderr, "WRITE common.md: %v\n", err)
	} else {
		fmt.Println("OK common.md")
	}
}

func runCheck(gen *render.Generator, cfg *config.Config) {
	failed := 0
	for _, info := range civdata.Civilizations {
		out, err := gen.Generate(info)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", info.Code, err)
			failed++
			continue
		}
		ovLines := strings.Count(out.Overview, "\n") + 1
		stLines := strings.Count(out.Structree, "\n") + 1
		mark := "OK"
		if ovLines < 30 || stLines < 100 {
			// Informational only. The 30/100 thresholds are deliberately loose
			// — only Generate/RenderCommon errors are fatal. Tighter regression
			// gate lives in TestGoldenGermStructure.
			mark = "WARN"
		}
		fmt.Printf("%s %s (overview=%d, structree=%d)\n", mark, info.Code, ovLines, stLines)
	}
	if _, err := gen.RenderCommon(); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL common: %v\n", err)
		failed++
	}
	_ = cfg
	if failed > 0 {
		os.Exit(1)
	}
}

func writeCivFiles(cfg *config.Config, info civdata.CivCode, out render.Output) error {
	if err := ensureOutDir(cfg); err != nil {
		return err
	}
	date := time.Now().Format("2006-01-02") // Go reference layout = YYYY-MM-DD
	codeUpper := strings.ToUpper(info.Code[:1]) + info.Code[1:]
	overview, err := skeleton.Render("overview", skeleton.Data{
		CivName:        info.NameEN,
		CivCodeUpper:   codeUpper,
		Date:           date,
		Lang:           cfg.Lang,
		IncludeHistory: cfg.IncludeHistory,
		IncludeIcons:   cfg.IncludeIcons,
		Body:           out.Overview,
	})
	if err != nil {
		return fmt.Errorf("render overview skeleton: %w", err)
	}
	structree, err := skeleton.Render("structree", skeleton.Data{
		CivName:        info.NameEN,
		CivCodeUpper:   codeUpper,
		Date:           date,
		Lang:           cfg.Lang,
		IncludeHistory: cfg.IncludeHistory,
		IncludeIcons:   cfg.IncludeIcons,
		Body:           out.Structree,
	})
	if err != nil {
		return fmt.Errorf("render structree skeleton: %w", err)
	}
	if err := os.WriteFile(filepath.Join(cfg.OutDir, info.OverviewFile()), []byte(overview), 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cfg.OutDir, info.StructreeFile()), []byte(structree), 0o600); err != nil {
		return err
	}
	return nil
}

func writeCommon(cfg *config.Config, gen *render.Generator) error {
	if err := ensureOutDir(cfg); err != nil {
		return err
	}
	body, err := gen.RenderCommon()
	if err != nil {
		return err
	}
	wrapped, err := skeleton.Render("common", skeleton.Data{
		Date: time.Now().Format("2006-01-02"),
		Body: body,
	})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(cfg.OutDir, "common.md"), []byte(wrapped), 0o600)
}

func ensureOutDir(cfg *config.Config) error {
	if cfg.OutDir == "" {
		cfg.OutDir = "."
	}
	return os.MkdirAll(cfg.OutDir, 0o755)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "civreport: "+format+"\n", args...)
	os.Exit(1)
}
