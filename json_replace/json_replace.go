/*
This program reads file(s) containing JSON records, and redacts or replaces the
private / client information in each based on some predefined parameters.

-i, -o, and -r flags must be specified.
-l and -n flags are optional.

The input path and output path can be either a file or a directory.
The rule path must be a JSON file in valid rule format.

Reading multiple JSON objects line-by-line is supported by specifying -l flag.
Note that a single JSON object in multiple lines is not supported if line-by-line mode is enabled.

The program is running concurrently by default.
This can be disabled by setting -n flag to 1.

Usage:

	./json_replace [flags]

Flags:

	-i input_path
		Set the path to the input file or directory.

	-o output_path
		Set the path to the output file or directory.

	-r rule_path
		Set the path to the rule file.

	-l
		Read multiple json objects line by line. Default: false

	-n [number of routines]
		Set the maximum number of routines running simultaneously. Default: 10
*/
package json_replace

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

// JSONReplace struct represents a JSONReplace object
type JSONReplace struct {
	// Configs
	config *Config

	// The list of all rules
	rules []*Rule

	// Synchronization
	sync *Sync
}

// Rule struct represents a rule object
type Rule struct {
	Order       int    `json:"order"`
	Type        string `json:"type"`
	FieldName   string `json:"field-name"`
	Original    string `json:"original"`
	Replacement string `json:"replacement"`
	Duration    int64  `json:"duration"`
	MaxRecords  int64  `json:"max-records"`
	StartMs     int64  `json:"start-ms"`
	replay      Replay
}

// Replay struct records replay related fields
type Replay struct {
	time    float64
	index   int64
	records int64
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

// Create a JSONReplace Object
func NewJSONReplace(config *Config) *JSONReplace {
	// Check if all arguments are specified
	if config.inputPath == "" || config.rulePath == "" || config.outputPath == "" {
		log.Fatal("Usage: ./json_replace -i input -o output -r rule [-l] [-n routines]")
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

	// Sort the rules by order
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Order < rules[j].Order
	})

	// Construct JSONReplace object
	replace := &JSONReplace{
		config: config,
		rules:  rules,
		sync:   new(Sync),
	}

	return replace
}

// Execute
func (replace *JSONReplace) Exec() {
	// Record start time
	startTime := time.Now()

	// Initiate time for replay
	for _, r := range replace.rules {
		if r.StartMs == 0 {
			r.replay.time = float64(time.Now().UnixMilli())
		} else {
			r.replay.time = float64(r.StartMs)
		}
		r.replay.records = r.MaxRecords
	}

	// Limit the max number of goroutines running simultaneously
	ch := make(chan int, replace.config.maxRoutines)

	// Walk through and process the input file tree
	err := filepath.WalkDir(replace.config.inputPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			// Assign the file and start a routine if the buffer is not full
			replace.sync.assignCounter++
			ch <- 1
			go replace.startRoutine(path, ch)
		}
		return err
	})
	if err != nil {
		log.Fatal("Error: Failed to walk through the input directory")
	}

	// Wait until all files are processed
	for replace.sync.assignCounter != replace.sync.processCounter {
	}

	// Log output
	log.Printf("Success: Processed %d file(s) in %.4f second(s)\n",
		replace.sync.processCounter, time.Since(startTime).Seconds())
}

// Start a goroutine
func (replace *JSONReplace) startRoutine(filePath string, ch chan int) {
	replace.handleFile(filePath)

	// Lock the processCounter to ensure synchronization
	replace.sync.lock.Lock()
	defer replace.sync.lock.Unlock()
	replace.sync.processCounter += <-ch
}

// Handle input json file
func (replace *JSONReplace) handleFile(filePath string) {
	// Read input file
	input, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("Error: Cannot read input file '" + filePath + "'")
	}

	// Store the result
	var result []byte

	// If is in line-by-line mode, split the input by \n, process each line, and append to result
	// If not, store the result directly
	if replace.config.lineByLine {
		inputs := bytes.Split(input, []byte("\n"))
		for l, i := range inputs {
			r, err := replace.handleJSON(i)
			if err != nil {
				log.Fatal("Error: Line " + strconv.Itoa(l+1) + " of '" + filePath + "' is not in valid JSON format")
			}
			r = append(r, byte('\n'))
			result = append(result, r...)
		}
	} else {
		result, err = replace.handleJSON(input)
		if err != nil {
			log.Fatal("Error: File '" + filePath + "' is not in valid JSON format")
		}
	}

	// Get target output path
	target := strings.Replace(filePath, replace.config.inputPath, replace.config.outputPath, 1)

	// Get parent directory of the target
	dir, _ := filepath.Split(target)

	// Create the directory if the file is not in root
	if dir != "" {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			log.Fatal("Error: Failed to create directory '" + dir + "'")
		}
	}

	// Write to target file
	err = os.WriteFile(target, result, 0666)
	if err != nil {
		log.Fatal("Error: Cannot write to '" + target + "'")
	}
}

// Handle a single JSON object
func (replace *JSONReplace) handleJSON(input []byte) ([]byte, error) {
	// Return if the input is empty
	if len(input) == 0 {
		return nil, nil
	}

	// Parse input file
	var m interface{}
	err := json.Unmarshal(input, &m)
	if err != nil {
		return nil, err
	}

	// Apply every rule on files
	for _, r := range replace.rules {
		switch r.Type {
		case "per-field":
			replace.process("", m, *r)
		case "global":
			replace.process("", m, *r)
		case "timestamp":
			replace.processReplay("", m, r)
		default:
			log.Fatal("Error: Invalid type '" + r.Type + "'")
		}
	}

	// Write file to output
	result, _ := json.Marshal(m)
	return result, nil
}

// Process non-string elements
func (replace *JSONReplace) process(k string, v interface{}, r Rule) {
	switch v.(type) {
	case map[string]interface{}:
		replace.processMap(v.(map[string]interface{}), r)
	case []interface{}:
		replace.processArray(v.([]interface{}), k, r)
	}
}

// Process maps
func (replace *JSONReplace) processMap(m map[string]interface{}, r Rule) {
	// If global rule applies, iterate every element in the map
	// If not, check if the particular field exists
	if r.Type == "global" {
		for k, v := range m {
			switch v.(type) {
			case string:
				m[k] = strings.Replace(v.(string), r.Original, r.Replacement, -1)
			default:
				replace.process(k, v, r)
			}
		}
	} else {
		k, next, _ := strings.Cut(r.FieldName, ".")
		v, found := m[k]
		if found {
			switch v.(type) {
			case string:
				if next == "" {
					m[k] = strings.Replace(v.(string), r.Original, r.Replacement, -1)
				}
			default:
				r.FieldName = next
				replace.process(k, v, r)
			}
		}
	}
}

// Process arrays
func (replace *JSONReplace) processArray(a []interface{}, k string, r Rule) {
	for i, v := range a {
		switch v.(type) {
		case string:
			if k == "" {
				a[i] = strings.Replace(v.(string), r.Original, r.FieldName, -1)
			}
		default:
			replace.process(k, v, r)
		}
	}
}

// Process replay
func (replace *JSONReplace) processReplay(k string, v interface{}, r *Rule) {
	switch v.(type) {
	case map[string]interface{}:
		replace.processReplayMap(v.(map[string]interface{}), r)
	case []interface{}:
		replace.processReplayArray(v.([]interface{}), k, r)
	}
}

// Process replay maps
func (replace *JSONReplace) processReplayMap(m map[string]interface{}, r *Rule) {
	k, next, _ := strings.Cut(r.FieldName, ".")
	if next == "" {
		replace.sync.lock.Lock()
		defer replace.sync.lock.Unlock()
		cur := r.replay.time + replace.calculateIncrement(r.replay.index, r.Duration, r.replay.records)
		m[k] = int64(cur)
		r.replay.time = cur
		r.replay.index++
	} else {
		for k, v := range m {
			replace.processReplay(k, v, r)
		}
	}
}

// Process replay arrays
func (replace *JSONReplace) processReplayArray(a []interface{}, k string, r *Rule) {
	for _, v := range a {
		replace.processReplay(k, v, r)
	}
}

// Calculate the increment of a record by integration
func (replace *JSONReplace) calculateIncrement(i int64, duration int64, records int64) float64 {
	var k = float64(duration) / float64(records)
	var fa = float64(i-1) * k
	var fb = float64(i) * k
	return fb - fa
}
