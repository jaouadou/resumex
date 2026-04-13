package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"resumex-cli/internal/pdf"
	"resumex-cli/internal/resume"
	"resumex-cli/internal/theme"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "resumex:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("resumex", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	var outputPath string
	var chromePath string
	var scale float64
	flags.StringVar(&outputPath, "o", "", "output PDF path")
	flags.StringVar(&chromePath, "chrome", "", "path to Chrome or Chromium executable")
	flags.Float64Var(&scale, "scale", 1, "theme scale factor; use values below 1 to fit more content")
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "usage: resumex [-o output.pdf] [-chrome /path/to/chrome] [-scale 0.9] resume.json")
		flags.PrintDefaults()
	}

	if err := flags.Parse(reorderFlags(args)); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		flags.Usage()
		return fmt.Errorf("expected exactly one resume JSON file")
	}
	if scale <= 0 {
		return fmt.Errorf("-scale must be greater than 0")
	}

	inputPath := flags.Arg(0)
	if outputPath == "" {
		outputPath = defaultOutputPath(inputPath)
	}

	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	parsed, err := resume.Decode(file)
	if err != nil {
		return err
	}

	html, err := theme.RenderWithOptions(parsed, theme.Options{Scale: scale})
	if err != nil {
		return err
	}

	if err := pdf.RenderHTML(chromePath, html, outputPath); err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, outputPath)
	return nil
}

func defaultOutputPath(inputPath string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	if base == "" {
		base = "resume"
	}
	return base + ".pdf"
}

func reorderFlags(args []string) []string {
	var flagArgs []string
	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-o" || arg == "-chrome" || arg == "-scale":
			flagArgs = append(flagArgs, arg)
			if i+1 < len(args) {
				i++
				flagArgs = append(flagArgs, args[i])
			}
		case strings.HasPrefix(arg, "-o=") || strings.HasPrefix(arg, "-chrome=") || strings.HasPrefix(arg, "-scale="):
			flagArgs = append(flagArgs, arg)
		default:
			positional = append(positional, arg)
		}
	}
	return append(flagArgs, positional...)
}
