package docs

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/orderedmap"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func GenerateProjectDocsCmd(rootCmd *cobra.Command, ch *cmdutil.Helper) *cobra.Command {
	var projectPath, rillyamlPath, outputDir string

	cmd := &cobra.Command{
		Use:   "generate-project",
		Short: "Generate Markdown docs from JSON Schemas for Project files",
		RunE: func(cmd *cobra.Command, args []string) error {

			if _, err := os.Stat(outputDir); os.IsNotExist(err) {
				if err := os.MkdirAll(outputDir, fs.ModePerm); err != nil {
					return err
				}
			}

			projectFilesSchema, err := parseSchemaWithRefs(projectPath)
			if err != nil {
				return fmt.Errorf("resource schema error: %w", err)
			}
			// ideally rillyaml should be part of project.schem.json but currenly it can't be
			rillYamlSchema, err := parseSchemaWithRefs(rillyamlPath)
			if err != nil {
				return fmt.Errorf("rillyaml schema error: %w", err)
			}
			projectFilesSchema.OneOf = append(projectFilesSchema.OneOf, rillYamlSchema)

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
				resourceFilebuf.WriteString(fmt.Sprintf("\n%s\n\n", resource.Description))
				resourceFilebuf.WriteString("## Properties\n")
				resourceFilebuf.WriteString(generateDoc("", resource, "", requiredMap))

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

	cmd.Flags().StringVar(&projectPath, "project", "", "Path to project.schema.json")
	cmd.Flags().StringVar(&rillyamlPath, "rillyaml", "", "Path to rillyaml.schema.json")
	cmd.Flags().StringVar(&outputDir, "out", "", "Output directory for generated docs")
	cmd.MarkFlagRequired("project")
	cmd.MarkFlagRequired("rillyaml")
	cmd.MarkFlagRequired("out")
	return cmd
}

type JSONSchema struct {
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Type        interface{}              `json:"type"`
	Properties  *orderedmap.OrderedMap   `json:"properties"`
	Examples    []*orderedmap.OrderedMap `json:"examples"`
	Required    []string                 `json:"required"`
	Items       *JSONSchema              `json:"items"`
	Enum        []interface{}            `json:"enum"`
	OneOf       []*JSONSchema            `json:"oneOf"`
	AnyOf       []*JSONSchema            `json:"anyOf"`
	AllOf       []*JSONSchema            `json:"allOf"`
}

func getRequiredMap(required []string) map[string]bool {
	reqMap := make(map[string]bool)
	for _, r := range required {
		reqMap[r] = true
	}
	return reqMap
}

func getTypeString(schema *JSONSchema) string {
	if schema == nil {
		return ""
	}
	if schema.Type == nil {
		if schema.OneOf != nil {
			return "- _[oneOf]_ "
		}
		if schema.AnyOf != nil {
			return "- _[anyOf]_ "
		}
		if schema.AnyOf != nil {
			return "- _[allOf]_ "
		}
		return ""
	}
	t := schema.Type
	if s, ok := t.(string); ok {
		if s == "array" {
			if schema.Items.Type == "string" {
				return "- _[array of string]_ "
			} else if schema.Items.Type == "object" {
				return "- _[array of object]_ "
			} else if schema.Items.Type == nil {
				if len(schema.Items.OneOf) > 0 {
					return "- _[array of oneOf]_ "
				}
				if len(schema.Items.AnyOf) > 0 {
					return "- _[array of anyOf]_ "
				}

			}
		}
		return fmt.Sprintf("- _[%s]_ ", s)
	}
	return ""
}

func getDescriptionString(description string) string {
	if description == "" {
		return ""
	} else {
		return fmt.Sprintf("- %s", description)
	}

}

func generateDoc(parentName string, schema *JSONSchema, indent string, requiredFields map[string]bool) string {
	var doc strings.Builder
	var listString = "- "
	if indent == "" {
		listString = ""
	}
	if schema.Type == "object" && schema.Properties != nil {
		for _, propName := range schema.Properties.Keys() {
			val, _ := schema.Properties.Get(propName)
			b, _ := json.Marshal(val)

			var propSchema *JSONSchema
			if err := json.Unmarshal(b, &propSchema); err != nil {
				panic(err)
			}
			required := ""
			if requiredFields[propName] {
				required = " _(required)_"
			}
			doc.WriteString(fmt.Sprintf("\n\n%s%s**`%s`**  %s%s %s", indent, listString, propName, getTypeString(propSchema), getDescriptionString(propSchema.Description), required))
			if propSchema.Type == "object" && !(propName == "dev" || propName == "prod") {
				doc.WriteString(generateDoc(propName, propSchema, indent+"  ", getRequiredMap(propSchema.Required)))
			} else if propSchema.Type == "array" || propSchema.Type == nil {
				doc.WriteString(generateDoc(propName, propSchema, indent+"  ", getRequiredMap(propSchema.Required)))
			}
		}
	} else if schema.Type == "array" && schema.Items != nil {
		doc.WriteString(generateDoc(parentName, schema.Items, indent, getRequiredMap(schema.Items.Required)))
	}

	if len(schema.OneOf) > 0 {
		// single oneof is always selected to print it as same level as properties.
		if len(schema.OneOf) == 1 {
			doc.WriteString(generateDoc(parentName, schema.OneOf[0], indent, getRequiredMap(schema.OneOf[0].Required)))
		} else {
			// root level(parentName == "") options handling is different
			if parentName == "" {
				doc.WriteString("\n\n## One of Properties Options\n")
				for _, subSchema := range schema.OneOf {
					doc.WriteString(fmt.Sprintf("- [%s](#%s)\n", subSchema.Title, subSchema.Title))
				}
				for _, subSchema := range schema.OneOf {
					if len(schema.OneOf) != 1 && (subSchema.Properties != nil || subSchema.Type != nil) {
						doc.WriteString(fmt.Sprintf("\n\n### %s", subSchema.Title))
						doc.WriteString(fmt.Sprintf("\n\n%s", subSchema.Description))
					}
					doc.WriteString(generateDoc(parentName, subSchema, indent, getRequiredMap(subSchema.Required)))
				}
			} else {
				for i, subSchema := range schema.OneOf {
					if len(schema.OneOf) != 1 && (subSchema.Properties != nil || subSchema.Type != nil) {
						doc.WriteString(fmt.Sprintf("\n\n%s*option %d* %s%s", indent, i+1, getTypeString(subSchema), getDescriptionString(subSchema.Description)))
					}
					doc.WriteString(generateDoc(parentName, subSchema, indent, getRequiredMap(subSchema.Required)))
				}
			}
		}
	}
	if len(schema.AnyOf) > 0 {
		for i, subSchema := range schema.AnyOf {
			if len(schema.AnyOf) != 1 && (subSchema.Properties != nil || subSchema.Type != nil) {
				doc.WriteString(fmt.Sprintf("\n\n%s*option %d* %s%s", indent, i+1, getTypeString(subSchema), getDescriptionString(subSchema.Description)))
			}
			doc.WriteString(generateDoc(parentName, subSchema, indent, getRequiredMap(subSchema.Required)))
		}
	}
	if len(schema.AllOf) > 0 {
		for _, subSchema := range schema.AllOf {
			doc.WriteString(generateDoc(parentName, subSchema, indent, getRequiredMap(subSchema.Required)))
		}
	}

	return doc.String()
}

// parseSchemaWithRefs reads and fully resolves all $ref in the JSON Schema
func parseSchemaWithRefs(path string) (*JSONSchema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	root := orderedmap.New()
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("unmarshal root schema: %w", err)
	}
	node := root //
	if err := resolveRefs(node, root); err != nil {
		return nil, fmt.Errorf("failed to resolve $refs: %w", err)
	}

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

func resolveRefs(node *orderedmap.OrderedMap, root *orderedmap.OrderedMap) error {
	// Handle $ref replacement
	if refVal, ok := node.Get("$ref"); ok {
		if refStr, ok := refVal.(string); ok && strings.HasPrefix(refStr, "#/") {
			resolved, err := resolveJSONPointer(root, refStr[2:])
			if err != nil {
				return err
			}
			node.Delete("$ref")
			for _, key := range resolved.Keys() {
				val, _ := resolved.Get(key)
				node.Set(key, val)
			}
		}
	}

	// Recurse into keys
	for _, key := range node.Keys() {
		val, _ := node.Get(key)

		switch v := val.(type) {
		case *orderedmap.OrderedMap:
			if err := resolveRefs(v, root); err != nil {
				return err
			}
		case orderedmap.OrderedMap:
			// Convert to pointer so we can mutate
			copy := v
			if err := resolveRefs(&copy, root); err != nil {
				return err
			}
			node.Set(key, copy)
		case []interface{}:
			for i, item := range v {
				switch itemTyped := item.(type) {
				case *orderedmap.OrderedMap:
					if err := resolveRefs(itemTyped, root); err != nil {
						return err
					}
				case orderedmap.OrderedMap:
					copy := itemTyped
					if err := resolveRefs(&copy, root); err != nil {
						return err
					}
					v[i] = copy
				}
			}
			node.Set(key, v)
		}
	}
	return nil
}

func resolveJSONPointer(root *orderedmap.OrderedMap, pointer string) (*orderedmap.OrderedMap, error) {
	parts := strings.Split(pointer, "/")
	var current any
	current = root
	for _, part := range parts {
		m, ok := current.(*orderedmap.OrderedMap)
		if !ok {
			return nil, fmt.Errorf("invalid pointer resolution at part '%s'", part)
		}

		val, exists := m.Get(part)
		if !exists {
			return nil, fmt.Errorf("key '%s' not found", part)
		}
		switch valTyped := val.(type) {
		case orderedmap.OrderedMap:
			current = &valTyped
		default:
			current = valTyped.(*orderedmap.OrderedMap)
		}
	}

	resolved, ok := current.(*orderedmap.OrderedMap)
	if !ok {
		return nil, fmt.Errorf("resolved reference is not an object")
	}
	return resolved, nil
}

func sanitizeFileName(name string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(name, " YAML", ""), " ", "-"))
}
