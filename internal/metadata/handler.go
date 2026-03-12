package metadata

// Handler is the interface that all format-specific metadata handlers must implement
type Handler interface {
	// RemoveMetadata removes sensitive metadata from the image file
	// Returns the path to the output file and any error
	RemoveMetadata(inputPath, outputPath string) error
	
	// SupportsFormat returns true if this handler supports the given format
	SupportsFormat(format string) bool
}

// MetadataType represents different types of metadata
type MetadataType string

const (
	// Sensitive metadata to be removed
	TypeGPS          MetadataType = "gps"
	TypeCamera       MetadataType = "camera"
	TypeLens         MetadataType = "lens"
	TypeSerialNumber MetadataType = "serial"
	TypeSoftware     MetadataType = "software"
	TypeCreator      MetadataType = "creator"
	TypeCopyright    MetadataType = "copyright"
	
	// Basic metadata to preserve
	TypeDimensions   MetadataType = "dimensions"
	TypeColorSpace   MetadataType = "colorspace"
	TypeOrientation  MetadataType = "orientation"
)
