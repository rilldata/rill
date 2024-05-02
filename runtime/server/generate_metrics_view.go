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

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
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
		attribute.String("args.connector", req.Connector),
		attribute.String("args.database", req.Database),
		attribute.String("args.database_schema", req.DatabaseSchema),
		attribute.String("args.table", req.Table),
		attribute.String("args.path", req.Path),
		attribute.Bool("args.use_ai", req.UseAi),
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	// Must have edit permissions on the repo
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	// Get instance
	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
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

	// The table may have been created by a model. Search for a model with the same name in the same connector.
	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	model, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: tbl.Name}, false)
	if err != nil && !errors.Is(err, drivers.ErrResourceNotFound) {
		return nil, err
	}
	if model != nil {
		modelState := model.GetModel().State
		if modelState.ResultConnector != req.Connector || modelState.ResultTable != tbl.Name {
			// The model is not for this table. Ignore it.
			model = nil
		}
	}

	// Try to generate the YAML with AI
	var data string
	var aiSucceeded bool
	if req.UseAi {
		// Generate
		start := time.Now()
		res, err := s.generateMetricsViewYAMLWithAI(ctx, req.InstanceId, olap.Dialect().String(), req.Connector, tbl, isDefaultConnector, model != nil)
		if err != nil {
			s.logger.Warn("failed to generate metrics view YAML using AI", zap.Error(err))
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
		data, err = generateMetricsViewYAMLSimple(req.Connector, tbl, isDefaultConnector, model != nil, tbl.Schema)
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
func (s *Server) generateMetricsViewYAMLWithAI(ctx context.Context, instanceID, dialect, connector string, tbl *drivers.Table, isDefaultConnector, isModel bool) (*generateMetricsViewYAMLWithres, error) {
	// Build messages
	msgs := []*drivers.CompletionMessage{
		{Role: "system", Data: metricsViewYAMLSystemPrompt()},
		{Role: "user", Data: metricsViewYAMLUserPrompt(dialect, tbl.Name, tbl.Schema)},
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
	res, err := ai.Complete(ctx, msgs)
	if err != nil {
		return nil, err
	}

	// The AI may produce Markdown output. Remove the code tags around the YAML.
	res.Data = strings.TrimPrefix(res.Data, "```yaml")
	res.Data = strings.TrimPrefix(res.Data, "```")
	res.Data = strings.TrimSuffix(res.Data, "```")

	// Parse the YAML structure
	var doc metricsViewYAML
	if err := yaml.Unmarshal([]byte(res.Data), &doc); err != nil {
		return nil, fmt.Errorf("invalid metrics view YAML: %w", err)
	}

	// The AI only generates metrics. We fill in the other properties using the simple logic.
	doc.Type = "metrics_view"
	doc.TimeDimension = generateMetricsViewYAMLSimpleTimeDimension(tbl.Schema)
	doc.Dimensions = generateMetricsViewYAMLSimpleDimensions(tbl.Schema)
	if isModel {
		doc.Model = tbl.Name
	} else {
		if !isDefaultConnector {
			doc.Connector = connector
		}
		doc.Table = tbl.Name
		if tbl.Database != "" && !tbl.IsDefaultDatabase {
			doc.Database = tbl.Database
		}
		if tbl.DatabaseSchema != "" && !tbl.IsDefaultDatabaseSchema {
			doc.DatabaseSchema = tbl.DatabaseSchema
		}
	}

	// Validate the generated measures (not validating other parts since those are not AI-generated)
	spec := &runtimev1.MetricsViewSpec{
		Connector:      connector,
		Database:       tbl.Database,
		DatabaseSchema: tbl.DatabaseSchema,
		Table:          tbl.Name,
	}
	for _, measure := range doc.Measures {
		spec.Measures = append(spec.Measures, &runtimev1.MetricsViewSpec_MeasureV2{
			Name:         measure.Name,
			Label:        measure.Label,
			Expression:   measure.Expression,
			FormatPreset: measure.FormatPreset,
		})
	}
	validateResult, err := s.runtime.ValidateMetricsView(ctx, instanceID, spec)
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
		Title: "<human-friendly title based on the table name and column names>",
		Measures: []*metricsViewMeasureYAML{
			{
				Name:                "<unique name for the metric in snake case, such as average_sales>",
				Label:               "<short descriptive label for the metric>",
				Expression:          "<SQL expression to calculate the KPI in the requested SQL dialect>",
				Description:         "<short description of the metric>",
				FormatPreset:        "<should always be 'humanize'>",
				ValidPercentOfTotal: "<true if the metric is summable otherwise false>",
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
func generateMetricsViewYAMLSimple(connector string, tbl *drivers.Table, isDefaultConnector, isModel bool, schema *runtimev1.StructType) (string, error) {
	doc := &metricsViewYAML{
		Type:          "metrics_view",
		Title:         identifierToTitle(tbl.Name),
		TimeDimension: generateMetricsViewYAMLSimpleTimeDimension(schema),
		Dimensions:    generateMetricsViewYAMLSimpleDimensions(schema),
		Measures:      generateMetricsViewYAMLSimpleMeasures(schema),
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
		doc.Table = tbl.Name
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
		case runtimev1.Type_CODE_BOOL, runtimev1.Type_CODE_STRING, runtimev1.Type_CODE_BYTES, runtimev1.Type_CODE_UUID:
			dims = append(dims, &metricsViewDimensionYAML{
				Label:       identifierToTitle(f.Name),
				Column:      f.Name,
				Description: "",
			})
		}
	}
	return dims
}

func generateMetricsViewYAMLSimpleMeasures(schema *runtimev1.StructType) []*metricsViewMeasureYAML {
	var measures []*metricsViewMeasureYAML
	measures = append(measures, &metricsViewMeasureYAML{
		Name:                "total_records",
		Label:               "Total records",
		Expression:          "COUNT(*)",
		Description:         "",
		FormatPreset:        "humanize",
		ValidPercentOfTotal: true,
	})
	for _, f := range schema.Fields {
		switch f.Type.Code {
		case runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64:
			measures = append(measures, &metricsViewMeasureYAML{
				Name:                f.Name,
				Label:               fmt.Sprintf("Sum of %s", identifierToTitle(f.Name)),
				Expression:          fmt.Sprintf("SUM(%s)", safeSQLName(f.Name)),
				Description:         "",
				FormatPreset:        "humanize",
				ValidPercentOfTotal: true,
			})
		}
	}
	return measures
}

// metricsViewYAML is a struct for generating a metrics view YAML file.
// We do not use the parser's structs since they are not suitable for generating pretty output YAML.
type metricsViewYAML struct {
	Type                string                      `yaml:"type,omitempty"`
	Title               string                      `yaml:"title,omitempty"`
	Connector           string                      `yaml:"connector,omitempty"`
	Database            string                      `yaml:"database,omitempty"`
	DatabaseSchema      string                      `yaml:"database_schema,omitempty"`
	Table               string                      `yaml:"table,omitempty"`
	Model               string                      `yaml:"model,omitempty"`
	TimeDimension       string                      `yaml:"timeseries,omitempty"`
	Dimensions          []*metricsViewDimensionYAML `yaml:"dimensions,omitempty"`
	Measures            []*metricsViewMeasureYAML   `yaml:"measures,omitempty"`
	AvailableTimeZones  []string                    `yaml:"available_time_zones,omitempty"`
	AvailableTimeRanges []string                    `yaml:"available_time_ranges,omitempty"`
}

type metricsViewDimensionYAML struct {
	Label       string
	Column      string
	Description string
}

type metricsViewMeasureYAML struct {
	Name                string
	Label               string
	Expression          string
	Description         string
	FormatPreset        string `yaml:"format_preset"`
	ValidPercentOfTotal any    `yaml:"valid_percent_of_total"`
}

func marshalMetricsViewYAML(doc *metricsViewYAML, aiPowered bool) (string, error) {
	doc.AvailableTimeZones = []string{
		"America/Los_Angeles",
		"America/Chicago",
		"America/New_York",
		"Europe/London",
		"Europe/Paris",
		"Asia/Jerusalem",
		"Europe/Moscow",
		"Asia/Kolkata",
		"Asia/Shanghai",
		"Asia/Tokyo",
		"Australia/Sydney",
	}

	doc.AvailableTimeRanges = []string{
		"PT6H",
		"PT24H",
		"P7D",
		"P14D",
		"P4W",
		"P3M",
		"P12M",
		"rill-TD",
		"rill-WTD",
		"rill-MTD",
		"rill-QTD",
		"rill-YTD",
		"rill-PDC",
		"rill-PWC",
		"rill-PMC",
		"rill-PQC",
		"rill-PYC",
	}

	buf := new(bytes.Buffer)

	buf.WriteString("# Dashboard YAML\n")
	buf.WriteString("# Reference documentation: https://docs.rilldata.com/reference/project-files/dashboards\n")
	if aiPowered {
		buf.WriteString("# This file was generated using AI.\n")
	}
	buf.WriteString("\n")

	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return "", err
	}
	if err := enc.Close(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func identifierToTitle(s string) string {
	return cases.Title(language.English).String(strings.ReplaceAll(s, "_", " "))
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
	return name
}
