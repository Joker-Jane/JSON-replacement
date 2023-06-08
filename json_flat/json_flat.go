/*
This program reads file(s) containing compressed JSON records with dots in keys, and flat these
records to output file(s) in the form of original records.

Usage:

./json_flat [flags]

Flags:

	-i input_path
		Set the path to the input file or directory.

	-o output_path
		Set the path to the output directory.
*/

package json_flat

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type JSONFlat struct {
	// Configs
	config *Config
}

func NewJSONFlat(config *Config) *JSONFlat {
	// Check if all arguments are specified
	if config.inputPath == "" || config.outputPath == "" {
		log.Fatal("Usage: ./json_select -i input -o output")
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

	// Construct JSONSelect object
	flat := &JSONFlat{
		config: config,
	}

	return flat
}

func (flat *JSONFlat) Exec() {
	// Record start time
	startTime := time.Now()

	// Record count
	count := 0

	// Walk through and process the input file tree
	err := filepath.WalkDir(flat.config.inputPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			flat.handleFile(path)
			count++
		}
		return err
	})
	if err != nil {
		log.Fatal("Error: Failed to walk through the input directory")
	}

	// Log output
	log.Printf("Success: Processed %d file(s) in %.4f second(s)\n",
		count, time.Since(startTime).Seconds())
}

func (flat *JSONFlat) handleFile(filePath string) {
	// Open the input file
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		log.Fatal("Error: Cannot read input file '" + filePath + "'")
	}

	scanner := bufio.NewScanner(f)

	// Record line number
	line := 0

	// Get target output path
	target := strings.Replace(filePath, flat.config.inputPath, flat.config.outputPath, 1)

	// Get parent directory of the target
	dir, _ := filepath.Split(target)

	// Create the directory if the file is not in root
	if dir != "" {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			log.Fatal("Error: Failed to create directory '" + dir + "'")
		}
	}

	// Open or create the file
	outputFile, err := os.Create(target)
	defer outputFile.Close()
	if err != nil {
		log.Fatal("Error: Failed to open or create file '" + target + "'")
	}

	// Scan the input file line by line
	for scanner.Scan() {
		line++
		// Skip if the line is empty
		if len(scanner.Text()) == 0 {
			continue
		}

		// Copy from scanner to a new slice to allocate memory
		bytes := make([]byte, len(scanner.Bytes()))
		copy(bytes, scanner.Bytes())

		// Handle the line and get result
		result := flat.handleJSON(&bytes, filePath, line)

		// Write to target file
		_, err = fmt.Fprintln(outputFile, string(result))

		if err != nil {
			log.Fatal("Error: Cannot write to '" + target + "'")
		}
	}
}

func (flat *JSONFlat) handleJSON(input *[]byte, filePath string, line int) []byte {
	// Parse input json
	var v map[string]interface{}
	err := json.Unmarshal(*input, &v)
	if err != nil {
		if errors.Is(&json.SyntaxError{}, err) {
			log.Fatal("Error: Line " + strconv.Itoa(line) + " of '" + filePath + "' is not in valid JSON format")
		} else {
			log.Fatal(err)
		}
	}

	output, _ := json.Marshal(flat.flat(v))
	return output
}

func (flat *JSONFlat) flat(input map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range input {
		keys := strings.Split(k, ".")
		lastKey := keys[len(keys)-1]
		keys = keys[:len(keys)-1]
		currentMap := result
		for _, key := range keys {
			if _, exists := currentMap[key]; !exists {
				currentMap[key] = make(map[string]interface{})
			}
			currentMap = currentMap[key].(map[string]interface{})
		}
		currentMap[lastKey] = v
	}
	return result
}
