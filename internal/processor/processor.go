package processor

import (
	"fmt"

	"github.com/yourusername/image-metadata-remover/internal/fileutil"
	"github.com/yourusername/image-metadata-remover/internal/metadata"
)

// Processor handles image metadata removal
type Processor struct {
	handlers map[ImageFormat]metadata.Handler
}

// NewProcessor creates a new image processor
func NewProcessor() *Processor {
	p := &Processor{
		handlers: make(map[ImageFormat]metadata.Handler),
	}

	// Register handlers
	p.handlers[FormatJPEG] = metadata.NewJPEGHandler()
	p.handlers[FormatPNG] = metadata.NewPNGHandler()
	p.handlers[FormatWebP] = metadata.NewWebPHandler()
	p.handlers[FormatHEIC] = metadata.NewHEICHandler()

	return p
}

// ProcessFile processes a single image file
func (p *Processor) ProcessFile(inputPath, outputPath string) error {
	// Check if input file exists
	if !fileutil.FileExists(inputPath) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}

	// Detect format
	format, err := DetectFormat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to detect format: %w", err)
	}

	if format == FormatUnknown {
		return fmt.Errorf("unsupported or unknown image format")
	}

	// Get handler for this format
	handler, ok := p.handlers[format]
	if !ok {
		return fmt.Errorf("no handler available for format: %s (coming in Phase 2)", format)
	}

	// Process the image
	if err := handler.RemoveMetadata(inputPath, outputPath); err != nil {
		return fmt.Errorf("failed to remove metadata: %w", err)
	}

	return nil
}

// PreviewMetadata displays metadata that would be removed without creating output
func (p *Processor) PreviewMetadata(inputPath string) error {
	// Check if input file exists
	if !fileutil.FileExists(inputPath) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}

	// Detect format
	format, err := DetectFormat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to detect format: %w", err)
	}

	if format == FormatUnknown {
		return fmt.Errorf("unsupported or unknown image format")
	}

	// Get handler for this format
	handler, ok := p.handlers[format]
	if !ok {
		return fmt.Errorf("no handler available for format: %s (coming in Phase 2)", format)
	}

	// Preview the metadata
	if err := handler.PreviewMetadata(inputPath); err != nil {
		return fmt.Errorf("failed to preview metadata: %w", err)
	}

	return nil
}

// GetSupportedFormats returns a list of supported formats
func (p *Processor) GetSupportedFormats() []ImageFormat {
	formats := make([]ImageFormat, 0, len(p.handlers))
	for format := range p.handlers {
		formats = append(formats, format)
	}
	return formats
}
