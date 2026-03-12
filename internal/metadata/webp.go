package metadata

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

// WebPHandler handles WebP image metadata
type WebPHandler struct{}

// NewWebPHandler creates a new WebP metadata handler
func NewWebPHandler() *WebPHandler {
	return &WebPHandler{}
}

// SupportsFormat returns true if the format is WebP
func (h *WebPHandler) SupportsFormat(format string) bool {
	return format == "webp"
}

// RIFF chunk structure
type riffChunk struct {
	FourCC [4]byte
	Size   uint32
	Data   []byte
}

// WebP file structure
var (
	riffMagic = [4]byte{'R', 'I', 'F', 'F'}
	webpMagic = [4]byte{'W', 'E', 'B', 'P'}
)

// PreviewMetadata displays metadata that would be removed
func (h *WebPHandler) PreviewMetadata(inputPath string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read RIFF header
	var riffHeader [4]byte
	var fileSize uint32
	var webpHeader [4]byte

	if err := binary.Read(file, binary.LittleEndian, &riffHeader); err != nil {
		return fmt.Errorf("failed to read RIFF header: %w", err)
	}
	if riffHeader != riffMagic {
		return fmt.Errorf("invalid RIFF header")
	}

	if err := binary.Read(file, binary.LittleEndian, &fileSize); err != nil {
		return fmt.Errorf("failed to read file size: %w", err)
	}

	if err := binary.Read(file, binary.LittleEndian, &webpHeader); err != nil {
		return fmt.Errorf("failed to read WebP header: %w", err)
	}
	if webpHeader != webpMagic {
		return fmt.Errorf("invalid WebP header")
	}

	toRemove := []string{}
	toPreserve := []string{}
	var exifData *exif.Exif

	// Read all chunks
	for {
		chunk, err := h.readChunk(file)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read chunk: %w", err)
		}

		chunkType := string(chunk.FourCC[:])

		switch chunkType {
		case "EXIF":
			// Parse EXIF data (skip first 4 bytes which are padding)
			if len(chunk.Data) > 4 {
				exifData, err = exif.Decode(bytes.NewReader(chunk.Data[4:]))
				if err != nil {
					toRemove = append(toRemove, fmt.Sprintf("EXIF (failed to parse: %v)", err))
				}
			}

		case "XMP ":
			// XMP metadata - show preview
			preview := string(chunk.Data)
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			// Check if it contains sensitive patterns
			lowerXMP := strings.ToLower(string(chunk.Data))
			hasSensitive := strings.Contains(lowerXMP, "author") ||
				strings.Contains(lowerXMP, "creator") ||
				strings.Contains(lowerXMP, "copyright") ||
				strings.Contains(lowerXMP, "gps") ||
				strings.Contains(lowerXMP, "location")
			
			if hasSensitive {
				toRemove = append(toRemove, fmt.Sprintf("XMP  (XML metadata, %d bytes)", len(chunk.Data)))
			} else {
				toPreserve = append(toPreserve, fmt.Sprintf("XMP  (XML metadata, %d bytes)", len(chunk.Data)))
			}

		case "ICCP":
			toPreserve = append(toPreserve, fmt.Sprintf("ICCP (ICC Color Profile, %d bytes)", len(chunk.Data)))

		case "VP8 ":
			toPreserve = append(toPreserve, "VP8  (Lossy Image Data)")
		case "VP8L":
			toPreserve = append(toPreserve, "VP8L (Lossless Image Data)")
		case "VP8X":
			toPreserve = append(toPreserve, "VP8X (Extended Format)")
		case "ALPH":
			toPreserve = append(toPreserve, "ALPH (Alpha Channel)")
		case "ANIM":
			toPreserve = append(toPreserve, "ANIM (Animation Parameters)")
		case "ANMF":
			// Don't list every animation frame
			if len(toPreserve) == 0 || toPreserve[len(toPreserve)-1] != "ANMF (Animation Frames)" {
				toPreserve = append(toPreserve, "ANMF (Animation Frames)")
			}

		default:
			// Unknown chunk
			toPreserve = append(toPreserve, fmt.Sprintf("%s (Unknown, %d bytes)", chunkType, len(chunk.Data)))
		}
	}

	// Handle EXIF data if present
	if exifData != nil {
		sensitiveTags := getSensitiveTagNames()
		
		walker := &webpExifCollector{
			sensitiveTags: sensitiveTags,
			toRemove:      &toRemove,
			toPreserve:    &toPreserve,
		}
		
		exifData.Walk(walker)
	}

	// Sort lists
	sort.Strings(toRemove)
	sort.Strings(toPreserve)

	// Display results
	fmt.Printf("\n=== Metadata Preview for: %s ===\n", inputPath)
	fmt.Printf("Format: WebP\n\n")

	if len(toRemove) > 0 {
		fmt.Println("⚠️  Metadata to be REMOVED (sensitive):")
		for _, item := range toRemove {
			fmt.Printf("  - %s\n", item)
		}
		fmt.Println()
	}

	if len(toPreserve) > 0 {
		fmt.Println("✓ Metadata to be PRESERVED:")
		for _, item := range toPreserve {
			fmt.Printf("  - %s\n", item)
		}
		fmt.Println()
	}

	if len(toRemove) == 0 && len(toPreserve) == 0 {
		fmt.Println("ℹ️  No metadata found in this WebP file")
	}

	return nil
}

// RemoveMetadata removes sensitive metadata from WebP files
func (h *WebPHandler) RemoveMetadata(inputPath, outputPath string) error {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer inputFile.Close()

	// Read RIFF header
	var riffHeader [4]byte
	var fileSize uint32
	var webpHeader [4]byte

	if err := binary.Read(inputFile, binary.LittleEndian, &riffHeader); err != nil {
		return fmt.Errorf("failed to read RIFF header: %w", err)
	}
	if riffHeader != riffMagic {
		return fmt.Errorf("invalid RIFF header")
	}

	if err := binary.Read(inputFile, binary.LittleEndian, &fileSize); err != nil {
		return fmt.Errorf("failed to read file size: %w", err)
	}

	if err := binary.Read(inputFile, binary.LittleEndian, &webpHeader); err != nil {
		return fmt.Errorf("failed to read WebP header: %w", err)
	}
	if webpHeader != webpMagic {
		return fmt.Errorf("invalid WebP header")
	}

	// Collect chunks to keep
	var chunksToKeep []*riffChunk

	// Read all chunks and filter
	for {
		chunk, err := h.readChunk(inputFile)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read chunk: %w", err)
		}

		chunkType := string(chunk.FourCC[:])
		shouldKeep := true

		switch chunkType {
		case "EXIF":
			// Filter EXIF chunk
			chunk, shouldKeep = h.filterEXIFChunk(chunk)

		case "XMP ":
			// Check if XMP contains sensitive data
			lowerXMP := strings.ToLower(string(chunk.Data))
			hasSensitive := strings.Contains(lowerXMP, "author") ||
				strings.Contains(lowerXMP, "creator") ||
				strings.Contains(lowerXMP, "copyright") ||
				strings.Contains(lowerXMP, "gps") ||
				strings.Contains(lowerXMP, "location")
			
			// Remove XMP if it has sensitive data
			// TODO: Could implement selective XMP filtering in the future
			shouldKeep = !hasSensitive

		case "ICCP", "VP8 ", "VP8L", "VP8X", "ALPH", "ANIM", "ANMF":
			// Keep all image data and color profile chunks
			shouldKeep = true

		default:
			// Keep unknown chunks by default (conservative approach)
			shouldKeep = true
		}

		if shouldKeep {
			chunksToKeep = append(chunksToKeep, chunk)
		}
	}

	// Calculate new file size
	newFileSize := uint32(4) // "WEBP" fourCC
	for _, chunk := range chunksToKeep {
		newFileSize += 8 + chunk.Size // fourCC + size + data
		if chunk.Size%2 == 1 {
			newFileSize++ // padding byte
		}
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Write RIFF header
	if err := binary.Write(outputFile, binary.LittleEndian, riffHeader); err != nil {
		return fmt.Errorf("failed to write RIFF header: %w", err)
	}
	if err := binary.Write(outputFile, binary.LittleEndian, newFileSize); err != nil {
		return fmt.Errorf("failed to write file size: %w", err)
	}
	if err := binary.Write(outputFile, binary.LittleEndian, webpHeader); err != nil {
		return fmt.Errorf("failed to write WebP header: %w", err)
	}

	// Write chunks
	for _, chunk := range chunksToKeep {
		if err := h.writeChunk(outputFile, chunk); err != nil {
			return fmt.Errorf("failed to write chunk: %w", err)
		}
	}

	return nil
}

// readChunk reads a single RIFF chunk
func (h *WebPHandler) readChunk(r io.Reader) (*riffChunk, error) {
	chunk := &riffChunk{}

	// Read fourCC
	if err := binary.Read(r, binary.LittleEndian, &chunk.FourCC); err != nil {
		return nil, err
	}

	// Read size
	if err := binary.Read(r, binary.LittleEndian, &chunk.Size); err != nil {
		return nil, err
	}

	// Read data
	chunk.Data = make([]byte, chunk.Size)
	if chunk.Size > 0 {
		if _, err := io.ReadFull(r, chunk.Data); err != nil {
			return nil, err
		}
	}

	// Skip padding byte if size is odd
	if chunk.Size%2 == 1 {
		var padding byte
		binary.Read(r, binary.LittleEndian, &padding)
	}

	return chunk, nil
}

// writeChunk writes a single RIFF chunk
func (h *WebPHandler) writeChunk(w io.Writer, chunk *riffChunk) error {
	// Write fourCC
	if err := binary.Write(w, binary.LittleEndian, chunk.FourCC); err != nil {
		return err
	}

	// Write size
	if err := binary.Write(w, binary.LittleEndian, chunk.Size); err != nil {
		return err
	}

	// Write data
	if len(chunk.Data) > 0 {
		if _, err := w.Write(chunk.Data); err != nil {
			return err
		}
	}

	// Add padding byte if size is odd
	if chunk.Size%2 == 1 {
		if err := binary.Write(w, binary.LittleEndian, byte(0)); err != nil {
			return err
		}
	}

	return nil
}

// filterEXIFChunk filters sensitive EXIF tags from EXIF chunk
func (h *WebPHandler) filterEXIFChunk(chunk *riffChunk) (*riffChunk, bool) {
	// WebP EXIF chunks have 4 bytes of padding at the start
	if len(chunk.Data) <= 4 {
		return chunk, false
	}

	// Parse EXIF data (skip first 4 bytes)
	exifData, err := exif.Decode(bytes.NewReader(chunk.Data[4:]))
	if err != nil {
		// If we can't parse it, remove it to be safe
		return chunk, false
	}

	sensitiveTags := getSensitiveTagNames()
	hasAnyTag := false
	hasSensitiveTag := false
	
	// Check tags
	walker := &webpExifChecker{
		sensitiveTags:   sensitiveTags,
		hasAnyTag:       &hasAnyTag,
		hasSensitiveTag: &hasSensitiveTag,
	}
	
	exifData.Walk(walker)

	// If no non-sensitive tags, remove the entire chunk
	if !hasAnyTag {
		return chunk, false
	}

	// If any sensitive data found, remove chunk (TODO: implement selective preservation)
	if hasSensitiveTag {
		return chunk, false
	}

	return chunk, true
}

// webpExifCollector implements the exif.Walker interface for metadata preview
type webpExifCollector struct {
	sensitiveTags map[string]bool
	toRemove      *[]string
	toPreserve    *[]string
}

func (w *webpExifCollector) Walk(name exif.FieldName, tag *tiff.Tag) error {
	tagName := string(name)
	value := formatTagValue(tag)
	
	if w.sensitiveTags[tagName] {
		*w.toRemove = append(*w.toRemove, fmt.Sprintf("EXIF.%s = %s", tagName, value))
	} else {
		*w.toPreserve = append(*w.toPreserve, fmt.Sprintf("EXIF.%s = %s", tagName, value))
	}
	return nil
}

// webpExifChecker implements the exif.Walker interface for checking sensitive tags
type webpExifChecker struct {
	sensitiveTags   map[string]bool
	hasAnyTag       *bool
	hasSensitiveTag *bool
}

func (w *webpExifChecker) Walk(name exif.FieldName, tag *tiff.Tag) error {
	tagName := string(name)
	*w.hasAnyTag = true
	
	if w.sensitiveTags[tagName] {
		*w.hasSensitiveTag = true
	}
	return nil
}
