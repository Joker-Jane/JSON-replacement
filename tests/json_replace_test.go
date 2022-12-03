package tests

import (
	"bytes"
	"github.com/Joker-Jane/JSON-replacement/json_replace"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test a single file with standard input
func TestSingleFile(t *testing.T) {
	inputPath := "case1/input.json"
	outputPath := "case1/output.json"
	outputExpectedPath := "case1/output_expected.json"
	rulePath := "case1/rules.json"

	cfg := json_replace.NewDefaultConfig(inputPath, outputPath, rulePath)
	replace := json_replace.NewJSONReplace(cfg)
	replace.Exec()

	actual, _ := os.ReadFile(outputPath)
	expected, _ := os.ReadFile(outputExpectedPath)
	if bytes.Compare(actual, expected) != 0 {
		t.Fatal("Test Case 1 Failed: Actual output and expected output do not match")
	}
}

// Test multiple files in a directory
func TestDirectory(t *testing.T) {
	inputPath := "case2/inputs"
	outputPath := "case2/outputs"
	outputExpectedPath := "case2/outputs_expected"
	rulePath := "case2/rules.json"

	cfg := json_replace.NewDefaultConfig(inputPath, outputPath, rulePath)
	replace := json_replace.NewJSONReplace(cfg)
	replace.Exec()

	_ = filepath.WalkDir(outputPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			actual, _ := os.ReadFile(path)
			target := strings.Replace(path, outputPath, outputExpectedPath, 1)
			expected, _ := os.ReadFile(target)
			if bytes.Compare(actual, expected) != 0 {
				t.Fatal("Test Case 2 Failed: Actual output and expected output do not match")
			}
		}
		return err
	})
}

/*
// Test massive multiple files in a directory
func TestMassive(t *testing.T) {
	inputPath := "case4/inputs10000"
	outputPath := "case4/outputs"
	rulePath := "case4/rules.json"

	cfg := json_replace.NewDefaultConfig(inputPath, outputPath, rulePath)
	replace := json_replace.NewJSONReplace(cfg)
	replace.Exec()
}
*/

// Test a single file in line-by-line mode
func TestLineByLine(t *testing.T) {
	inputPath := "case3/input.txt"
	outputPath := "case3/output.txt"
	outputExpectedPath := "case3/output_expected.txt"
	rulePath := "case3/rules.json"

	cfg := json_replace.NewConfig(inputPath, outputPath, rulePath, true, 10)
	replace := json_replace.NewJSONReplace(cfg)
	replace.Exec()

	actual, _ := os.ReadFile(outputPath)
	expected, _ := os.ReadFile(outputExpectedPath)
	if bytes.Compare(actual, expected) != 0 {
		t.Fatal("Test Case 3 Failed: Actual output and expected output do not match")
	}
}

// Test multiple files in a directory
func TestTimestamp(t *testing.T) {
	inputPath := "case5/inputs"
	outputPath := "case5/outputs"
	outputExpectedPath := "case5/outputs_expected"
	rulePath := "case5/rules.json"

	cfg := json_replace.NewDefaultConfig(inputPath, outputPath, rulePath)
	replace := json_replace.NewJSONReplace(cfg)
	replace.Exec()

	_ = filepath.WalkDir(outputPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			actual, _ := os.ReadFile(path)
			target := strings.Replace(path, outputPath, outputExpectedPath, 1)
			expected, _ := os.ReadFile(target)
			if bytes.Compare(actual, expected) != 0 {
				t.Fatal("Test Case 5 Failed: Actual output and expected output do not match")
			}
		}
		return err
	})
}
