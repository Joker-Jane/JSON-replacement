package tests

import (
	"github.com/Joker-Jane/JSON-replacement/json_flat"
	"testing"
)

func TestFlatSingleFile(t *testing.T) {
	inputPath := "json_flat_tests/case1/input.json"
	outputPath := "json_flat_tests/case1/output.json"

	cfg := json_flat.NewDefaultConfig(inputPath, outputPath)
	flat := json_flat.NewJSONFlat(cfg)
	flat.Exec()
}

func TestFlatSingleFileWithMultipleLines(t *testing.T) {
	inputPath := "json_flat_tests/case2/input.json"
	outputPath := "json_flat_tests/case2/output.json"

	cfg := json_flat.NewDefaultConfig(inputPath, outputPath)
	flat := json_flat.NewJSONFlat(cfg)
	flat.Exec()
}

func TestFlatMultipleFiles(t *testing.T) {
	inputPath := "json_flat_tests/case3/inputs"
	outputPath := "json_flat_tests/case3/outputs"

	cfg := json_flat.NewDefaultConfig(inputPath, outputPath)
	flat := json_flat.NewJSONFlat(cfg)
	flat.Exec()
}

func TestComplex(t *testing.T) {
	inputPath := "json_flat_tests/case4/inputs"
	outputPath := "json_flat_tests/case4/outputs"

	cfg := json_flat.NewDefaultConfig(inputPath, outputPath)
	flat := json_flat.NewJSONFlat(cfg)
	flat.Exec()
}
