package json_select

import (
	"flag"
	"path/filepath"
)

type Config struct {
	inputPath   string
	outputPath  string
	rulePath    string
	maxRoutines int
}

func NewConfig(inputPath string, outputPath string, rulePath string, maxRoutines int) *Config {
	// Clean paths to standard format
	inputPath = filepath.Clean(inputPath)
	outputPath = filepath.Clean(outputPath)

	c := Config{
		inputPath:   inputPath,
		outputPath:  outputPath,
		rulePath:    rulePath,
		maxRoutines: maxRoutines,
	}
	return &c
}

func NewDefaultConfig(inputPath string, outputPath string, rulePath string) *Config {
	return NewConfig(inputPath, outputPath, rulePath, 10)
}

func NewConfigFromConsole() *Config {
	// Config and parse flags
	inputPath := flag.String("i", "", "input path")
	outputPath := flag.String("o", "", "output path")
	rulePath := flag.String("r", "", "rule path")
	maxRoutines := flag.Int("n", 10, "maximum routines")

	flag.Parse()

	return NewConfig(*inputPath, *outputPath, *rulePath, *maxRoutines)
}
