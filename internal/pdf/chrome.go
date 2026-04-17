package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type Options struct {
	ChromePath string
	OutputPath string
	Replace    bool
	Timeout    time.Duration
}

type Result struct {
	Path string
}

func RenderHTML(htmlContent string, options Options) (Result, error) {
	chrome, err := DetectChrome(options.ChromePath)
	if err != nil {
		return Result{}, err
	}

	outputPath, err := PrepareOutputPath(options.OutputPath, options.Replace)
	if err != nil {
		return Result{}, err
	}

	tempDir, err := os.MkdirTemp("", "resumex-*")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(tempDir)

	htmlPath := filepath.Join(tempDir, "resume.html")
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0o600); err != nil {
		return Result{}, err
	}

	cmd := exec.Command(chrome, chromeArgs(tempDir, htmlPath, outputPath)...)
	var commandOutput bytes.Buffer
	cmd.Stdout = &commandOutput
	cmd.Stderr = &commandOutput
	if err := cmd.Start(); err != nil {
		return Result{}, err
	}

	if err := waitForPDF(cmd, outputPath, timeoutOrDefault(options.Timeout)); err != nil {
		return Result{}, fmt.Errorf("chrome pdf generation failed: %w\n%s", err, commandOutput.String())
	}
	return Result{Path: outputPath}, nil
}

func chromeArgs(tempDir, htmlPath, outputPath string) []string {
	return []string{
		"--headless=new",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--disable-background-networking",
		"--disable-extensions",
		"--disable-sync",
		"--metrics-recording-only",
		"--mute-audio",
		"--no-first-run",
		"--no-default-browser-check",
		"--run-all-compositor-stages-before-draw",
		"--user-data-dir=" + filepath.Join(tempDir, "chrome-profile"),
		"--print-to-pdf-no-header",
		"--print-to-pdf=" + outputPath,
		fileURL(htmlPath),
	}
}

func timeoutOrDefault(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return 20 * time.Second
	}
	return timeout
}

func waitForPDF(cmd *exec.Cmd, outputPath string, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	var lastSize int64 = -1
	stableTicks := 0

	for {
		select {
		case err := <-done:
			if err != nil && !hasPDFHeader(outputPath) {
				return err
			}
			return requirePDF(outputPath)
		case <-ticker.C:
			size, ok := pdfSize(outputPath)
			if !ok {
				continue
			}
			if size == lastSize {
				stableTicks++
			} else {
				lastSize = size
				stableTicks = 0
			}
			if stableTicks >= 3 && hasPDFHeader(outputPath) {
				_ = cmd.Process.Kill()
				<-done
				return requirePDF(outputPath)
			}
		case <-deadline.C:
			_ = cmd.Process.Kill()
			<-done
			if hasPDFHeader(outputPath) {
				return requirePDF(outputPath)
			}
			return fmt.Errorf("timed out after %s", timeout)
		}
	}
}

func pdfSize(path string) (int64, bool) {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return 0, false
	}
	return info.Size(), true
}

func requirePDF(path string) error {
	if !hasPDFHeader(path) {
		return fmt.Errorf("output is not a PDF: %s", path)
	}
	return nil
}

func hasPDFHeader(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	header := make([]byte, 4)
	n, err := file.Read(header)
	return err == nil && n == 4 && bytes.Equal(header, []byte("%PDF"))
}

func DetectChrome(override string) (string, error) {
	if override != "" {
		return executablePath(override)
	}

	candidates := chromeCandidates()
	for _, candidate := range candidates {
		if path, err := executablePath(candidate); err == nil {
			return path, nil
		}
	}

	names := []string{
		"google-chrome",
		"google-chrome-stable",
		"chromium",
		"chromium-browser",
		"chrome",
		"msedge",
		"brave-browser",
	}
	for _, name := range names {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}

	return "", errors.New("Chrome/Chromium was not found; install it or pass -chrome")
}

func chromeCandidates() []string {
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		return []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
			filepath.Join(home, "Applications/Google Chrome.app/Contents/MacOS/Google Chrome"),
			filepath.Join(home, "Applications/Chromium.app/Contents/MacOS/Chromium"),
		}
	case "windows":
		return []string{
			filepath.Join(os.Getenv("ProgramFiles"), "Google/Chrome/Application/chrome.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Google/Chrome/Application/chrome.exe"),
			filepath.Join(os.Getenv("LocalAppData"), "Google/Chrome/Application/chrome.exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "Microsoft/Edge/Application/msedge.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft/Edge/Application/msedge.exe"),
		}
	default:
		return []string{
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
		}
	}
}

func executablePath(path string) (string, error) {
	if path == "" {
		return "", os.ErrNotExist
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory", path)
	}
	return path, nil
}

func fileURL(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(abs)}
	return u.String()
}
