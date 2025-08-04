package docs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func GenerateProjectDocsCmd(rootCmd *cobra.Command, ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "generate-project",
		Short:  "Generate Markdown docs from JSON Schemas for Project files",
		Args:   cobra.ExactArgs(1),
		Hidden: !ch.IsDev(),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputDir := args[0]
			if _, err := os.Stat(outputDir); os.IsNotExist(err) {
				if err := os.MkdirAll(outputDir, fs.ModePerm); err != nil {
					return err
				}
			}

			projectPath := "runtime/parser/schema/project.schema.yaml"
			projectFilesSchema, err := parseSchemaYAML(projectPath)
			if err != nil {
				return fmt.Errorf("resource schema error: %w", err)
			}

			rillyamlPath := "runtime/parser/schema/rillyaml.schema.yaml"
			rillYamlSchema, err := parseSchemaYAML(rillyamlPath)
			if err != nil {
				return fmt.Errorf("rillyaml schema error: %w", err)
			}

			// Add rillyaml to projectFilesSchema's oneOf
			oneOfNode := getNodeForKey(projectFilesSchema, "oneOf")
			if oneOfNode == nil {
				oneOfNode = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
				projectFilesSchema.Content = append(projectFilesSchema.Content,
					&yaml.Node{Kind: yaml.ScalarNode, Value: "oneOf"},
					oneOfNode,
				)
			}
			oneOfNode.Content = append(oneOfNode.Content, rillYamlSchema)

			var projectFilesbuf strings.Builder
			sidebarPosition := 30

			title := getScalarValue(projectFilesSchema, "title")
			desc := getPrintableDescription(projectFilesSchema, "", "")

			projectFilesbuf.WriteString("---\n")
			projectFilesbuf.WriteString("note: GENERATED. DO NOT EDIT.\n")
			projectFilesbuf.WriteString(fmt.Sprintf("title: %s\n", title))
			projectFilesbuf.WriteString(fmt.Sprintf("sidebar_position: %d\n", sidebarPosition))
			projectFilesbuf.WriteString("---\n\n")

			projectFilesbuf.WriteString("## Overview\n\n")
			projectFilesbuf.WriteString(fmt.Sprintf("%s\n\n", desc))
			projectFilesbuf.WriteString("## Project files types\n\n")

			for _, resource := range oneOfNode.Content {
				sidebarPosition++
				var resourceFilebuf strings.Builder
				requiredMap := getRequiredMapFromNode(resource)
				resourceFilebuf.WriteString(generateDoc(sidebarPosition, 0, resource, "", requiredMap))
				resTitle := getScalarValue(resource, "title")
				resId := getScalarValue(resource, "$id")
				
				// Use $id if available, otherwise fall back to title
				var fileName string
				if resId != "" {
					// Remove .schema.yaml extension and convert to .md
					fileName = strings.TrimSuffix(resId, ".schema.yaml") + ".md"
				} else {
					fileName = sanitizeFileName(resTitle) + ".md"
				}
				
				filePath := filepath.Join(outputDir, fileName)
				if err := os.WriteFile(filePath, []byte(resourceFilebuf.String()), 0o644); err != nil {
					return fmt.Errorf("failed writing resource doc: %w", err)
				}
				projectFilesbuf.WriteString(fmt.Sprintf("\n- [%s](%s)", resTitle, fileName))
			}

			if err := os.WriteFile(filepath.Join(outputDir, "index.md"), []byte(projectFilesbuf.String()), 0o644); err != nil {
				return fmt.Errorf("failed writing index.md: %w", err)
			}
			fmt.Printf("Documentation generated in %s\n", outputDir)

			return nil
		},
	}
	return cmd
}

func sanitizeFileName(name string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(name, " YAML", ""), " ", "-"))
}

func getScalarValue(node *yaml.Node, key string) string {
	val := getNodeForKey(node, key)
	if val != nil && val.Kind == yaml.ScalarNode {
		return val.Value
	}
	return ""
}

// Get value node of a mapping key
func getNodeForKey(node *yaml.Node, key string) *yaml.Node {
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func parseSchemaYAML(path string) (*yaml.Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("unmarshal YAML schema: %w", err)
	}

	// root should be a document node
	if len(root.Content) == 0 {
		return nil, fmt.Errorf("empty YAML document")
	}
	doc := root.Content[0]

	if err := resolveRefsYAML(doc, doc); err != nil {
		return nil, fmt.Errorf("resolve $refs: %w", err)
	}

	return doc, nil
}

// resolveRefsYAML traverses the YAML node tree and resolves $refs in-place.
func resolveRefsYAML(node, root *yaml.Node) error {
	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]

			if keyNode.Value == "$ref" && valNode.Kind == yaml.ScalarNode {
				if strings.HasPrefix(valNode.Value, "#/") {
					// Resolve local reference
					ptrPath := strings.TrimPrefix(valNode.Value, "#/")
					resolved, err := resolveYAMLPointer(root, ptrPath)
					if err != nil {
						return fmt.Errorf("resolve $ref %q: %w", valNode.Value, err)
					}

					// Replace the entire mapping with the resolved content
					// First, remove $ref entry
					node.Content = append(node.Content[:i], node.Content[i+2:]...)
					// Then merge resolved content into current node
					if resolved.Kind == yaml.MappingNode {
						// Insert resolved mapping node's content at current position
						node.Content = append(resolved.Content, node.Content...)
					} else {
						return fmt.Errorf("$ref does not point to a mapping node")
					}
					// We modified Content length; restart loop
					return resolveRefsYAML(node, root)
				} else if strings.HasSuffix(valNode.Value, ".yaml#") {
					// Resolve external file reference
					fileName := strings.TrimSuffix(valNode.Value, "#")
					// Remove quotes if present
					fileName = strings.Trim(fileName, "'\"")

					// Load the external schema file
					externalSchema, err := parseSchemaYAML("runtime/parser/schema/" + fileName)
					if err != nil {
						return fmt.Errorf("failed to load external schema %q: %w", fileName, err)
					}

					// Replace the entire mapping with the external schema content
					// First, remove $ref entry
					node.Content = append(node.Content[:i], node.Content[i+2:]...)
					// Then merge external schema content into current node
					if externalSchema.Kind == yaml.MappingNode {
						// Insert external schema's content at current position
						node.Content = append(externalSchema.Content, node.Content...)
					} else {
						return fmt.Errorf("external schema %q does not contain a mapping node", fileName)
					}
					// We modified Content length; restart loop
					return resolveRefsYAML(node, root)
				}
			}
			if err := resolveRefsYAML(valNode, root); err != nil {
				return err
			}
		}

	case yaml.SequenceNode:
		for _, item := range node.Content {
			if err := resolveRefsYAML(item, root); err != nil {
				return err
			}
		}

	case yaml.DocumentNode:
		if len(node.Content) > 0 {
			return resolveRefsYAML(node.Content[0], root)
		}
	}
	return nil
}

// resolveYAMLPointer traverses a YAML node using a JSON pointer-style path.
func resolveYAMLPointer(root *yaml.Node, path string) (*yaml.Node, error) {
	parts := strings.Split(path, "/")
	curr := root
	if curr.Kind == yaml.DocumentNode && len(curr.Content) > 0 {
		curr = curr.Content[0]
	}

	for _, part := range parts {
		if curr.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("unexpected kind: expected mapping node at %q", part)
		}

		found := false
		for i := 0; i < len(curr.Content); i += 2 {
			k := curr.Content[i]
			v := curr.Content[i+1]
			if k.Value == part {
				curr = v
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("path not found: %q", part)
		}
	}
	return curr, nil
}

func getPrintableType(node *yaml.Node) string {
	if node == nil || node.Kind != yaml.MappingNode {
		return "_[no type]_"
	}

	// Get the "type" value
	typ := getNodeForKey(node, "type")

	// If no type is present, check for combinators
	if typ == nil {
		if getNodeForKey(node, "oneOf") != nil {
			return "_[oneOf]_"
		}
		if getNodeForKey(node, "anyOf") != nil {
			return "_[anyOf]_"
		}
		if getNodeForKey(node, "allOf") != nil {
			return "_[allOf]_"
		}
		return "_[no type]_"
	}

	if typ.Kind == yaml.ScalarNode && typ.Value == "array" {
		items := getNodeForKey(node, "items")
		if items == nil || items.Kind != yaml.MappingNode {
			return "_[array]_"
		}
		itemsType := getNodeForKey(items, "type")
		if itemsType != nil && itemsType.Kind == yaml.ScalarNode {
			return fmt.Sprintf("_[array of %s]_", itemsType.Value)
		}

		if getNodeForKey(items, "oneOf") != nil {
			return "_[array of oneOf]_"
		}
		if getNodeForKey(items, "anyOf") != nil {
			return "_[array of anyOf]_"
		}
		if getNodeForKey(items, "allOf") != nil {
			return "_[array of allOf]_"
		}
		return "_[array]_"
	} else if typ.Kind == yaml.ScalarNode {
		return fmt.Sprintf("_[%s]_", typ.Value)
	} else if typ.Kind == yaml.SequenceNode {
		types := make([]string, 0, len(typ.Content))
		for _, item := range typ.Content {
			if item.Kind == yaml.ScalarNode {
				types = append(types, item.Value)
			}
		}
		return fmt.Sprintf("_[%s]_", strings.Join(types, ", "))
	}

	return "_[no type]_"
}

func getRequiredMapFromNode(node *yaml.Node) map[string]bool {
	req := map[string]bool{}
	if node == nil || node.Kind != yaml.MappingNode {
		return req
	}

	required := getNodeForKey(node, "required")

	if required != nil && required.Kind == yaml.SequenceNode {
		for _, item := range required.Content {
			req[item.Value] = true
		}
	}
	return req
}

func getPrintableDescription(node *yaml.Node, indentation, defaultValue string) string {
	if node == nil || node.Kind != yaml.MappingNode {
		return defaultValue
	}
	desc := getScalarValue(node, "description")
	if desc == "" {
		return defaultValue
	}

	lines := strings.Split(desc, "\n")
	if len(lines) > 1 {
		// Add indentation to all lines except the first one
		for i := 1; i < len(lines)-1; i++ {
			lines[i] = indentation + lines[i]
		}
		return strings.Join(lines, "\n")
	}
	return desc
}

func generateDoc(sidebarPosition, level int, node *yaml.Node, indent string, requiredFields map[string]bool) string {
	if node == nil || node.Kind != yaml.MappingNode {
		return ""
	}

	var doc strings.Builder
	currentLevel := level
	title := getScalarValue(node, "title")
	description := getPrintableDescription(node, indent, "")

	// Check if this is a connector type for special handling
	isConnector := title == "Connector YAML"

	if level == 0 {
		doc.WriteString("---\n")
		doc.WriteString("note: GENERATED. DO NOT EDIT.\n")
		doc.WriteString(fmt.Sprintf("title: %s\n", title))
		doc.WriteString(fmt.Sprintf("sidebar_position: %d\n", sidebarPosition))
		doc.WriteString("---")
		if description != "" {
			doc.WriteString(fmt.Sprintf("\n\n%s", description))
		}

		level++ // level zero is to print base level info and its only onetime for a page so increasing level
	} else if level == 1 {
		if title != "" {
			doc.WriteString(fmt.Sprintf("\n\n## %s", title))
		}
		if description != "" {
			doc.WriteString(fmt.Sprintf("\n\n%s", description))
		}
	}

	// Properties
	if properties := getNodeForKey(node, "properties"); properties != nil && properties.Kind == yaml.MappingNode {
		for i := 0; i < len(properties.Content); i += 2 {
			propertiesName := properties.Content[i].Value
			propertiesValueNode := properties.Content[i+1]
			required := ""
			if requiredFields[propertiesName] {
				required = "_(required)_"
			}
			if level == 1 {
				doc.WriteString(fmt.Sprintf("\n\n### `%s`", propertiesName))
				doc.WriteString(fmt.Sprintf("\n\n%s - %s %s", getPrintableType(propertiesValueNode), getPrintableDescription(propertiesValueNode, indent, "(no description)"), required))
			} else {
				doc.WriteString(fmt.Sprintf("\n\n%s- **`%s`** - %s - %s %s", indent, propertiesName, getPrintableType(propertiesValueNode), getPrintableDescription(propertiesValueNode, indent, "(no description)"), required))
			}

			propType := getScalarValue(propertiesValueNode, "type")
			if propType == "object" || propType == "array" || hasCombinators(propertiesValueNode) {
				newlevel := level + 1
				doc.WriteString(generateDoc(sidebarPosition, newlevel, propertiesValueNode, indent+"  ", getRequiredMapFromNode(propertiesValueNode)))
			}

			if examples := getNodeForKey(propertiesValueNode, "examples"); examples != nil && examples.Kind == yaml.SequenceNode {
				for _, example := range examples.Content {
					b, err := yaml.Marshal(example)
					if err != nil {
						panic(err)
					}
					doc.WriteString(fmt.Sprintf("\n\n```yaml\n%s```", string(b)))
				}
			}
		}
	} else if items := getNodeForKey(node, "items"); items != nil && items.Kind == yaml.MappingNode {
		items := getNodeForKey(node, "items")
		doc.WriteString(generateDoc(sidebarPosition, level, items, indent, getRequiredMapFromNode(items)))
	}

	// OneOf
	if oneOf := getNodeForKey(node, "oneOf"); oneOf != nil && oneOf.Kind == yaml.SequenceNode {
		// Skip oneOf processing for connectors at level 1 since we handle it in allOf
		if !(isConnector && level == 1) {
			if len(oneOf.Content) == 1 {
				doc.WriteString(generateDoc(sidebarPosition, level, oneOf.Content[0], indent, getRequiredMapFromNode(oneOf.Content[0])))
			} else {
				if level == 1 {
					doc.WriteString("\n\n## One of Properties Options")
					for _, item := range oneOf.Content {
						title := getScalarValue(item, "title")
						if title != "" {
							anchor := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
							doc.WriteString(fmt.Sprintf("\n- [%s](#%s)", title, anchor))
						}
					}
				}

				for i, item := range oneOf.Content {
					if hasType(item) || hasProperties(item) || hasCombinators(item) {
						title := getScalarValue(item, "title")
						if title != "" {
							doc.WriteString(fmt.Sprintf("\n\n#### Option %d: %s", i+1, title))
						} else {
							doc.WriteString(fmt.Sprintf("\n\n#### Option %d", i+1))
						}
						doc.WriteString(fmt.Sprintf("\n\n**Type:** %s\n\n**Description:** %s",
							getPrintableType(item),
							getPrintableDescription(item, indent, "(no description)")))
						doc.WriteString(generateDoc(sidebarPosition, level, item, indent+"  ", getRequiredMapFromNode(item)))
					}
				}
			}
		}
	}

	// AnyOf
	if anyOf := getNodeForKey(node, "anyOf"); anyOf != nil && anyOf.Kind == yaml.SequenceNode {
		for i, item := range anyOf.Content {
			if hasType(item) || hasProperties(item) || hasCombinators(item) {
				doc.WriteString(fmt.Sprintf("\n\n%s- **option %d** - %s - %s", indent, i+1, getPrintableType(item), getPrintableDescription(item, indent, "(no description)")))
				doc.WriteString(generateDoc(sidebarPosition, level, item, indent+"  ", getRequiredMapFromNode(item)))
			}
		}
	}

	// AllOf
	if allOf := getNodeForKey(node, "allOf"); allOf != nil && allOf.Kind == yaml.SequenceNode {
		// Special handling for connector allOf
		if isConnector && level == 1 {
			doc.WriteString("\n\n## Available Connector Types\n\n")
			doc.WriteString("Choose from the following connector types based on your data source:\n\n")

			// Find the oneOf section within allOf
			for _, item := range allOf.Content {
				oneOf := getNodeForKey(item, "oneOf")
				if oneOf == nil || oneOf.Kind != yaml.SequenceNode {
					continue
				}

				doc.WriteString("### OLAP Engines\n\n")
				doc.WriteString("- [**DuckDB**](#duckdb) - Embedded DuckDB engine (default)\n")
				doc.WriteString("- [**ClickHouse**](#clickhouse) - ClickHouse analytical database\n")
				doc.WriteString("- [**MotherDuck**](#motherduck) - MotherDuck cloud database\n")
				doc.WriteString("- [**Druid**](#druid) - Apache Druid\n")
				doc.WriteString("- [**Pinot**](#pinot) - Apache Pinot\n\n")

				doc.WriteString("### Data Warehouses\n\n")
				doc.WriteString("- [**Snowflake**](#snowflake) - Snowflake data warehouse\n")
				doc.WriteString("- [**BigQuery**](#bigquery) - Google BigQuery\n")
				doc.WriteString("- [**Redshift**](#redshift) - Amazon Redshift\n")
				doc.WriteString("- [**Athena**](#athena) - Amazon Athena\n\n")

				doc.WriteString("### Databases\n\n")
				doc.WriteString("- [**PostgreSQL**](#postgres) - PostgreSQL databases\n")
				doc.WriteString("- [**MySQL**](#mysql) - MySQL databases\n")
				doc.WriteString("- [**SQLite**](#sqlite) - SQLite databases\n\n")

				doc.WriteString("### Cloud Storage\n\n")
				doc.WriteString("- [**GCS**](#gcs) - Google Cloud Storage\n")
				doc.WriteString("- [**S3**](#s3) - Amazon S3 storage\n")
				doc.WriteString("- [**Azure**](#azure) - Azure Blob Storage\n\n")

				doc.WriteString("### Other\n\n")
				doc.WriteString("- [**HTTPS**](#https) - Public files via HTTP/HTTPS\n")
				doc.WriteString("- [**Salesforce**](#salesforce) - Salesforce data\n")
				doc.WriteString("- [**Slack**](#slack) - Slack data\n")
				doc.WriteString("- [**Local File**](#local_file) - Local file system\n\n")

				doc.WriteString(":::warning Security Recommendation\n")
				doc.WriteString("For all credential parameters (passwords, tokens, keys), use environment variables with the syntax `{{.env.<connector_type>.<parameter_name>}}`. This keeps sensitive data out of your YAML files and version control. See our [credentials documentation](/build/credentials/) for complete setup instructions.\n")
				doc.WriteString(":::\n\n")

				doc.WriteString("## Connector Details\n\n")
				break
			}
		}

		for _, item := range allOf.Content {
			// Skip oneOf items for connectors since we handle them separately
			if isConnector && getNodeForKey(item, "oneOf") != nil {
				continue
			}

			if hasIf(item) {
				ifNode := getNodeForKey(item, "if")
				title := getScalarValue(ifNode, "title")
				if level == 1 {
					doc.WriteString(fmt.Sprintf("\n\n## %s", title))
				} else {
					doc.WriteString(fmt.Sprintf("\n\n%s**%s**", indent, title))
				}
				thenNode := getNodeForKey(item, "then")
				doc.WriteString(generateDoc(sidebarPosition, level, thenNode, indent, getRequiredMapFromNode(item)))
			} else {
				doc.WriteString(generateDoc(sidebarPosition, level, item, indent, getRequiredMapFromNode(item)))
			}
		}
	}

	// Definitions (for connectors)
	if definitions := getNodeForKey(node, "definitions"); definitions != nil && definitions.Kind == yaml.MappingNode && isConnector && level == 1 {
		for i := 0; i < len(definitions.Content); i += 2 {
			connectorDef := definitions.Content[i+1]

			title := getScalarValue(connectorDef, "title")
			if title != "" {
				doc.WriteString(fmt.Sprintf("\n\n## %s\n\n", title))
				// Add description first
				description := getPrintableDescription(connectorDef, indent, "")
				if description != "" {
					doc.WriteString(fmt.Sprintf("%s\n\n", description))
				}
				// Add YAML example after description
				doc.WriteString(generateConnectorExample(title, connectorDef))
			}

			// Generate the connector definition documentation (properties, etc.) but skip the header
			// We need to process properties manually to avoid duplicate headers
			if properties := getNodeForKey(connectorDef, "properties"); properties != nil && properties.Kind == yaml.MappingNode {
				for j := 0; j < len(properties.Content); j += 2 {
					propName := properties.Content[j].Value
					propValue := properties.Content[j+1]
					required := ""
					if requiredFields := getRequiredMapFromNode(connectorDef); requiredFields[propName] {
						required = "_(required)_"
					}

					doc.WriteString(fmt.Sprintf("\n\n#### `%s`\n\n", propName))
					doc.WriteString(fmt.Sprintf("%s - %s %s",
						getPrintableType(propValue),
						getPrintableDescription(propValue, indent, "(no description)"),
						required))
				}
			}
		}
	}

	// Examples
	if examples := getNodeForKey(node, "examples"); examples != nil && examples.Kind == yaml.SequenceNode && currentLevel == 0 {
		doc.WriteString("\n\n## Examples")
		for _, example := range examples.Content {
			b, err := yaml.Marshal(example)
			if err != nil {
				panic(err)
			}
			doc.WriteString(fmt.Sprintf("\n\n```yaml\n%s```", string(b)))
		}
	}

	return doc.String()
}

func hasIf(node *yaml.Node) bool {
	return getNodeForKey(node, "if") != nil
}

func hasType(node *yaml.Node) bool {
	return getNodeForKey(node, "type") != nil
}

func hasProperties(node *yaml.Node) bool {
	return getNodeForKey(node, "properties") != nil
}

func hasCombinators(node *yaml.Node) bool {
	return getNodeForKey(node, "anyOf") != nil || getNodeForKey(node, "oneOf") != nil || getNodeForKey(node, "allOf") != nil
}

func generateConnectorExample(connectorType string, connectorDef *yaml.Node) string {
	if connectorDef == nil {
		return ""
	}

	var example strings.Builder
	example.WriteString("```yaml\n")
	example.WriteString("type: connector                                  # Must be `connector` (required)\n")

	// Get the driver from the schema and add it first
	driverAdded := false
	if driver := getNodeForKey(connectorDef, "driver"); driver != nil {
		if constVal := getScalarValue(driver, "const"); constVal != "" {
			example.WriteString(fmt.Sprintf("driver: %s                                   # Must be `%s` _(required)_\n\n", constVal, constVal))
			driverAdded = true
		}
	}

	// Fallback: if driver wasn't found, use the connector type name
	if !driverAdded {
		// Special case for MotherDuck which uses duckdb driver
		if connectorType == "MotherDuck" {
			example.WriteString("driver: duckdb                                   # Must be `duckdb` _(required)_\n\n")
		} else {
			example.WriteString(fmt.Sprintf("driver: %s                                   # Must be `%s` _(required)_\n\n", strings.ToLower(connectorType), strings.ToLower(connectorType)))
		}
	}

	// Get all properties from the schema
	if properties := getNodeForKey(connectorDef, "properties"); properties != nil && properties.Kind == yaml.MappingNode {
		for i := 0; i < len(properties.Content); i += 2 {
			propName := properties.Content[i].Value
			propValue := properties.Content[i+1]

			// Skip the driver property since we already added it
			if propName == "driver" {
				continue
			}

			// Get property description
			description := getPrintableDescription(propValue, "", "")
			if description == "" {
				description = "Property description"
			}

			// Get property type and generate appropriate example value
			propType := getScalarValue(propValue, "type")
			exampleValue := generateExampleValue(propName, propType, propValue)

			// Check if it's required
			required := ""
			if requiredFields := getRequiredMapFromNode(connectorDef); requiredFields[propName] {
				required = " _(required)_"
			}

			// Format the line with proper alignment
			example.WriteString(fmt.Sprintf("%s: %s", propName, exampleValue))

			// Add padding for alignment
			padding := 35 - len(propName) - len(exampleValue)
			if padding > 0 {
				example.WriteString(strings.Repeat(" ", padding))
			}

			example.WriteString(fmt.Sprintf("# %s%s\n", description, required))
		}
	}

	example.WriteString("```\n\n")
	return example.String()
}

func generateExampleValue(propName, propType string, propNode *yaml.Node) string {
	// Check for const values first
	if constVal := getScalarValue(propNode, "const"); constVal != "" {
		return fmt.Sprintf("%q", constVal)
	}

	// Check for enum values
	if enum := getNodeForKey(propNode, "enum"); enum != nil && enum.Kind == yaml.SequenceNode && len(enum.Content) > 0 {
		return fmt.Sprintf("%q", enum.Content[0].Value)
	}

	// Generate based on type and property name
	switch propType {
	case "string":
		// Generate contextual examples based on property name
		switch {
		case strings.Contains(propName, "key") && strings.Contains(propName, "secret"):
			return "\"myawssecretkey\""
		case strings.Contains(propName, "key") && strings.Contains(propName, "access"):
			return "\"myawsaccesskey\""
		case strings.Contains(propName, "token"):
			return "\"mytemporarytoken\""
		case strings.Contains(propName, "password"):
			return "\"mypassword\""
		case strings.Contains(propName, "username"):
			return "\"myusername\""
		case strings.Contains(propName, "host"):
			return "\"localhost\""
		case strings.Contains(propName, "port"):
			return "5432"
		case strings.Contains(propName, "database"):
			return "\"mydatabase\""
		case strings.Contains(propName, "bucket"):
			return "\"my-bucket\""
		case strings.Contains(propName, "region"):
			return "\"us-east-1\""
		case strings.Contains(propName, "workgroup"):
			return "\"primary\""
		case strings.Contains(propName, "dsn"):
			return "\"postgresql://user:pass@localhost:5432/db\""
		case strings.Contains(propName, "path"):
			return "\"md:my_db\""
		case strings.Contains(propName, "uri"):
			return "\"s3://bucket/path\""
		case strings.Contains(propName, "endpoint"):
			return "\"https://api.example.com\""
		case strings.Contains(propName, "account"):
			return "\"myaccount\""
		case strings.Contains(propName, "project"):
			return "\"myproject\""
		case strings.Contains(propName, "cluster"):
			return "\"mycluster\""
		case strings.Contains(propName, "role") && strings.Contains(propName, "arn"):
			return "\"arn:aws:iam::123456789012:role/MyRole\""
		case strings.Contains(propName, "session"):
			return "\"MySession\""
		case strings.Contains(propName, "external"):
			return "\"MyExternalID\""
		case strings.Contains(propName, "location"):
			return "\"s3://my-bucket/athena-output/\""
		case strings.Contains(propName, "init_sql"):
			return "|\n  INSTALL 'motherduck';\n  LOAD 'motherduck';\n  SET motherduck_token= '{{ .env.motherduck_token }}'"
		default:
			return "\"example_value\""
		}
	case "integer":
		return "123"
	case "boolean":
		// Check if there's a default value
		if defaultVal := getScalarValue(propNode, "default"); defaultVal != "" {
			return defaultVal
		}
		return "true"
	case "number":
		return "123.45"
	default:
		return "\"example_value\""
	}
}
