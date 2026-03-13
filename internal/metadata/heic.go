package metadata

import (
	"fmt"
	"sort"

	"github.com/maw629/heic-meta/pkg/heicmeta"
)

// HEICHandler handles HEIC/HEIF image metadata via heic-meta package.
type HEICHandler struct{}

// NewHEICHandler creates a new HEIC metadata handler.
func NewHEICHandler() *HEICHandler {
	return &HEICHandler{}
}

// SupportsFormat returns true if the format is HEIC/HEIF.
func (h *HEICHandler) SupportsFormat(format string) bool {
	return format == "heic" || format == "heif"
}

// RemoveMetadata removes sensitive metadata from HEIC/HEIF files.
func (h *HEICHandler) RemoveMetadata(inputPath, outputPath string) error {
	if err := heicmeta.RemoveMetadata(inputPath, outputPath, heicmeta.DefaultOptions); err != nil {
		return fmt.Errorf("failed to remove HEIC metadata: %w", err)
	}
	return nil
}

// PreviewMetadata displays metadata that would be removed.
func (h *HEICHandler) PreviewMetadata(inputPath string) error {
	meta, err := heicmeta.PreviewMetadata(inputPath)
	if err != nil {
		return fmt.Errorf("failed to preview HEIC metadata: %w", err)
	}

	toRemove := append([]string{}, meta.SensitiveTags...)
	if len(toRemove) == 0 {
		for _, tag := range meta.EXIFSensitiveTags {
			toRemove = append(toRemove, "EXIF."+tag)
		}
		for _, field := range meta.XMPSensitiveFields {
			toRemove = append(toRemove, "XMP."+field)
		}
	}
	sort.Strings(toRemove)

	fmt.Printf("\n=== Metadata Preview for: %s ===\n", inputPath)
	fmt.Printf("Format: HEIC/HEIF\n\n")

	if len(toRemove) > 0 {
		fmt.Println("⚠️  Metadata to be REMOVED (sensitive):")
		for _, item := range toRemove {
			fmt.Printf("  - %s\n", item)
		}
		fmt.Println()
	}

	fmt.Println("✓ Metadata to be PRESERVED:")
	fmt.Println("  - HEIC image data (mdat)")
	fmt.Println("  - Non-sensitive HEIC container structure")
	fmt.Println()

	if len(toRemove) == 0 && !meta.EXIFPresent && !meta.XMPPresent {
		fmt.Println("ℹ️  No metadata found in this HEIC/HEIF file")
	}

	return nil
}
