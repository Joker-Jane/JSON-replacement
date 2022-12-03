package json_replace

import (
	"flag"
	"path/filepath"
)

type Config struct {
	inputPath   string
	outputPath  string
	rulePath    string
	lineByLine  bool
	maxRoutines int
}

func NewConfig(inputPath string, outputPath string, rulePath string, lineByline bool, maxRoutines int) *Config {
	// Clean paths to standard format
	inputPath = filepath.Clean(inputPath)
	outputPath = filepath.Clean(outputPath)

	c := Config{
		inputPath:   inputPath,
		outputPath:  outputPath,
		rulePath:    rulePath,
		lineByLine:  lineByline,
		maxRoutines: maxRoutines,
	}
	return &c
}

func NewDefaultConfig(inputPath string, outputPath string, rulePath string) *Config {
	return NewConfig(inputPath, outputPath, rulePath, false, 10)
}

func NewConfigFromConsole() *Config {
	// Config and parse flags
	inputPath := flag.String("i", "", "input path")
	outputPath := flag.String("o", "", "output path")
	rulePath := flag.String("c", "", "config path")
	lineByLine := flag.Bool("l", false, "line-by-line mode")
	maxRoutines := flag.Int("r", 10, "maximum routines")

	flag.Parse()

	return NewConfig(*inputPath, *outputPath, *rulePath, *lineByLine, *maxRoutines)
}
