package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"resumex-cli/internal/pdf"
	"resumex-cli/internal/resume"
	"resumex-cli/internal/theme"
)

type config struct {
	inputPath  string
	outputPath string
	chromePath string
	scale      float64
	replace    bool
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "resumex:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := parseConfig(args, os.Stderr)
	if err != nil {
		return err
	}

	parsed, err := readResume(cfg.inputPath)
	if err != nil {
		return err
	}

	html, err := theme.RenderWithOptions(parsed, theme.Options{Scale: cfg.scale})
	if err != nil {
		return err
	}

	result, err := pdf.RenderHTML(html, pdf.Options{
		ChromePath: cfg.chromePath,
		OutputPath: cfg.outputPath,
		Replace:    cfg.replace,
	})
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, result.Path)
	return nil
}

func parseConfig(args []string, output io.Writer) (config, error) {
	flags := flag.NewFlagSet("resumex", flag.ContinueOnError)
	flags.SetOutput(output)

	cfg := config{scale: 1}
	flags.StringVar(&cfg.outputPath, "o", "", "output PDF path")
	flags.StringVar(&cfg.chromePath, "chrome", "", "path to Chrome or Chromium executable")
	flags.Float64Var(&cfg.scale, "scale", cfg.scale, "theme scale factor; use values below 1 to fit more content")
	flags.BoolVar(&cfg.replace, "replace", false, "replace the output PDF if it already exists")
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "usage: resumex [-o output.pdf] [-chrome /path/to/chrome] [-scale 0.9] [-replace] resume.json")
		flags.PrintDefaults()
	}

	if err := flags.Parse(normalizeArgs(args)); err != nil {
		return config{}, err
	}
	if flags.NArg() != 1 {
		flags.Usage()
		return config{}, fmt.Errorf("expected exactly one resume JSON file")
	}
	if cfg.scale <= 0 {
		return config{}, fmt.Errorf("-scale must be greater than 0")
	}

	cfg.inputPath = flags.Arg(0)
	cfg.outputPath = defaultString(cfg.outputPath, defaultOutputPath(cfg.inputPath))
	return cfg, nil
}

func readResume(path string) (*resume.Resume, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return resume.Decode(file)
}

func defaultOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	if base == "" {
		base = "resume"
	}
	return base + ".pdf"
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func normalizeArgs(args []string) []string {
	var flagArgs []string
	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]

		name, hasValueInline := flagName(arg)
		switch flagKind(name) {
		case boolFlag:
			flagArgs = append(flagArgs, arg)
		case valueFlag:
			flagArgs = append(flagArgs, arg)
			if !hasValueInline && i+1 < len(args) {
				i++
				flagArgs = append(flagArgs, args[i])
			}
		default:
			positional = append(positional, arg)
		}
	}
	return append(flagArgs, positional...)
}

type argFlagKind int

const (
	notAFlag argFlagKind = iota
	boolFlag
	valueFlag
)

func flagKind(name string) argFlagKind {
	switch name {
	case "replace":
		return boolFlag
	case "o", "chrome", "scale":
		return valueFlag
	default:
		return notAFlag
	}
}

func flagName(arg string) (string, bool) {
	if !strings.HasPrefix(arg, "-") || arg == "-" {
		return "", false
	}
	trimmed := strings.TrimLeft(arg, "-")
	name, _, hasValue := strings.Cut(trimmed, "=")
	return name, hasValue
}
