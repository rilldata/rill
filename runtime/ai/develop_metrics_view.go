package ai

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

const DevelopMetricsViewName = "develop_metrics_view"

type DevelopMetricsView struct {
	Runtime *runtime.Runtime
}

var _ Tool[*DevelopMetricsViewArgs, *DevelopMetricsViewResult] = (*DevelopMetricsView)(nil)

type DevelopMetricsViewArgs struct {
	Path   string `json:"path" jsonschema:"The path of a .yaml file in which to create or update a Rill metrics view definition."`
	Prompt string `json:"prompt" jsonschema:"Optional description of changes to make if editing an existing metrics view."`
	Model  string `json:"model" jsonschema:"Optional Rill model to derive from if creating a new metrics view."`
}

type DevelopMetricsViewResult struct {
	MetricsViewName string `json:"metrics_view_name" jsonschema:"The name of the developed Rill metrics view."`
}

func (t *DevelopMetricsView) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        DevelopMetricsViewName,
		Title:       "Develop Metrics View",
		Description: "Developer agent that creates or edits a single Rill metrics view.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Developing metrics view...",
			"openai/toolInvocation/invoked":  "Developed metrics view",
		},
	}
}

func (t *DevelopMetricsView) CheckAccess(ctx context.Context) (bool, error) {
	return checkDeveloperAgentAccess(ctx, t.Runtime)
}

func (t *DevelopMetricsView) Handler(ctx context.Context, args *DevelopMetricsViewArgs) (*DevelopMetricsViewResult, error) {
	// Validate input
	if !strings.HasPrefix(args.Path, "/") && args.Path == "" {
		args.Path = "/" + args.Path
	}

	// If a prompt is provided, use the agent-based editing flow
	if args.Prompt != "" {
		return t.handleWithAgent(ctx, args)
	}

	// Otherwise, use the original creation flow
	return t.handleCreation(ctx, args)
}

func (t *DevelopMetricsView) handleWithAgent(ctx context.Context, args *DevelopMetricsViewArgs) (*DevelopMetricsViewResult, error) {
	// Pre-invoke file listing to get context about existing metrics views
	session := GetSession(ctx)
	_, err := session.CallTool(ctx, RoleAssistant, ListFilesName, nil, &ListFilesArgs{})
	if err != nil {
		return nil, err
	}

	// Pre-invoke file read for the target file
	_, _ = session.CallTool(ctx, RoleAssistant, ReadFileName, nil, &ReadFileArgs{
		Path: args.Path,
	})
	if ctx.Err() != nil { // Ignore tool error since the file may not exist
		return nil, ctx.Err()
	}

	// Add the system prompt
	systemPrompt, err := t.systemPrompt(ctx)
	if err != nil {
		return nil, err
	}

	// Add the user prompt
	userPrompt, err := t.userPrompt(ctx, args)
	if err != nil {
		return nil, err
	}

	// Run an LLM tool call loop
	var response string
	err = session.Complete(ctx, "Metrics view developer loop", &response, &CompleteOptions{
		Messages: []*aiv1.CompletionMessage{
			NewTextCompletionMessage(RoleSystem, systemPrompt),
			NewTextCompletionMessage(RoleUser, userPrompt),
		},
		Tools:         []string{SearchFilesName, ReadFileName, WriteFileName},
		MaxIterations: 10,
		UnwrapCall:    true,
	})
	if err != nil {
		return nil, err
	}

	return &DevelopMetricsViewResult{
		MetricsViewName: fileutil.Stem(args.Path), // Get name from input path
	}, nil
}

func (t *DevelopMetricsView) handleCreation(ctx context.Context, args *DevelopMetricsViewArgs) (*DevelopMetricsViewResult, error) {
	// Resolve the table information
	connector, dialect, tbl, isModel, err := t.resolveTable(ctx, args.Model)
	if err != nil {
		return nil, err
	}

	// Determine if it's the default connector
	s := GetSession(ctx)
	inst, err := t.Runtime.Instance(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}
	isDefaultConnector := connector == inst.ResolveOLAPConnector()

	// Try to generate the YAML with AI

	// Generate
	var data string
	res, err := t.generateMetricsViewYAMLWithAI(ctx, s.InstanceID(), dialect.String(), connector, tbl, isDefaultConnector, isModel)
	if err == nil {
		data = res.data
	}

	// If we didn't manage to generate the YAML using AI, we fall back to the simple generator
	if data == "" {
		data, err = generateMetricsViewYAMLSimple(connector, tbl, isDefaultConnector, isModel)
		if err != nil {
			return nil, err
		}
	}

	// Write the file to the repo
	repo, release, err := t.Runtime.Repo(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}
	defer release()
	err = repo.Put(ctx, args.Path, strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	// Wait for it to reconcile
	ctrl, err := t.Runtime.Controller(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}
	err = ctrl.Reconcile(ctx, runtime.GlobalProjectParserName) // TODO: Only if not streaming
	if err != nil {
		return nil, err
	}
	select {
	case <-time.After(time.Millisecond * 500):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	err = ctrl.WaitUntilIdle(ctx, true)
	if err != nil {
		return nil, err
	}

	return &DevelopMetricsViewResult{
		MetricsViewName: fileutil.Stem(args.Path), // Get name from input path
	}, nil
}

func (t *DevelopMetricsView) systemPrompt(ctx context.Context) (string, error) {
	// Prepare template data
	session := GetSession(ctx)
	data := map[string]any{
		"ai_instructions": session.ProjectInstructions(),
	}

	// Generate the system prompt
	return executeTemplate(`<role>You are a data engineer agent specialized in developing metrics views in the Rill business intelligence platform.</role>

<concepts>
Rill is a "business intelligence as code" platform where all resources are defined using YAML files in a project directory.
For the purposes of your work, you will only deal with **metrics view** resources, which are YAML files that define dimensions and measures (SQL expressions and related metadata) for grouping and aggregating data in an underlying model (database table).
A metrics view in Rill is equivalent to a "cube", "semantic layer", or "metrics layer" in other BI platforms.
In Rill, when you write a file, the platform discovers and "reconciles" it immediately. For a metrics view, reconcile validates the metrics view definition and makes it available for querying.
</concepts>

<process>
At a high level, you should follow these steps:
1. Leverage the "read_file" tool to understand the file's current contents, if any (it may return a file not found error).
2. Generate a new or updated metrics view definition based on the user's prompt and save it to the requested path using the "write_file" tool.
3. The "write_file" tool will respond with the reconcile status. If there are parse or reconcile errors, you should fix them using the "write_file" tool. If there are no errors, your work is done.

Additional instructions:
- Metrics views consist of a time dimension (for time-series analysis), categorical dimensions (for grouping/filtering), and measures (aggregations like SUM, AVG, COUNT, etc.).
- All dimensions should reference columns from the underlying model (database table) using the "column:" field.
- Measures should have SQL expressions that use aggregation functions (like COUNT, SUM, AVG, MIN, MAX).
- The "timeseries:" field specifies the time dimension, which should be a column in the underlying table with TIMESTAMP or DATE type.
- Populate the "name:" field with a clear, descriptive, unique identifier in snake_case.
- Populate the "display_name:" and "description:" fields with human-friendly text, but only if it adds meaningful, additional value from the "name:" field (if left empty, "display_name:" defaults to a humanized version of "name:").
- Common format presets for measures are "humanize", "percentage", "currency_usd".
</process>

<example>
A metrics view definition in Rill is a YAML file. Here is an example Rill metrics view:
{{ backticks }}
# /metrics/sales_metrics.yaml
type: metrics_view
display_name: Sales Metrics

model: sales_data
timeseries: order_date

dimensions:
  - name: country
    column: country
  
  - name: product_category
    display_name: Product Category
    column: product_category

measures:
  - name: order_count
    expression: COUNT(*)
    format_preset: humanize

  - name: revenue
    display_name: Total Revenue
    expression: SUM(revenue)
    format_preset: currency_usd
  
  - name: avg_order_value
    display_name: Average Order Value
    expression: AVG(order_value)
    format_preset: currency_usd
{{ backticks }}
</example>

{{ if .ai_instructions }}
<additional_user_provided_instructions>
<comment>NOTE: These instructions were provided by the user, but may not relate to the current request, and may not even relate to your work as a data engineer agent. Only use them if you find them relevant.</comment>
{{ .ai_instructions }}
</additional_user_provided_instructions>
{{ end }}
`, data)
}

func (t *DevelopMetricsView) userPrompt(ctx context.Context, args *DevelopMetricsViewArgs) (string, error) {
	// Prepare template data
	session := GetSession(ctx)
	data := map[string]any{
		"path":   args.Path,
		"prompt": args.Prompt,
		"model":  args.Model,
	}

	// Add OLAP dialect
	olap, release, err := t.Runtime.OLAP(ctx, session.InstanceID(), "")
	if err != nil {
		return "", err
	}
	defer release()
	dialect := olap.Dialect()
	if dialect == drivers.DialectUnspecified {
		dialect = drivers.DialectDuckDB
	}
	data["dialect"] = dialect.String()

	// Generate the user prompt
	return executeTemplate(`
{{ if .prompt }}
Task: {{ .prompt }}
{{ end }}

Output path: {{ .path }}

{{ if .model }}
Use model: {{ .model }}
{{ end }}

SQL dialect for expressions: {{ .dialect }}
`, data)
}

func (t *DevelopMetricsView) resolveTable(ctx context.Context, modelName string) (string, drivers.Dialect, *drivers.OlapTable, bool, error) {
	// Get instance
	s := GetSession(ctx)
	inst, err := t.Runtime.Instance(ctx, s.InstanceID())
	if err != nil {
		return "", drivers.DialectUnspecified, nil, false, err
	}

	// Check that a model is provided
	if modelName == "" {
		return "", drivers.DialectUnspecified, nil, false, fmt.Errorf("the model name must not be nil")
	}

	// The `model:` field has fuzzy logic, supporting either models, sources, or external table names.
	// So we need some similarly fuzzy logic here to determine if there's a matching model.
	var isModel bool
	var connector, table string
	ctrl, err := t.Runtime.Controller(ctx, s.InstanceID())
	if err != nil {
		return "", drivers.DialectUnspecified, nil, false, err
	}
	model, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: modelName}, false)
	if err != nil && !errors.Is(err, drivers.ErrResourceNotFound) {
		return "", drivers.DialectUnspecified, nil, false, err
	}
	if model != nil {
		isModel = true
		connector = model.GetModel().State.ResultConnector
		table = model.GetModel().State.ResultTable
	}

	// If we did not find a model. We proceed to check if "model" references an external table in the default OLAP.
	if connector == "" {
		connector = inst.ResolveOLAPConnector()
	}
	if table == "" {
		table = modelName
	}

	// Connect to connector and check it's an OLAP db
	olap, release, err := t.Runtime.OLAP(ctx, s.InstanceID(), connector)
	if err != nil {
		return "", drivers.DialectUnspecified, nil, false, err
	}
	defer release()

	// Get table info
	tbl, err := olap.InformationSchema().Lookup(ctx, "", "", table)
	if err != nil {
		return "", drivers.DialectUnspecified, nil, false, fmt.Errorf("table not found: %w", err)
	}

	return connector, olap.Dialect(), tbl, isModel, nil
}

// generateMetricsViewYAMLWithres is a struct for the result of generateMetricsViewYAMLWithAI.
type generateMetricsViewYAMLWithres struct {
	data            string
	validMeasures   int
	invalidMeasures int
}

// generateMetricsViewYAMLWithAI attempts to generate a metrics view YAML definition from a table schema using AI.
// It validates that the result is a valid metrics view. Due to the unpredictable nature of AI (and chance of downtime), this function may error non-deterministically.
func (t *DevelopMetricsView) generateMetricsViewYAMLWithAI(ctx context.Context, instanceID, dialect, connector string, tbl *drivers.OlapTable, isDefaultConnector, isModel bool) (*generateMetricsViewYAMLWithres, error) {
	// Call AI service to infer a metrics view YAML
	var responseText string
	err := GetSession(ctx).Complete(ctx, "Generate metrics view", &responseText, &CompleteOptions{
		Messages: []*aiv1.CompletionMessage{
			NewTextCompletionMessage(RoleSystem, metricsViewYAMLSystemPrompt()),
			NewTextCompletionMessage(RoleUser, metricsViewYAMLUserPrompt(dialect, tbl.Name, tbl.Schema)),
		},
	})
	if err != nil {
		return nil, err
	}

	// The AI may produce Markdown output. Remove the code tags around the YAML.
	responseText = strings.TrimPrefix(responseText, "```yaml")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")

	// Parse the YAML structure
	var doc metricsViewYAML
	if err := yaml.Unmarshal([]byte(responseText), &doc); err != nil {
		return nil, fmt.Errorf("invalid metrics view YAML: %w", err)
	}

	// The AI only generates metrics. We fill in the other properties using the simple logic.
	doc.Type = "metrics_view"
	doc.TimeDimension = generateMetricsViewYAMLSimpleTimeDimension(tbl.Schema)
	doc.Dimensions = generateMetricsViewYAMLSimpleDimensions(tbl.Schema)
	for _, measure := range doc.Measures {
		// Apply the default format preset to measures (the AI doesn't set the format preset).
		measure.FormatPreset = "humanize"
	}
	if isModel {
		doc.Model = tbl.Name
	} else {
		if !isDefaultConnector {
			doc.Connector = connector
		}
		doc.Model = tbl.Name // Note: We also reference externally managed tables with `model:`. This is supported in the metrics view YAML.
		if tbl.Database != "" && !tbl.IsDefaultDatabase {
			doc.Database = tbl.Database
		}
		if tbl.DatabaseSchema != "" && !tbl.IsDefaultDatabaseSchema {
			doc.DatabaseSchema = tbl.DatabaseSchema
		}
	}

	// Create a map of column names, which are used to ensure the generated measure names do not collide with column names.
	columns := make(map[string]struct{})
	for _, f := range tbl.Schema.Fields {
		columns[f.Name] = struct{}{}
	}

	// Validate the generated measures (not validating other parts since those are not AI-generated)
	spec := &runtimev1.MetricsViewSpec{
		Connector:      connector,
		Database:       tbl.Database,
		DatabaseSchema: tbl.DatabaseSchema,
		Table:          tbl.Name,
	}
	for _, measure := range doc.Measures {
		// Prevent measure name collisions with column names
		if _, ok := columns[measure.Name]; !ok {
			measure.Name += "_measure"
		}

		spec.Measures = append(spec.Measures, &runtimev1.MetricsViewSpec_Measure{
			Name:         measure.Name,
			DisplayName:  measure.DisplayName,
			Expression:   measure.Expression,
			Type:         runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
			FormatPreset: measure.FormatPreset,
		})
	}

	e, err := executor.New(ctx, t.Runtime, instanceID, spec, !isModel, runtime.ResolvedSecurityOpen, 0)
	if err != nil {
		return nil, err
	}
	defer e.Close()
	validateResult, err := e.ValidateAndNormalizeMetricsView(ctx)
	if err != nil {
		return nil, err
	}

	// Remove all invalid measures. We do this in two steps to preserve the indexes returned in MeasureErrs:
	// First, we set the invalid measures to nil. Second, we remove the nil entries.
	var valid, invalid int
	for _, ie := range validateResult.MeasureErrs {
		doc.Measures[ie.Idx] = nil
	}
	for idx := 0; idx < len(doc.Measures); {
		if doc.Measures[idx] == nil {
			invalid++
			doc.Measures = slices.Delete(doc.Measures, idx, idx+1)
		} else {
			valid++
			idx++
		}
	}

	// If there are no valid measures left, bail
	if len(doc.Measures) == 0 {
		return nil, errors.New("no valid measures were generated")
	}

	// Render the updated YAML
	out, err := marshalMetricsViewYAML(&doc, true)
	if err != nil {
		return nil, err
	}

	return &generateMetricsViewYAMLWithres{
		data:            out,
		validMeasures:   valid,
		invalidMeasures: invalid,
	}, nil
}

// metricsViewYAMLSystemPrompt returns the static system prompt for the metrics view generation AI.
func metricsViewYAMLSystemPrompt() string {
	// We use the metricsViewYAML to generate a template of the YAML to guide the AI
	// NOTE: The YAML fields have omitempty, so the AI will only know about (and populate) the fields below. We will populate the rest manually after inference.
	template := metricsViewYAML{
		DisplayName: "<human-friendly display name based on the table name and column names>",
		Measures: []*metricsViewMeasureYAML{
			{
				Name:        "<unique name for the metric in snake case, such as average_sales>",
				DisplayName: "<short descriptive display name for the metric>",
				Expression:  "<SQL expression to calculate the KPI in the requested SQL dialect>",
				Description: "<short description of the metric>",
			},
		},
	}
	out, err := yaml.Marshal(template)
	if err != nil {
		panic(err)
	}

	prompt := fmt.Sprintf(`
You are an agent whose only task is to suggest relevant business metrics (KPIs) based on a table schema.
The metrics should be valid SQL aggregation expressions that use only the COUNT, SUM, MIN, MAX, and AVG functions.
Do not use any complex aggregations and do not use WHERE or FILTER in the metrics expressions.
Your output should only consist of valid YAML in the format below:

%s
`, string(out))

	return prompt
}

// metricsViewYAMLUserPrompt returns the dynamic user prompt for the metrics view generation AI.
func metricsViewYAMLUserPrompt(dialect, tblName string, schema *runtimev1.StructType) string {
	prompt := fmt.Sprintf(`
Give me up to 10 suggested metrics using the %q SQL dialect based on the table named %q, which has the following schema:
`, dialect, tblName)
	for _, field := range schema.Fields {
		prompt += fmt.Sprintf("- column=%s, type=%s\n", field.Name, field.Type.Code.String())
	}
	return prompt
}

// generateMetricsViewYAMLSimple generates a simple metrics view YAML definition from a table schema.
func generateMetricsViewYAMLSimple(connector string, tbl *drivers.OlapTable, isDefaultConnector, isModel bool) (string, error) {
	doc := &metricsViewYAML{
		Type:          "metrics_view",
		DisplayName:   identifierToDisplayName(tbl.Name),
		TimeDimension: generateMetricsViewYAMLSimpleTimeDimension(tbl.Schema),
		Dimensions:    generateMetricsViewYAMLSimpleDimensions(tbl.Schema),
		Measures:      generateMetricsViewYAMLSimpleMeasures(tbl),
	}

	if isModel {
		doc.Model = tbl.Name
	} else {
		if !isDefaultConnector {
			doc.Connector = connector
		}
		if tbl.Database != "" && !tbl.IsDefaultDatabase {
			doc.Database = tbl.Database
		}
		if tbl.DatabaseSchema != "" && !tbl.IsDefaultDatabaseSchema {
			doc.DatabaseSchema = tbl.DatabaseSchema
		}
		doc.Model = tbl.Name // Note: We also reference externally managed tables with `model:`. This is supported in the metrics view YAML.
	}

	return marshalMetricsViewYAML(doc, false)
}

func generateMetricsViewYAMLSimpleTimeDimension(schema *runtimev1.StructType) string {
	for _, f := range schema.Fields {
		switch f.Type.Code {
		case runtimev1.Type_CODE_TIMESTAMP, runtimev1.Type_CODE_TIME, runtimev1.Type_CODE_DATE:
			return f.Name
		}
	}
	return ""
}

func generateMetricsViewYAMLSimpleDimensions(schema *runtimev1.StructType) []*metricsViewDimensionYAML {
	var dims []*metricsViewDimensionYAML
	for _, f := range schema.Fields {
		switch f.Type.Code {
		case runtimev1.Type_CODE_BOOL, runtimev1.Type_CODE_STRING, runtimev1.Type_CODE_BYTES, runtimev1.Type_CODE_UUID, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIMESTAMP:
			dims = append(dims, &metricsViewDimensionYAML{
				Name:        f.Name,
				DisplayName: identifierToDisplayName(f.Name),
				Column:      f.Name,
			})
		}
	}
	return dims
}

func generateMetricsViewYAMLSimpleMeasures(tbl *drivers.OlapTable) []*metricsViewMeasureYAML {
	// Add a count measure
	var measures []*metricsViewMeasureYAML
	measures = append(measures, &metricsViewMeasureYAML{
		Name:         "total_records",
		DisplayName:  "Total records",
		Expression:   "COUNT(*)",
		Description:  "",
		FormatPreset: "humanize",
	})

	// Add sum measures for float columns
	for _, f := range tbl.Schema.Fields {
		switch f.Type.Code {
		case runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64:
			measures = append(measures, &metricsViewMeasureYAML{
				Name:         fmt.Sprintf("%s_sum", f.Name),
				DisplayName:  fmt.Sprintf("Sum of %s", identifierToDisplayName(f.Name)),
				Expression:   fmt.Sprintf("SUM(%s)", safeSQLName(f.Name)),
				Description:  "",
				FormatPreset: "humanize",
			})
		}
	}

	// Create a map of column names, which are used to ensure the generated measure names do not collide with column names.
	columns := make(map[string]struct{})
	for _, f := range tbl.Schema.Fields {
		columns[f.Name] = struct{}{}
	}

	// If a measure name collides with a table column name, append `_measure` until it's unique
	for _, m := range measures {
		for i := 0; i < 10; i++ {
			if _, ok := columns[m.Name]; !ok {
				break
			}
			m.Name += "_measure"
		}
	}

	return measures
}

// metricsViewYAML is a struct for generating a metrics view YAML file.
// We do not use the parser's structs since they are not suitable for generating pretty output YAML.
type metricsViewYAML struct {
	Type           string                      `yaml:"type,omitempty"`
	DisplayName    string                      `yaml:"display_name,omitempty"`
	Connector      string                      `yaml:"connector,omitempty"`
	Database       string                      `yaml:"database,omitempty"`
	DatabaseSchema string                      `yaml:"database_schema,omitempty"`
	Model          string                      `yaml:"model,omitempty"`
	TimeDimension  string                      `yaml:"timeseries,omitempty"`
	Dimensions     []*metricsViewDimensionYAML `yaml:"dimensions,omitempty"`
	Measures       []*metricsViewMeasureYAML   `yaml:"measures,omitempty"`
}

type metricsViewDimensionYAML struct {
	Name        string `yaml:"name"`
	DisplayName string `yaml:"display_name"`
	Column      string `yaml:"column"`
}

type metricsViewMeasureYAML struct {
	Name         string `yaml:"name"`
	DisplayName  string `yaml:"display_name"`
	Expression   string `yaml:"expression"`
	Description  string `yaml:"description"`
	FormatPreset string `yaml:"format_preset,omitempty"`
}

func marshalMetricsViewYAML(doc *metricsViewYAML, aiPowered bool) (string, error) {
	buf := new(bytes.Buffer)

	buf.WriteString("# Metrics view YAML\n")
	buf.WriteString("# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics-views\n")
	if aiPowered {
		buf.WriteString("# This file was generated using AI.\n")
	}
	buf.WriteString("\n")

	yamlBytes, err := yaml.Marshal(doc)
	if err != nil {
		return "", err
	}

	var rootNode yaml.Node
	if err := yaml.Unmarshal(yamlBytes, &rootNode); err != nil {
		return "", err
	}

	insertEmptyLinesInYaml(&rootNode)

	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	if err := enc.Encode(&rootNode); err != nil {
		return "", err
	}

	if err := enc.Close(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func insertEmptyLinesInYaml(node *yaml.Node) {
	for i := 0; i < len(node.Content); i++ {
		if node.Content[i].Kind == yaml.MappingNode {
			for j := 0; j < len(node.Content[i].Content); j += 2 {
				keyNode := node.Content[i].Content[j]
				valueNode := node.Content[i].Content[j+1]

				if keyNode.Value == "dimensions" || keyNode.Value == "measures" {
					keyNode.HeadComment = "\n"
				}
				if keyNode.Value == "type" {
					valueNode.LineComment = "\n\n"
				}
				insertEmptyLinesInYaml(valueNode)
			}
		} else if node.Content[i].Kind == yaml.SequenceNode {
			for j := 0; j < len(node.Content[i].Content); j++ {
				if node.Content[i].Content[j].Kind == yaml.MappingNode {
					node.Content[i].Content[j].HeadComment = "\n"
				}
			}
		}
	}
}

func identifierToDisplayName(s string) string {
	return strings.TrimLeft(cases.Title(language.English).String(strings.ReplaceAll(s, "_", " ")), " ")
}

var alphanumericUnderscoreRegexp = regexp.MustCompile("^[_a-zA-Z0-9]+$")

// safeSQLName escapes a SQL column identifier.
// If the name is simple (only contains alphanumeric characters and underscores), it does not escape the string.
// This is because the output is user-facing, so we want to return as simple names as possible.
func safeSQLName(name string) string {
	if name == "" {
		return name
	}
	if alphanumericUnderscoreRegexp.MatchString(name) {
		return name
	}
	return drivers.DialectDuckDB.EscapeIdentifier(name)
}
