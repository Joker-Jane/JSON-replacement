package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Rule struct represents a rule object
type Rule struct {
	Order       int     `json:"order"`
	Type        string  `json:"type"`
	FieldName   string  `json:"field-name"`
	Original    string  `json:"original"`
	Replacement string  `json:"replacement"`
	Duration    int64   `json:"duration"`
	MaxSamples  int64   `json:"max-records"`
	Time        float64 `json:"start-ms"`
	Index       int64
}

// Flags
var inputPath *string
var outputPath *string
var configPath *string
var lineByLine *bool
var maxRoutines *int

// The list of all rules
var rules []*Rule

// Assigned files
var assignCounter int

// Processed files
var fileCounter int

// Record start time
var startTime time.Time

// Lock for updating file counter
var lock sync.Mutex

/*
This program reads file(s) containing JSON records, and redacts or replaces the
private / client information in each based on some predefined parameters.

-i, -o, and -c flags must be specified.
-l and -r flags are optional.

Reading multiple JSON objects line-by-line is supported by specifying -l flag.
Note that a single JSON object in multiple lines is not supported if line-by-line mode is enabled.

The program is running in concurrency by default.

Usage:

	./json_replace [flags]

Flags:

	-i input_path
		Set the path to the input file or directory. Path to a file must be a json file.

	-o output_path
		Set the path to the out file or directory. Path to a file must be a json file.

	-c config_path
		Set the path to the config file. The file must be a json file.

	-l
		Read multiple json objects line by line. Default: false

	-r [number of routines]
		Set the maximum number of routines running simultaneously. Default: 10
*/
func main() {
	// Config and parse flags
	inputPath = flag.String("i", "", "input path")
	outputPath = flag.String("o", "", "output path")
	configPath = flag.String("c", "", "config path")
	lineByLine = flag.Bool("l", false, "line-by-line mode")
	maxRoutines = flag.Int("r", 10, "maximum routines")

	flag.Parse()

	// Clean paths to standard format
	*inputPath = filepath.Clean(*inputPath)
	*outputPath = filepath.Clean(*outputPath)

	// Record start time
	startTime = time.Now()

	// Check if all arguments are specified
	if *inputPath == "" || *configPath == "" || *outputPath == "" {
		log.Fatal("Usage: ./json_replace -i input -o output -c config [-l] [-r]")
	}

	if *maxRoutines <= 0 {
		log.Fatal("Error: Maximum number of routines must be greater than 0")
	}

	// Check if input path exists
	_, err := os.Stat(*inputPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatal("Error: Input path '" + *inputPath + "' not found")
		} else {
			log.Fatal("Error: Cannot read input path '" + *inputPath + "'")
		}
	}

	// Check if config file exists
	_, err = os.Stat(*configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatal("Error: Config file '" + *configPath + "' not found")
		} else {
			log.Fatal("Error: Cannot read config file '" + *configPath + "'")
		}
	}

	// Read config file
	rule, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatal("Error: Cannot read config file '" + *configPath + "'")
	}

	// Parse config file
	err = json.Unmarshal(rule, &rules)
	if err != nil {
		log.Fatal("Error: Config file must be in the format of arrays of rule json objects")
	}

	// Initiate start time for replay if not specified
	for _, r := range rules {
		if r.Time == 0 {
			r.Time = float64(time.Now().UnixMilli())
		}
	}

	// Limit the max number of goroutines running simultaneously
	ch := make(chan int, *maxRoutines)

	// Walk through and process the input file tree
	err = filepath.WalkDir(*inputPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			// Assign the file and start a routine if the buffer is not full
			assignCounter++
			ch <- 1
			go startRoutine(path, ch)
		}
		return err
	})
	if err != nil {
		log.Fatal("Error: Failed to walk through the input directory")
	}

	// Wait until all files are processed
	for assignCounter != fileCounter {
	}

	log.Printf("Success: Processed %d file(s) in %.4f second(s)\n", fileCounter, time.Since(startTime).Seconds())
}

// Start a goroutine
func startRoutine(filePath string, ch chan int) {
	handleFile(filePath)
	// Lock the fileCounter to ensure synchronization
	lock.Lock()
	fileCounter += <-ch
	lock.Unlock()
}

// Handle input json file
func handleFile(filePath string) {
	// Read input file
	input, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("Error: Cannot read input file '" + filePath + "'")
	}

	// Store the result
	var result []byte

	// If is in line-by-line mode, split the input by \n, process each line, and append to result
	// If not, store the result directly
	if *lineByLine {
		inputs := bytes.Split(input, []byte("\n"))
		for l, i := range inputs {
			r, err := handleJSON(i)
			if err != nil {
				log.Fatal("Error: Line " + strconv.Itoa(l+1) + " of '" + filePath + "' is not in valid JSON format")
			}
			r = append(r, byte('\n'))
			result = append(result, r...)
		}
	} else {
		result, err = handleJSON(input)
		if err != nil {
			log.Fatal("Error: File '" + filePath + "' is not in valid JSON format")
		}
	}

	// Get target output path
	target := strings.Replace(filePath, *inputPath, *outputPath, 1)

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
func handleJSON(input []byte) ([]byte, error) {
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
	for _, r := range rules {
		switch r.Type {
		case "per-field":
			process("", m, r.Original, r.Replacement, r.FieldName, false)
		case "global":
			process("", m, r.Original, r.Replacement, r.FieldName, true)
		case "timestamp":
			processReplay("", m, r.FieldName, r.Duration, r.MaxSamples, &r.Time, &r.Index)
		default:
			log.Fatal("Error: Invalid type '" + r.Type + "'")
		}
	}

	// Write file to output
	result, _ := json.Marshal(m)
	return result, nil
}

// Process non-string elements
func process(k string, v interface{}, from string, to string, field string, isGlobal bool) {
	switch v.(type) {
	case map[string]interface{}:
		processMap(v.(map[string]interface{}), from, to, field, isGlobal)
	case []interface{}:
		processArray(v.([]interface{}), k, from, to, field, isGlobal)
	}
}

// Process maps
func processMap(m map[string]interface{}, from string, to string, field string, isGlobal bool) {
	// If global rule applies, iterate every element in the map
	// If not, check if the particular field exists
	if isGlobal {
		for k, v := range m {
			switch v.(type) {
			case string:
				m[k] = strings.Replace(v.(string), from, to, -1)
			default:
				process(k, v, from, to, "", isGlobal)
			}
		}
	} else {
		k, next, _ := strings.Cut(field, ".")
		v, found := m[k]
		if found {
			switch v.(type) {
			case string:
				if next == "" {
					m[k] = strings.Replace(v.(string), from, to, -1)
				}
			default:
				process(k, v, from, to, next, isGlobal)
			}
		}
	}
}

// Process arrays
func processArray(a []interface{}, k string, from string, to string, field string, isGlobal bool) {
	for i, v := range a {
		switch v.(type) {
		case string:
			if k == "" {
				a[i] = strings.Replace(v.(string), from, to, -1)
			}
		default:
			process(k, v, from, to, field, isGlobal)
		}
	}
}

// Process replay
func processReplay(k string, v interface{}, field string, duration int64, samples int64, time *float64, index *int64) {
	switch v.(type) {
	case map[string]interface{}:
		processReplayMap(v.(map[string]interface{}), field, duration, samples, time, index)
	case []interface{}:
		processReplayArray(v.([]interface{}), k, field, duration, samples, time, index)
	}
}

// Process replay maps
func processReplayMap(m map[string]interface{}, field string, duration int64, samples int64, time *float64, index *int64) {
	k, next, _ := strings.Cut(field, ".")
	if next == "" {
		lock.Lock()
		cur := *time + calculateIncrement(*index, duration, samples)
		m[k] = int64(cur)
		*time = cur
		*index++
		lock.Unlock()
	} else {
		for k, v := range m {
			processReplay(k, v, field, duration, samples, time, index)
		}
	}
}

// Process replay arrays
func processReplayArray(a []interface{}, k string, field string, duration int64, samples int64, time *float64, index *int64) {
	for _, v := range a {
		processReplay(k, v, field, duration, samples, time, index)
	}
}

// Calculate the increment of a record by integration
func calculateIncrement(i int64, duration int64, samples int64) float64 {
	var k = float64(samples) / float64(duration)
	var fa = float64(i-1) * k
	var fb = float64(i) * k
	return fb - fa
}
