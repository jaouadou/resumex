package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"resumex-cli/internal/pdf"
)

func TestCLIProducesPDF(t *testing.T) {
	chromePath, err := pdf.DetectChrome("")
	if err != nil {
		t.Skip(err)
	}

	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "nested", "missing")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatal(err)
	}
	outputPath := filepath.Join(outputDir, "resume.pdf")

	cmd := exec.Command("go", "run", ".", "../../testdata/resume-sample.json", "-chrome", chromePath, "-scale", "0.9", "-o", outputDir)
	cmd.Dir = "."
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go run failed: %v\n%s", err, stderr.String())
	}

	cmd = exec.Command("go", "run", ".", "../../testdata/resume-sample.json", "-chrome", chromePath, "-scale", "0.9", "-replace", "-o", outputDir)
	cmd.Dir = "."
	stderr.Reset()
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go run with -replace failed: %v\n%s", err, stderr.String())
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) < 1000 {
		t.Fatalf("expected non-trivial PDF, got %d bytes", len(raw))
	}
	if !bytes.HasPrefix(raw, []byte("%PDF")) {
		t.Fatalf("expected PDF header, got %q", raw[:4])
	}
}
