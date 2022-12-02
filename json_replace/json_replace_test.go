package json_replace

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test a single file with standard input
func TestSingleFile(t *testing.T) {
	inputPath := "../tests/case1/input.json"
	outputPath := "../tests/case1/output.json"
	outputExpectedPath := "../tests/case1/output_expected.json"
	rulePath := "../tests/case1/rules.json"

	cfg := NewDefaultConfig(inputPath, outputPath, rulePath)
	replace := NewJSONReplace(cfg)
	replace.Exec()

	actual, _ := os.ReadFile(outputPath)
	expected, _ := os.ReadFile(outputExpectedPath)
	if bytes.Compare(actual, expected) != 0 {
		t.Fatal("Test Case 1 Failed: Actual output and expected output do not match")
	}
}

// Test multiple files in a directory
func TestDirectory(t *testing.T) {
	inputPath := "../tests/case2/inputs"
	outputPath := "../tests/case2/outputs"
	outputExpectedPath := "../tests/case2/outputs_expected"
	rulePath := "../tests/case2/rules.json"

	cfg := NewDefaultConfig(inputPath, outputPath, rulePath)
	replace := NewJSONReplace(cfg)
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

// Test a single file in line-by-line mode
func TestLineByLine(t *testing.T) {
	inputPath := "../tests/case3/input.txt"
	outputPath := "../tests/case3/output.txt"
	outputExpectedPath := "../tests/case3/output_expected.txt"
	rulePath := "../tests/case3/rules.json"

	cfg := NewConfig(inputPath, outputPath, rulePath, true, 10)
	replace := NewJSONReplace(cfg)
	replace.Exec()

	actual, _ := os.ReadFile(outputPath)
	expected, _ := os.ReadFile(outputExpectedPath)
	if bytes.Compare(actual, expected) != 0 {
		t.Fatal("Test Case 3 Failed: Actual output and expected output do not match")
	}
}

// Test multiple files in a directory
func TestTimestamp(t *testing.T) {
	inputPath := "../tests/case5/inputs"
	outputPath := "../tests/case5/outputs"
	outputExpectedPath := "../tests/case5/outputs_expected"
	rulePath := "../tests/case5/rules.json"

	cfg := NewDefaultConfig(inputPath, outputPath, rulePath)
	replace := NewJSONReplace(cfg)
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
