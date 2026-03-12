package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateOutputPath creates the output filename with the specified suffix
// Example: photo.jpg with suffix "_imr" -> photo_imr.jpg
func GenerateOutputPath(inputPath, suffix string) string {
	dir := filepath.Dir(inputPath)
	filename := filepath.Base(inputPath)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)
	
	outputFilename := nameWithoutExt + suffix + ext
	return filepath.Join(dir, outputFilename)
}

// FileExists checks if a file exists at the given path
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}
	
	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}
	
	return nil
}

// GetFileInfo returns file info for the given path
func GetFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}
