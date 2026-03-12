package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/image-metadata-remover/internal/fileutil"
	"github.com/yourusername/image-metadata-remover/internal/processor"
)

const (
	version = "0.1.0"
	appName = "imr"
)

type Config struct {
	showVersion bool
	showHelp    bool
	files       []string
	suffix      string
	verbose     bool
}

func main() {
	config := parseFlags()

	if config.showVersion {
		printVersion()
		return
	}

	if config.showHelp || len(config.files) == 0 {
		printUsage()
		return
	}

	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() Config {
	config := Config{}

	flag.BoolVar(&config.showVersion, "version", false, "Show version information")
	flag.BoolVar(&config.showVersion, "v", false, "Show version information (shorthand)")
	flag.BoolVar(&config.showHelp, "help", false, "Show help message")
	flag.BoolVar(&config.showHelp, "h", false, "Show help message (shorthand)")
	flag.StringVar(&config.suffix, "suffix", "_imr", "Suffix to append to output filenames")
	flag.BoolVar(&config.verbose, "verbose", false, "Enable verbose output")

	flag.Parse()

	config.files = flag.Args()

	return config
}

func printVersion() {
	fmt.Printf("%s version %s\n", appName, version)
}

func printUsage() {
	fmt.Printf("Usage: %s [OPTIONS] <image_files...>\n\n", appName)
	fmt.Println("Image Metadata Remover - Remove sensitive metadata from images")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Supported formats:")
	fmt.Println("  JPEG, PNG, WebP, TIFF, GIF")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s photo.jpg\n", appName)
	fmt.Printf("  %s -suffix _clean photo1.jpg photo2.png\n", appName)
	fmt.Printf("  %s -verbose images/*.jpg\n", appName)
	fmt.Println()
	fmt.Println("Output files will be saved with the specified suffix (default: _imr)")
	fmt.Println("Example: photo.jpg -> photo_imr.jpg")
}

func run(config Config) error {
	if config.verbose {
		fmt.Printf("Processing %d file(s)...\n", len(config.files))
		fmt.Printf("Output suffix: %s\n", config.suffix)
	}

	// Create processor
	proc := processor.NewProcessor()

	// Process each file
	successCount := 0
	errorCount := 0

	for _, file := range config.files {
		if err := processFile(file, config, proc); err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: %v\n", file, err)
			errorCount++
		} else {
			successCount++
		}
	}

	if config.verbose {
		fmt.Printf("\nProcessing complete: %d succeeded, %d failed\n", successCount, errorCount)
	}

	if errorCount > 0 {
		return fmt.Errorf("%d file(s) failed to process", errorCount)
	}

	return nil
}

func processFile(filename string, config Config, proc *processor.Processor) error {
	if config.verbose {
		fmt.Printf("Processing: %s\n", filename)
	}

	// Check if file exists
	if !fileutil.FileExists(filename) {
		return fmt.Errorf("file not found")
	}

	// Generate output path
	outputPath := fileutil.GenerateOutputPath(filename, config.suffix)

	// Check if output file already exists
	if fileutil.FileExists(outputPath) {
		return fmt.Errorf("output file already exists: %s", outputPath)
	}

	// Process the file
	if err := proc.ProcessFile(filename, outputPath); err != nil {
		return err
	}

	fmt.Printf("✓ %s -> %s\n", filename, outputPath)
	
	return nil
}
