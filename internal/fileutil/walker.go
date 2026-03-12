package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SupportedExtensions lists image file extensions
var SupportedExtensions = []string{
	".jpg", ".jpeg", ".png", ".webp", ".tiff", ".tif", ".gif",
}

// IsImageFile checks if a file has a supported image extension
func IsImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, supportedExt := range SupportedExtensions {
		if ext == supportedExt {
			return true
		}
	}
	return false
}

// CollectImageFiles walks a path and collects all image files
// If path is a file, returns just that file
// If path is a directory and recursive is true, walks all subdirectories
// If path is a directory and recursive is false, only processes files in that directory
func CollectImageFiles(path string, recursive bool) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	// If it's a file, return it directly
	if !info.IsDir() {
		if IsImageFile(path) {
			return []string{path}, nil
		}
		return []string{path}, nil // Return it anyway, let format detection handle it
	}

	// It's a directory, collect image files
	var imageFiles []string

	if recursive {
		// Recursive walk
		err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && IsImageFile(filePath) {
				imageFiles = append(imageFiles, filePath)
			}
			return nil
		})
	} else {
		// Non-recursive, just read the directory
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				filePath := filepath.Join(path, entry.Name())
				if IsImageFile(filePath) {
					imageFiles = append(imageFiles, filePath)
				}
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return imageFiles, nil
}
