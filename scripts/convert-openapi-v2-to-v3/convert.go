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
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

var Version string

func main() {
	var force bool
	var publicOnly bool
	flag.BoolVar(&force, "force", false, "Overwrite output file if it exists")
	flag.BoolVar(&publicOnly, "public-only", false, "Only include public APIs (those with x-visibility: public)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <input-file> <output-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Converts OpenAPI v2 to v3\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  <input-file>  Path to the input OpenAPI v2 file (JSON or YAML).\n")
		fmt.Fprintf(os.Stderr, "  <output-file> Path to the output OpenAPI v3 file (JSON or YAML, inferred from output file extension).\n")
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

	if err := convertOpenAPIDocs(input, output, force, publicOnly); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to convert input: %v\n", err)
		os.Exit(1)
	}
}

func convertOpenAPIDocs(input, output string, force, publicOnly bool) error {
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

	outExt := strings.ToLower(filepath.Ext(output))
	isYAMLOutput := outExt == ".yaml" || outExt == ".yml"
	isJSONOutput := outExt == ".json"

	if !isYAMLOutput && !isJSONOutput {
		return fmt.Errorf("unsupported output file format: %s", outExt)
	}

	var temp interface{}
	// Try to detect format from content: first try JSON, then YAML
	if err = json.Unmarshal(inputData, &temp); err != nil {
		// If JSON fails, try YAML
		if yamlErr := yaml.Unmarshal(inputData, &temp); yamlErr != nil {
			return fmt.Errorf("unable to unmarshal document as JSON or YAML: JSON error: %w, YAML error: %w", err, yamlErr)
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

	// Prune to only keep public visibility
	prunePublicOnly(convertedDocs, publicOnly)

	// Inject version if set
	if Version != "" {
		convertedDocs.Info.Version = Version
	}

	if err = os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return fmt.Errorf("unable to create output directory: %w", err)
	}

	var out []byte
	if isYAMLOutput {
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

func prunePublicOnly(doc *openapi3.T, publicOnly bool) {
	if !publicOnly {
		return
	}

	for path, pathItem := range doc.Paths.Map() {
		if pathItem == nil {
			continue
		}

		// Remove operations that don't meet criteria
		operations := []*openapi3.Operation{pathItem.Get, pathItem.Post, pathItem.Put, pathItem.Delete, pathItem.Options, pathItem.Head, pathItem.Patch, pathItem.Trace}
		for i, op := range operations {
			if op == nil {
				continue
			}
			hasPublic := hasPublicVisibility(op.Extensions)
			if !hasPublic {
				// remove the operation
				switch i {
				case 0:
					pathItem.Get = nil
				case 1:
					pathItem.Post = nil
				case 2:
					pathItem.Put = nil
				case 3:
					pathItem.Delete = nil
				case 4:
					pathItem.Options = nil
				case 5:
					pathItem.Head = nil
				case 6:
					pathItem.Patch = nil
				case 7:
					pathItem.Trace = nil
				}
			}
		}

		// If no operations left, remove the path
		if pathItem.Get == nil && pathItem.Post == nil && pathItem.Put == nil && pathItem.Delete == nil &&
			pathItem.Options == nil && pathItem.Head == nil && pathItem.Patch == nil && pathItem.Trace == nil {
			doc.Paths.Map()[path] = nil
		}
	}
}

func hasPublicVisibility(ext map[string]interface{}) bool {
	if v, ok := ext["x-visibility"]; ok {
		if s, ok := v.(string); ok && s == "public" {
			return true
		}
	}
	return false
}
