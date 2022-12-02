package json_replace

import (
	"bytes"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test a single file with standard input
func TestSingleFile(t *testing.T) {
	inputPath := "tests/case1/input.json"
	outputPath := "tests/case1/output.json"
	outputExpectedPath := "tests/case1/output_expected.json"
	configPath := "tests/case1/rules.json"
	os.Args = []string{"json_replace", "-i", inputPath, "-o", outputPath, "-c", configPath}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	main()

	actual, _ := os.ReadFile(outputPath)
	expected, _ := os.ReadFile(outputExpectedPath)
	if bytes.Compare(actual, expected) != 0 {
		t.Fatal("Test Case 1 Failed: Actual output and expected output do not match")
	}
}

// Test multiple files in a directory
func TestDirectory(t *testing.T) {
	inputPath := "tests/case2/inputs"
	outputPath := "tests/case2/outputs"
	outputExpectedPath := "tests/case2/outputs_expected"
	configPath := "tests/case2/rules.json"
	os.Args = []string{"json_replace", "-i", inputPath, "-o", outputPath, "-c", configPath}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	main()

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
	inputPath := "tests/case3/input.txt"
	outputPath := "tests/case3/output.txt"
	outputExpectedPath := "tests/case3/output_expected.txt"
	configPath := "tests/case3/rules.json"
	os.Args = []string{"json_replace", "-i", inputPath, "-o", outputPath, "-c", configPath, "-l"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	main()

	actual, _ := os.ReadFile(outputPath)
	expected, _ := os.ReadFile(outputExpectedPath)
	if bytes.Compare(actual, expected) != 0 {
		t.Fatal("Test Case 3 Failed: Actual output and expected output do not match")
	}
}

// Test multiple files in a directory
func TestTimestamp(t *testing.T) {
	inputPath := "tests/case5/inputs"
	outputPath := "tests/case5/outputs"
	outputExpectedPath := "tests/case5/outputs_expected"
	configPath := "tests/case5/rules.json"
	os.Args = []string{"json_replace", "-i", inputPath, "-o", outputPath, "-c", configPath}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	main()

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
