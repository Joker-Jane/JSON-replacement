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
	"bufio"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
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

	// Store file pointers to output files
	outputMap *map[string]*os.File
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
		config:    config,
		rules:     rules,
		outputMap: &map[string]*os.File{},
	}

	return s
}

// Create output files from rules and store file pointers to a map
func (s *JSONSelect) CreateOutputFiles() {
	err := os.MkdirAll(s.config.outputPath, 0700)
	if err != nil {
		log.Fatal("Error: Failed to create directory '" + s.config.outputPath + "'")
	}

	s.CreateOutputFile("default")
	s.CreateOutputFile("drop")
	for _, r := range s.rules {
		s.CreateOutputFile(r.Output)
	}
}

// Create a single output file
func (s *JSONSelect) CreateOutputFile(output string) {
	if (*s.outputMap)[output] == nil {
		p := filepath.Join(s.config.outputPath, output)
		f, err := os.Create(p)
		if err != nil {
			log.Fatal("Error: Failed to create file '" + p + "'")
		}
		(*s.outputMap)[output] = f
	}
}

// Close output files
func (s *JSONSelect) CloseOutputFiles() {
	for output, f := range *s.outputMap {
		err := f.Close()
		if err != nil {
			log.Fatal("Error: Failed to close file '" +
				filepath.Join(s.config.outputPath, output) + "'")
		}
	}
}

// Execute
func (s *JSONSelect) Exec() {
	// Record start time
	startTime := time.Now()

	// Record record count
	count := 0

	// Create outputs files
	s.CreateOutputFiles()

	// Limit the max number of goroutines running simultaneously
	ch := make(chan int, s.config.maxRoutines)

	// Handle synchronization
	var wg sync.WaitGroup

	// Walk through and process the input file tree
	err := filepath.WalkDir(s.config.inputPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			count += s.handleFile(path, ch, &wg)
		}
		return err
	})
	if err != nil {
		log.Fatal("Error: Failed to walk through the input directory")
	}

	// Wait until all routines finish
	wg.Wait()

	// Close output files
	s.CloseOutputFiles()

	// Log output
	log.Printf("Success: Processed %d records(s) in %.4f second(s)\n",
		count, time.Since(startTime).Seconds())
}

// Handle input json file
func (s *JSONSelect) handleFile(filePath string, ch chan int, wg *sync.WaitGroup) int {
	// Open the input file
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		log.Fatal("Error: Cannot read input file '" + filePath + "'")
	}

	scanner := bufio.NewScanner(f)

	// Record line number
	line := 0

	// Record record count
	count := 0

	// Scan the input file line by line
	for scanner.Scan() {
		line++
		// Skip if the line is empty
		if len(scanner.Bytes()) == 0 {
			continue
		}

		// Copy from scanner to a new slice to allocate memory
		bytes := make([]byte, len(scanner.Bytes()))
		copy(bytes, scanner.Bytes())

		// Increment count, occupy a channel, add to wait group, and start the routine
		count++
		ch <- 1
		wg.Add(1)
		go s.startRoutine(&bytes, ch, filePath, line, wg)
	}
	// return count of processed records
	return count
}

// Start a goroutine to handle a single record
func (s *JSONSelect) startRoutine(input *[]byte, ch chan int, filePath string, line int, wg *sync.WaitGroup) {
	s.handleJSON(input, filePath, line)

	// Finish the routine
	wg.Done()
	<-ch
}

// Handle a single JSON object
func (s *JSONSelect) handleJSON(input *[]byte, filePath string, line int) {
	// Parse input json
	var v interface{}
	err := json.Unmarshal(*input, &v)
	if err != nil {
		if errors.Is(&json.SyntaxError{}, err) {
			log.Fatal("Error: Line " + strconv.Itoa(line) + " of '" + filePath + "' is not in valid JSON format")
		} else {
			log.Fatal(err)
		}
	}

	// Apply every rule on files, stop if match any rule
	for _, r := range s.rules {
		// If all conditions are met, write to specific output
		if s.processRule(v, *r) {
			s.write(input, r.Output)
			return
		}
	}

	// If no rule is met, send to default
	s.write(input, "default")
}

// Return if all conditions in the rule is met
func (s *JSONSelect) processRule(v interface{}, r Rule) bool {
	for _, c := range r.Conditions {
		if !s.processCondition(v, c) {
			return false
		}
	}
	return true
}

// Return if the condition is met
func (s *JSONSelect) processCondition(v interface{}, c *Condition) bool {
	return s.process("", v, c) != c.Exclude
}

// Process non-string elements
func (s *JSONSelect) process(k string, v interface{}, c *Condition) bool {
	switch v.(type) {
	case map[string]interface{}:
		return s.processMap(v.(map[string]interface{}), c)
	case []interface{}:
		return s.processArray(v.([]interface{}), k, c)
	}
	return false
}

// Process maps
func (s *JSONSelect) processMap(m map[string]interface{}, c *Condition) bool {
	k, next, _ := strings.Cut(c.Key, ".")
	v, found := m[k]
	if found {
		switch v.(type) {
		case string:
			if next == "" {
				return s.test(v.(string), c)
			}
		default:
			c.Key = next
			return s.process(k, v, c)
		}
	}
	return false
}

// Process arrays
func (s *JSONSelect) processArray(a []interface{}, k string, c *Condition) bool {
	for _, v := range a {
		switch v.(type) {
		case string:
			if k == "" {
				return s.test(v.(string), c)
			}
		default:
			return s.process(k, v, c)
		}
	}
	return false
}

// Test if the field matches the condition
func (s *JSONSelect) test(v string, c *Condition) bool {
	for _, value := range c.Values {
		switch c.Type {
		case "match":
			if v == value {
				return true
			}
			break
		case "prefix":
			if strings.HasPrefix(v, value) {
				return true
			}
			break
		case "suffix":
			if strings.HasSuffix(v, value) {
				return true
			}
			break
		case "exist":
			return true
		case "regex":
			// Match regex pattern, parsing error is ignored and return false
			m, _ := regexp.MatchString(value, v)
			if m {
				return true
			}
			break
		default:
			log.Fatal("Error: Invalid condition type '" + c.Type + "'")
		}
	}
	return false
}

// Write to the output file
func (s *JSONSelect) write(json *[]byte, output string) {
	// Get the file pointer from map
	f := (*s.outputMap)[output]

	// Append a new line character
	*json = append(*json, byte('\n'))

	// Write to file, internally thread safe
	_, err := f.Write(*json)
	if err != nil {
		log.Fatal("Error: Failed to write to '" + path.Join(s.config.outputPath, output) + "'")
	}
}
