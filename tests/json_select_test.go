package tests

import (
	"github.com/Joker-Jane/JSON-replacement/json_select"
	"testing"
)

// Test a simple input with standard input
func TestSelectMatch01(t *testing.T) {
	inputPath := "json_select_tests/case1/input"
	outputPath := "json_select_tests/case1/output"
	rulePath := "json_select_tests/case1/rules.json"

	cfg := json_select.NewDefaultConfig(inputPath, outputPath, rulePath)
	s := json_select.NewJSONSelect(cfg)
	s.Exec()
}

// Test another simple input with standard input
func TestSelectMatch02(t *testing.T) {
	inputPath := "json_select_tests/case2/input"
	outputPath := "json_select_tests/case2/output"
	rulePath := "json_select_tests/case2/rules.json"

	cfg := json_select.NewDefaultConfig(inputPath, outputPath, rulePath)
	s := json_select.NewJSONSelect(cfg)
	s.Exec()
}
