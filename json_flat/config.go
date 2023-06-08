package json_flat

import (
	"flag"
	"path/filepath"
)

type Config struct {
	inputPath  string
	outputPath string
}

func NewConfig(inputPath string, outputPath string) *Config {
	// Clean paths to standard format
	inputPath = filepath.Clean(inputPath)
	outputPath = filepath.Clean(outputPath)

	c := Config{
		inputPath:  inputPath,
		outputPath: outputPath,
	}
	return &c
}

func NewDefaultConfig(inputPath string, outputPath string) *Config {
	return NewConfig(inputPath, outputPath)
}

func NewConfigFromConsole() *Config {
	// Config and parse flags
	inputPath := flag.String("i", "", "input path")
	outputPath := flag.String("o", "", "output path")

	flag.Parse()

	return NewConfig(*inputPath, *outputPath)
}
