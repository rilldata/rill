package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

// aiGenerateTimeout is the maximum time to wait for the AI to generate a file.
// If the AI takes longer than this, we should use fallback logic.
const aiGenerateTimeout = 30 * time.Second

// GenerateMetricsViewFile generates a metrics view YAML file from a table in an OLAP database
func (s *Server) GenerateMetricsViewFile(ctx context.Context, req *runtimev1.GenerateMetricsViewFileRequest) (*runtimev1.GenerateMetricsViewFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.model", req.Model),
		attribute.String("args.connector", req.Connector),
		attribute.String("args.database", req.Database),
		attribute.String("args.database_schema", req.DatabaseSchema),
		attribute.String("args.table", req.Table),
		attribute.String("args.path", req.Path),
		attribute.Bool("args.use_ai", req.UseAi),
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	// Must have edit permissions on the repo
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
		return nil, ErrForbidden
	}

	// Get instance
	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	// Check that only one of model or table is provided
	if req.Model != "" && req.Table != "" {
		return nil, status.Error(codes.InvalidArgument, "only one of model or table can be provided")
	} else if req.Model == "" && req.Table == "" {
		return nil, status.Error(codes.InvalidArgument, "either model or table must be provided")
	}

	// The `model:` field has fuzzy logic, supporting either models, sources, or external table names.
	// So we need some similarly fuzzy logic here to determine if there's a matching model.
	var modelFound bool
	var modelConnector, modelTable string
	searchName := req.Table
	if req.Model != "" {
		searchName = req.Model
	}
	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	model, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: searchName}, false)
	if err != nil && !errors.Is(err, drivers.ErrResourceNotFound) {
		return nil, err
	}
	if model != nil {
		modelFound = true
		modelConnector = model.GetModel().State.ResultConnector
		modelTable = model.GetModel().State.ResultTable
	} else {
		// Check if it's a source
		source, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindSource, Name: searchName}, false)
		if err != nil && !errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, err
		}
		if source != nil {
			modelFound = true
			modelConnector = source.GetSource().State.Connector
			modelTable = source.GetSource().State.Table
		}
	}

	// Depending on the result, we populate req.Connector and req.Table which we use subsequently.
	if modelFound {
		// We found a matching model.
		if req.Model != "" { // We found req.Model
			req.Connector = modelConnector
			req.Table = modelTable
		} else if (req.Connector == "" || req.Connector == modelConnector) && strings.EqualFold(req.Table, modelTable) { // We found a model with the same name as req.Table
			req.Connector = modelConnector
			req.Table = modelTable
		} else { // We found a model that doesn't match the request.
			modelFound = false
		}
	} else {
		// We did not find a model. We proceed to check if "model" references an external table in the default OLAP.
		if req.Model != "" {
			req.Connector = ""
			req.Table = req.Model
		}
	}

	// If a connector is not provided, default to the instance's OLAP connector
	if req.Connector == "" {
		req.Connector = inst.ResolveOLAPConnector()
	}
	isDefaultConnector := req.Connector == inst.ResolveOLAPConnector()

	// Connect to connector and check it's an OLAP db
	olap, release, err := s.runtime.OLAP(ctx, req.InstanceId, req.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	// Get table info
	tbl, err := olap.InformationSchema().Lookup(ctx, req.Database, req.DatabaseSchema, req.Table)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "table not found: %s", err)
	}

	// Try to generate the YAML with AI
	var data string
	var aiSucceeded bool
	if req.UseAi {
		// Generate
		start := time.Now()
		res, err := s.generateMetricsViewYAMLWithAI(ctx, req.InstanceId, olap.Dialect().String(), req.Connector, tbl, isDefaultConnector, modelFound)
		if err != nil {
			s.logger.Warn("failed to generate metrics view YAML using AI", zap.Error(err), observability.ZapCtx(ctx))
		} else {
			data = res.data
			aiSucceeded = true
		}

		// Emit event
		attrs := []attribute.KeyValue{attribute.Int("table_column_count", len(tbl.Schema.Fields))}
		attrs = append(attrs,
			attribute.Bool("succeeded", aiSucceeded),
			attribute.Int64("elapsed_ms", time.Since(start).Milliseconds()),
		)
		if res != nil {
			attrs = append(attrs,
				attribute.Int("valid_measures_count", res.validMeasures),
				attribute.Int("invalid_measures_count", res.invalidMeasures),
			)
		}
		if err != nil {
			attrs = append(attrs, attribute.String("error", err.Error()))
		}
		s.activity.Record(ctx, activity.EventTypeLog, "ai_generated_metrics_view_yaml", attrs...)
	}

	// If we didn't manage to generate the YAML using AI, we fall back to the simple generator
	if data == "" {
		data, err = generateMetricsViewYAMLSimple(req.Connector, tbl, isDefaultConnector, modelFound)
		if err != nil {
			return nil, err
		}
	}

	// Write the file to the repo
	repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	defer release()
	err = repo.Put(ctx, req.Path, strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	return &runtimev1.GenerateMetricsViewFileResponse{AiSucceeded: aiSucceeded}, nil
}

// generateMetricsViewYAMLWithres is a struct for the result of generateMetricsViewYAMLWithAI.
type generateMetricsViewYAMLWithres struct {
	data            string
	validMeasures   int
	invalidMeasures int
}

// generateMetricsViewYAMLWithAI attempts to generate a metrics view YAML definition from a table schema using AI.
// It validates that the result is a valid metrics view. Due to the unpredictable nature of AI (and chance of downtime), this function may error non-deterministically.
func (s *Server) generateMetricsViewYAMLWithAI(ctx context.Context, instanceID, dialect, connector string, tbl *drivers.OlapTable, isDefaultConnector, isModel bool) (*generateMetricsViewYAMLWithres, error) {
	// Build messages
	systemPrompt := metricsViewYAMLSystemPrompt()
	userPrompt := metricsViewYAMLUserPrompt(dialect, tbl.Name, tbl.Schema)

	msgs := []*aiv1.CompletionMessage{
		{
			Role: "system",
			Content: []*aiv1.ContentBlock{
				{
					BlockType: &aiv1.ContentBlock_Text{
						Text: systemPrompt,
					},
				},
			},
		},
		{
			Role: "user",
			Content: []*aiv1.ContentBlock{
				{
					BlockType: &aiv1.ContentBlock_Text{
						Text: userPrompt,
					},
				},
			},
		},
	}

	// Connect to the AI service configured for the instance
	ai, release, err := s.runtime.AI(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, aiGenerateTimeout)
	defer cancel()

	// Call AI service to infer a metrics view YAML
	res, err := ai.Complete(ctx, &drivers.CompleteOptions{
		Messages: msgs,
	})
	if err != nil {
		return nil, err
	}

	// Extract text from content blocks
	var responseText string
	for _, block := range res.Message.Content {
		switch blockType := block.GetBlockType().(type) {
		case *aiv1.ContentBlock_Text:
			if text := blockType.Text; text != "" {
				responseText += text
			}
		default:
			// For metrics view generation, we only expect text responses
			return nil, fmt.Errorf("unexpected content block type in AI response: %T", blockType)
		}
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
	doc.Version = 1
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

	e, err := executor.New(ctx, s.runtime, instanceID, spec, !isModel, runtime.ResolvedSecurityOpen, 0)
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
		Version:       1,
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
	Version        int                         `yaml:"version,omitempty"`
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
