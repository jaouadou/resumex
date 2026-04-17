package pdf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultDirectoryOutputName = "resume.pdf"

func PrepareOutputPath(requestedPath string, replace bool) (string, error) {
	outputPath, err := ResolveOutputPath(requestedPath)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return "", err
	}
	if err := prepareOutputFile(outputPath, replace); err != nil {
		return "", err
	}
	return outputPath, nil
}

func ResolveOutputPath(requestedPath string) (string, error) {
	if requestedPath == "" {
		return "", errors.New("output path is empty")
	}

	dirLike := hasTrailingPathSeparator(requestedPath)
	abs, err := filepath.Abs(requestedPath)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(abs)
	switch {
	case err == nil && info.IsDir():
		return filepath.Join(abs, defaultDirectoryOutputName), nil
	case err == nil:
		return abs, nil
	case err != nil && !os.IsNotExist(err):
		return "", err
	case dirLike:
		return filepath.Join(abs, defaultDirectoryOutputName), nil
	default:
		return abs, nil
	}
}

func prepareOutputFile(outputPath string, replace bool) error {
	info, err := os.Stat(outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("output path is a directory: %s", outputPath)
	}
	if !replace {
		return fmt.Errorf("output file already exists: %s (use -replace to overwrite)", outputPath)
	}
	return os.Remove(outputPath)
}

func hasTrailingPathSeparator(path string) bool {
	return strings.HasSuffix(path, "/") || strings.HasSuffix(path, `\`)
}
