package tests

import (
	"github.com/Joker-Jane/JSON-replacement/json_replace"
	"testing"
)

// Test a single file with standard input
func TestReplaceSingleFile(t *testing.T) {
	inputPath := "json_replace_tests/case1/input.json"
	outputPath := "json_replace_tests/case1/output.json"
	rulePath := "json_replace_tests/case1/rules.json"

	cfg := json_replace.NewDefaultConfig(inputPath, outputPath, rulePath)
	replace := json_replace.NewJSONReplace(cfg)
	replace.Exec()
}

// Test multiple files in a directory
func TestReplaceDirectory(t *testing.T) {
	inputPath := "json_replace_tests/case2/inputs"
	outputPath := "json_replace_tests/case2/outputs"
	rulePath := "json_replace_tests/case2/rules.json"

	cfg := json_replace.NewDefaultConfig(inputPath, outputPath, rulePath)
	replace := json_replace.NewJSONReplace(cfg)
	replace.Exec()
}

/*
// Test massive multiple files in a directory
func TestReplaceMassive(t *testing.T) {
	inputPath := "json_replace_tests/case4/inputs10000"
	outputPath := "json_replace_tests/case4/outputs"
	rulePath := "json_replace_tests/case4/rules.json"

	cfg := json_replace.NewDefaultConfig(inputPath, outputPath, rulePath)
	replace := json_replace.NewJSONReplace(cfg)
	replace.Exec()
}
*/

// Test a single file in line-by-line mode
func TestReplaceLineByLine(t *testing.T) {
	inputPath := "json_replace_tests/case3/input.txt"
	outputPath := "json_replace_tests/case3/output.txt"
	rulePath := "json_replace_tests/case3/rules.json"

	cfg := json_replace.NewConfig(inputPath, outputPath, rulePath, true, 10)
	replace := json_replace.NewJSONReplace(cfg)
	replace.Exec()
}

// Test multiple files in a directory
func TestReplaceTimestamp(t *testing.T) {
	inputPath := "json_replace_tests/case5/inputs"
	outputPath := "json_replace_tests/case5/outputs"
	rulePath := "json_replace_tests/case5/rules.json"

	cfg := json_replace.NewDefaultConfig(inputPath, outputPath, rulePath)
	replace := json_replace.NewJSONReplace(cfg)
	replace.Exec()
}
