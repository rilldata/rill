package docs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

type JSONSchema struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Type        interface{}            `json:"type"`
	Properties  map[string]*JSONSchema `json:"properties"`
	Required    []string               `json:"required"`
	Items       *JSONSchema            `json:"items"`
	Enum        []interface{}          `json:"enum"`
	OneOf       []*JSONSchema          `json:"oneOf"`
	AnyOf       []*JSONSchema          `json:"anyOf"`
	AllOf       []*JSONSchema          `json:"allOf"`
}

func getRequiredMap(required []string) map[string]bool {
	reqMap := make(map[string]bool)
	for _, r := range required {
		reqMap[r] = true
	}
	return reqMap
}

func generateDoc(schema *JSONSchema, indent string, requiredFields map[string]bool) string {
	var doc strings.Builder

	if schema.Title != "" {
		doc.WriteString(fmt.Sprintf("%s## %s\n\n", indent, schema.Title))
	}
	if schema.Description != "" {
		doc.WriteString(fmt.Sprintf("%s%s\n\n", indent, schema.Description))
	}
	if schema.Type != "" {
		doc.WriteString(fmt.Sprintf("%sType: `%s`\n\n", indent, schema.Type))
	}
	if len(schema.Enum) > 0 {
		doc.WriteString(fmt.Sprintf("%sEnum: `%v`\n\n", indent, schema.Enum))
	}
	if schema.Type == "object" {
		doc.WriteString(fmt.Sprintf("%s## Properties:\n", indent))
		for propName, propSchema := range schema.Properties {
			required := ""
			if requiredFields[propName] {
				required = " _(required)_"
			}
			doc.WriteString(fmt.Sprintf("\n%s- **%s**%s:\n", indent, propName, required))
			doc.WriteString(generateDoc(propSchema, indent+"  ", getRequiredMap(propSchema.Required)))
		}
	} else if schema.Type == "array" && schema.Items != nil {
		doc.WriteString(fmt.Sprintf("%s#### Array Items:\n", indent))
		doc.WriteString(generateDoc(schema.Items, indent+"  ", getRequiredMap(schema.Items.Required)))
	}

	if len(schema.OneOf) > 0 {
		doc.WriteString(fmt.Sprintf("%s#### One of the following:\n", indent))
		for i, subSchema := range schema.OneOf {
			doc.WriteString(fmt.Sprintf("%s- Option %d:\n", indent, i+1))
			doc.WriteString(generateDoc(subSchema, indent+"  ", getRequiredMap(subSchema.Required)))
		}
	}
	if len(schema.AnyOf) > 0 {
		doc.WriteString(fmt.Sprintf("%s#### Any of the following:\n", indent))
		for i, subSchema := range schema.AnyOf {
			doc.WriteString(fmt.Sprintf("%s- Option %d:\n", indent, i+1))
			doc.WriteString(generateDoc(subSchema, indent+"  ", getRequiredMap(subSchema.Required)))
		}
	}
	if len(schema.AllOf) > 0 {
		doc.WriteString(fmt.Sprintf("%s#### All of the following:\n", indent))
		for i, subSchema := range schema.AllOf {
			doc.WriteString(fmt.Sprintf("%s- Part %d:\n", indent, i+1))
			doc.WriteString(generateDoc(subSchema, indent+"  ", getRequiredMap(subSchema.Required)))
		}
	}

	return doc.String()
}

// parseSchema reads and fully resolves all $ref in the JSON Schema
func parseSchemaWithRefs(path string) (*JSONSchema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("unmarshal root schema: %w", err)
	}

	// Resolve all internal $ref recursively
	if err := resolveRefs(root, root); err != nil {
		return nil, fmt.Errorf("failed to resolve $refs: %w", err)
	}

	// Marshal and unmarshal into JSONSchema struct
	finalData, err := json.Marshal(root)
	if err != nil {
		return nil, fmt.Errorf("marshal resolved schema: %w", err)
	}

	var schema JSONSchema
	if err := json.Unmarshal(finalData, &schema); err != nil {
		return nil, fmt.Errorf("unmarshal resolved JSONSchema: %w", err)
	}
	return &schema, nil
}

func resolveRefs(node interface{}, root map[string]interface{}) error {
	switch typed := node.(type) {
	case map[string]interface{}:
		if ref, ok := typed["$ref"].(string); ok && strings.HasPrefix(ref, "#/") {
			resolved, err := resolveJSONPointer(root, ref[2:]) // Strip "#/"
			if err != nil {
				return err
			}
			delete(typed, "$ref")
			for k, v := range resolved {
				typed[k] = v
			}
		}
		for _, v := range typed {
			if err := resolveRefs(v, root); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, v := range typed {
			if err := resolveRefs(v, root); err != nil {
				return err
			}
		}
	}
	return nil
}

func resolveJSONPointer(root map[string]interface{}, pointer string) (map[string]interface{}, error) {
	parts := strings.Split(pointer, "/")
	current := interface{}(root)

	for _, part := range parts {
		part = strings.ReplaceAll(part, "~1", "/")
		part = strings.ReplaceAll(part, "~0", "~")

		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid pointer resolution at %s", part)
		}
		current = m[part]
	}

	resolved, ok := current.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("resolved ref is not an object")
	}
	return resolved, nil
}

func sanitizeFileName(name string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(name, " YAML", ""), " ", "-"))
}

func GenerateProjectDocsCmd(rootCmd *cobra.Command, ch *cmdutil.Helper) *cobra.Command {
	var resourcePath, rillPath, outputDir string

	cmd := &cobra.Command{
		Use:   "generate-docs",
		Short: "Generate Markdown docs from JSON Schemas (resource + rillyaml)",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse resource.schema.json
			projectFilesSchema, err := parseSchemaWithRefs(resourcePath)
			if err != nil {
				return fmt.Errorf("resource schema error: %w", err)
			}
			rillYamlSchema, err := parseSchemaWithRefs(rillPath)
			if err != nil {
				return fmt.Errorf("rillyaml schema error: %w", err)
			}
			//
			projectFilesSchema.OneOf = append(projectFilesSchema.OneOf, rillYamlSchema)

			// Prepare index content
			var projectFilesbuf strings.Builder

			sidebar_position := 0
			projectFilesbuf.WriteString("---\n")
			projectFilesbuf.WriteString("note: GENERATED. DO NOT EDIT.\n")
			projectFilesbuf.WriteString(fmt.Sprintf("title: %s\n", projectFilesSchema.Title))
			projectFilesbuf.WriteString(fmt.Sprintf("sidebar_position: %d\n", sidebar_position))
			projectFilesbuf.WriteString("---\n")

			projectFilesbuf.WriteString("## Overview\n\n")
			projectFilesbuf.WriteString(fmt.Sprintf("%s\n\n", projectFilesSchema.Description))
			projectFilesbuf.WriteString("## Project files types\n\n")

			for _, resource := range projectFilesSchema.OneOf {
				sidebar_position += 1
				fileName := sanitizeFileName(resource.Title) + ".md"
				filePath := filepath.Join(outputDir, fileName)
				requiredMap := getRequiredMap(resource.Required)
				var resourceFilebuf strings.Builder
				resourceFilebuf.WriteString("---\n")
				resourceFilebuf.WriteString("note: GENERATED. DO NOT EDIT.\n")
				resourceFilebuf.WriteString(fmt.Sprintf("title: %s\n", resource.Title))
				resourceFilebuf.WriteString(fmt.Sprintf("sidebar_position: %d\n", sidebar_position))
				resourceFilebuf.WriteString("---\n")

				resourceFilebuf.WriteString(generateDoc(resource, "", requiredMap))

				if err := os.WriteFile(filePath, []byte(resourceFilebuf.String()), 0644); err != nil {
					return fmt.Errorf("failed writing resource doc: %w", err)
				}
				projectFilesbuf.WriteString(fmt.Sprintf("- [%s](%s)\n", resource.Title, fileName))
			}

			if err := os.WriteFile(filepath.Join(outputDir, "index.md"), []byte(projectFilesbuf.String()), 0644); err != nil {
				return fmt.Errorf("failed writing index.md: %w", err)
			}

			fmt.Printf("Documentation generated in %s\n", outputDir)
			return nil
		},
	}

	cmd.Flags().StringVar(&resourcePath, "resource", "", "Path to resource.schema.json")
	cmd.Flags().StringVar(&rillPath, "rill", "", "Path to rillyaml.schema.json")
	cmd.Flags().StringVarP(&outputDir, "out", "o", "./docs", "Output directory for generated docs")
	cmd.MarkFlagRequired("resource")
	cmd.MarkFlagRequired("rill")
	return cmd
}
