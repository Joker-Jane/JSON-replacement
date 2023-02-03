/*
This program reads file(s) containing JSON records, and sort or redirect these
records to output file(s) based on some predefined parameters.

The records should be passed in the format of one line per JSON record.

The input path can be either a file or a directory.
The output path must be a directory.
The rule path must be a JSON file that contains an array of valid rule JSONs.

-i, -o, and -c flags must be specified.
-n flags is optional.

The program is running concurrently by default.
This can be disabled by setting -n flag to 1.

Usage:

	./json_select [flags]

Flags:

	-i input_path
		Set the path to the input file or directory.

	-o output_path
		Set the path to the output directory.

	-r rule_path
		Set the path to the rule file.

	-n [number of routines]
		Set the maximum number of routines running simultaneously. Default: 10
*/
package json_select

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// JSONSelect struct represents a JSONSelect object
type JSONSelect struct {
	// Configs
	config *Config

	// The list of all rules
	rules []*Rule

	// Synchronization
	sync *Sync
}

// Rule struct represents a rule object
type Rule struct {
	Position   int          `json:"position"`
	Output     string       `json:"output"`
	Conditions []*Condition `json:"conditions"`
}

type Condition struct {
	Type    string   `json:"type"`
	Key     string   `json:"key"`
	Values  []string `json:"values"`
	Exclude bool     `json:"exclude"`
}

// Sync struct ensures synchronization
type Sync struct {
	// Assigned files
	assignCounter int

	// Processed files
	processCounter int

	// Lock for updating file counter
	lock sync.Mutex
}

// Create a NewJSONSelect Object
func NewJSONSelect(config *Config) *JSONSelect {
	// Check if all arguments are specified
	if config.inputPath == "" || config.rulePath == "" || config.outputPath == "" {
		log.Fatal("Usage: ./json_select -i input -o output -r rule [-n routines]")
	}

	// Check if max routines is positive
	if config.maxRoutines <= 0 {
		log.Fatal("Error: Maximum number of routines must be greater than 0")
	}

	// Check if input path exists
	_, err := os.Stat(config.inputPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatal("Error: Input path '" + config.inputPath + "' not found")
		} else {
			log.Fatal("Error: Cannot read input path '" + config.inputPath + "'")
		}
	}

	// Check if config file exists
	_, err = os.Stat(config.rulePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatal("Error: Config file '" + config.rulePath + "' not found")
		} else {
			log.Fatal("Error: Cannot read rule file '" + config.rulePath + "'")
		}
	}

	// Read config file
	rule, err := os.ReadFile(config.rulePath)
	if err != nil {
		log.Fatal("Error: Cannot read rule file '" + config.rulePath + "'")
	}

	// Parse config file and store to rules
	var rules []*Rule
	err = json.Unmarshal(rule, &rules)
	if err != nil {
		log.Fatal("Error: Rule file must be in the format of arrays of rule json objects")
	}

	// Sort the rules by position
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Position < rules[j].Position
	})

	// Construct JSONSelect object
	s := &JSONSelect{
		config: config,
		rules:  rules,
		sync:   new(Sync),
	}

	return s
}

// Execute
func (s *JSONSelect) Exec() {
	// Record start time
	startTime := time.Now()

	// Limit the max number of goroutines running simultaneously
	ch := make(chan int, s.config.maxRoutines)

	// Walk through and process the input file tree
	err := filepath.WalkDir(s.config.inputPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			// Assign the file and start a routine if the buffer is not full
			s.sync.assignCounter++
			ch <- 1
			go s.startRoutine(path, ch)
		}
		return err
	})
	if err != nil {
		log.Fatal("Error: Failed to walk through the input directory")
	}

	// Wait until all files are processed
	for s.sync.assignCounter != s.sync.processCounter {
	}

	// Log output
	log.Printf("Success: Processed %d file(s) in %.4f second(s)\n",
		s.sync.processCounter, time.Since(startTime).Seconds())
}

// Start a goroutine
func (s *JSONSelect) startRoutine(filePath string, ch chan int) {
	s.handleFile(filePath)

	// Lock the processCounter to ensure synchronization
	s.sync.lock.Lock()
	defer s.sync.lock.Unlock()
	s.sync.processCounter += <-ch
}

// Handle input json file
func (s *JSONSelect) handleFile(filePath string) {
	// Read input file
	input, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("Error: Cannot read input file '" + filePath + "'")
	}

	// split the input by \n, process each line, and append to result
	inputs := bytes.Split(input, []byte("\n"))
	for l, i := range inputs {
		err := s.handleJSON(filePath, i)
		if err != nil {
			log.Fatal("Error: Line " + strconv.Itoa(l+1) + " of '" + filePath + "' is not in valid JSON format")
		}
	}
}

// Handle a single JSON object
func (s *JSONSelect) handleJSON(filePath string, input []byte) error {
	// Return if the input is empty
	if len(input) == 0 {
		return nil
	}

	// Parse input json
	var v interface{}
	err := json.Unmarshal(input, &v)
	if err != nil {
		return err
	}

	// Apply every rule on files, stop if match any rule
	for _, r := range s.rules {
		// If all rules are met, write to specific output
		if s.processRule(v, *r) {
			s.write(filePath, &input, r.Output)
			return nil
		}
	}

	// If no rule is met, send to drop
	s.write(filePath, &input, "drop")
	return nil
}

// Return if all conditions in the rule is met
func (s *JSONSelect) processRule(v interface{}, r Rule) bool {
	for _, c := range r.Conditions {
		if !s.processCondition(v, *c) {
			return false
		}
	}
	return true
}

// Return if the condition is met
func (s *JSONSelect) processCondition(v interface{}, c Condition) bool {
	return s.process("", v, c) != c.Exclude
}

// Process non-string elements
func (s *JSONSelect) process(k string, v interface{}, c Condition) bool {
	switch v.(type) {
	case map[string]interface{}:
		return s.processMap(v.(map[string]interface{}), c)
	case []interface{}:
		return s.processArray(v.([]interface{}), k, c)
	}
	return false
}

// Process maps
func (s *JSONSelect) processMap(m map[string]interface{}, c Condition) bool {
	k, next, _ := strings.Cut(c.Key, ".")
	v, found := m[k]
	if found {
		switch v.(type) {
		case string:
			if next == "" {
				for _, value := range c.Values {
					if v.(string) == value {
						return true
					}
				}
				return false
			}
		default:
			c.Key = next
			return s.process(k, v, c)
		}
	}
	return false
}

// Process arrays
func (s *JSONSelect) processArray(a []interface{}, k string, c Condition) bool {
	for _, v := range a {
		switch v.(type) {
		case string:
			if k == "" {
				for _, value := range c.Values {
					if v.(string) == value {
						return true
					}
				}
				return false
			}
		default:
			return s.process(k, v, c)
		}
	}
	return false
}

func (s *JSONSelect) write(filePath string, json *[]byte, output string) {
	// Get target output path
	target := strings.Replace(filePath, s.config.inputPath, s.config.outputPath, 1)

	err := os.MkdirAll(target, 0700)
	if err != nil {
		log.Fatal("Error: Failed to create directory '" + target + "'")
	}

	if output == "" {
		output = "default"
	}

	// Get directory of output file
	target += string(filepath.Separator) + output

	f, err := os.OpenFile(target, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	*json = append(*json, byte('\n'))

	if _, err := f.Write(*json); err != nil {
		log.Fatal(err)
	}
}
