package docs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func ConvertOpenAPIDocsCmd(rootCmd *cobra.Command, ch *cmdutil.Helper) *cobra.Command {
	var force bool
	docsCmd := &cobra.Command{
		Use:   "convert-openapi",
		Short: "Converts OpenAPI v2 to v3",
		Long: `
Converts OpenAPI v2 documents to OpenAPI v3 format.

Usage:
  rill docs convert-openapi <input-file> <output-file>

Arguments:
  <input-file>  Path to the input OpenAPI v2 file (JSON or YAML).
  <output-file> Path to the output OpenAPI v3 file (JSON or YAML, inferred from extension).
`,
		Args:   cobra.ExactArgs(2),
		Hidden: !ch.IsDev(),
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.DisableAutoGenTag = true
			input := args[0]
			output := args[1]

			if err := convertOpenAPIDocs(ch, input, output, force); err != nil {
				cmd.PrintErrf("Unable to convert input: %v", err)
			}
		},
	}

	docsCmd.Flags().BoolVar(&force, "force", false, "Overwrite output file if it exists")

	return docsCmd
}

func convertOpenAPIDocs(ch *cmdutil.Helper, input, output string, force bool) error {
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

	var openAPIDocs openapi2.T
	if isYAML {
		if err = yaml.Unmarshal(inputData, &openAPIDocs); err != nil {
			return fmt.Errorf("unable to unmarshal OpenAPI v2 YAML document: %w", err)
		}
	} else {
		if err = json.Unmarshal(inputData, &openAPIDocs); err != nil {
			return fmt.Errorf("unable to unmarshal OpenAPI v2 JSON document: %w", err)
		}
	}

	// Convert OpenAPI v2 to v3
	convertedDocs, err := openapi2conv.ToV3(&openAPIDocs)
	if err != nil {
		return fmt.Errorf("unable to convert OpenAPI v2 document to v3: %w", err)
	}

	// Inject version details
	if ch.Version.Number != "" {
		convertedDocs.Info.Version = ch.Version.Number
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
