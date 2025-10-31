package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"gopkg.in/yaml.v3"
)

var Version string

func main() {
	var force bool
	flag.BoolVar(&force, "force", false, "Overwrite output file if it exists")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <input-file> <output-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Converts OpenAPI v2 to v3\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <input-file>  Path to the input OpenAPI v2 file (JSON or YAML).\n")
		fmt.Fprintf(os.Stderr, "  <output-file> Path to the output OpenAPI v3 file (JSON or YAML, inferred from extension).\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		flag.Usage()
		os.Exit(1)
	}

	input := args[0]
	output := args[1]

	if err := convertOpenAPIDocs(input, output, force); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to convert input: %v\n", err)
		os.Exit(1)
	}
}

func convertOpenAPIDocs(input, output string, force bool) error {
	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", input)
	}

	if !force {
		if _, err := os.Stat(output); err == nil {
			return fmt.Errorf("output file already exists; use --force to overwrite: %s", output)
		}
	}

	inputData, err := os.ReadFile(input)
	if err != nil {
		return fmt.Errorf("unable to read input file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(input))
	isYAML := ext == ".yaml" || ext == ".yml"
	isJSON := ext == ".json"

	if !isYAML && !isJSON {
		return fmt.Errorf("unsupported input file format: %s", ext)
	}

	var temp interface{}
	if isYAML {
		if err = yaml.Unmarshal(inputData, &temp); err != nil {
			return fmt.Errorf("unable to unmarshal YAML document: %w", err)
		}
	} else {
		if err = json.Unmarshal(inputData, &temp); err != nil {
			return fmt.Errorf("unable to unmarshal JSON document: %w", err)
		}
	}

	jsonData, err := json.Marshal(temp)
	if err != nil {
		return fmt.Errorf("unable to marshal to JSON: %w", err)
	}

	var openAPIDocs openapi2.T
	if err = json.Unmarshal(jsonData, &openAPIDocs); err != nil {
		return fmt.Errorf("unable to unmarshal OpenAPI v2 document: %w", err)
	}

	// Convert OpenAPI v2 to v3
	convertedDocs, err := openapi2conv.ToV3(&openAPIDocs)
	if err != nil {
		return fmt.Errorf("unable to convert OpenAPI v2 document to v3: %w", err)
	}

	// Inject version if set
	if Version != "" {
		convertedDocs.Info.Version = Version
	}

	if err = os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return fmt.Errorf("unable to create output directory: %w", err)
	}

	var out []byte
	if isYAML {
		out, err = yaml.Marshal(convertedDocs)
		if err != nil {
			return fmt.Errorf("unable to marshal OpenAPI v3 YAML document: %w", err)
		}
	} else {
		out, err = json.MarshalIndent(convertedDocs, "", "  ")
		if err != nil {
			return fmt.Errorf("unable to marshal OpenAPI v3 JSON document: %w", err)
		}
	}

	if err = os.WriteFile(output, out, 0o644); err != nil {
		return fmt.Errorf("unable to write OpenAPI v3 document: %w", err)
	}

	return nil
}
