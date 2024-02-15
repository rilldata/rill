package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
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

// GenerateMetricsViewFile generates a metrics view YAML file from a table in an OLAP database
func (s *Server) GenerateMetricsViewFile(ctx context.Context, req *runtimev1.GenerateMetricsViewFileRequest) (*runtimev1.GenerateMetricsViewFileResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.connector", req.Connector),
		attribute.String("args.table", req.Table),
		attribute.String("args.path", req.Path),
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
		req.Connector = inst.OLAPConnector
	}

	// Connect to connector and check it's an OLAP db
	handle, release, err := s.runtime.AcquireHandle(ctx, req.InstanceId, req.Connector)
	if err != nil {
		return nil, err
	}
	defer release()
	olap, ok := handle.AsOLAP(req.InstanceId)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "connector is not an OLAP connector")
	}

	// Get table info
	tbl, err := olap.InformationSchema().Lookup(ctx, req.Table)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "table not found")
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
		if modelState.Connector != req.Connector || modelState.Table != tbl.Name {
			// The model is not for this table. Ignore it.
			model = nil
		}
	}

	// Try to generate the YAML with AI
	var data string
	if req.UseAi {
		data, err = s.generateMetricsViewYAMLWithAI(ctx, req.InstanceId, olap.Dialect().String(), tbl.Name, model != nil, tbl.Schema)
		if err != nil {
			s.logger.Warn("failed to generate metrics view YAML using AI", zap.Error(err))
		}
	}

	// If we didn't manage to generate the YAML using AI, we fall back to the simple generator
	if data == "" {
		data, err = generateMetricsViewYAMLSimple(tbl.Name, model != nil, tbl.Schema)
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

	return &runtimev1.GenerateMetricsViewFileResponse{}, nil
}

// generateMetricsViewYAMLWithAI attempts to generate a metrics view YAML definition from a table schema using AI.
// It validates that the result is a valid metrics view. Due to the unpredictable nature of AI (and chance of downtime), this function may error non-deterministically.
func (s *Server) generateMetricsViewYAMLWithAI(ctx context.Context, instanceID, dialect, tblName string, isModel bool, schema *runtimev1.StructType) (string, error) {
	// Build messages
	msgs := []*drivers.CompletionMessage{
		{Role: "system", Data: metricsViewYAMLSystemPrompt(isModel)},
		{Role: "user", Data: metricsViewYAMLUserPrompt(dialect, tblName, schema)},
	}

	// Connect to the AI service configured for the instance
	ai, release, err := s.runtime.AI(ctx, instanceID)
	if err != nil {
		return "", err
	}
	defer release()

	// Call AI service to infer a metrics view YAML
	res, err := ai.Complete(ctx, msgs)
	if err != nil {
		return "", err
	}

	// TODO: Validate the YAML using the parser
	// TODO: Remove invalid dimensions and measures (use validation logic from the reconciler)
	// TODO: Add a comment in the output for whether it was generated with AI or static analysis
	// TODO: Set AvailableTimeZones? Maybe also Model explicitly?

	return res.Data, nil
}

func metricsViewYAMLSystemPrompt(isModel bool) string {
	template := metricsViewYAML{
		Title:         "<human-friendly title for the metrics set>",
		TimeDimension: "<name of the primary time dimension>",
		Dimensions: []metricsViewDimensionYAML{
			{
				Label:       "<short descriptive label for the dimension>",
				Column:      "<column name>",
				Description: "<short description of the dimension>",
			},
		},
		Measures: []metricsViewMeasureYAML{
			{
				Name:                "<unique name for the metric, such as average_sales>",
				Label:               "<short descriptive label for the metric>",
				Expression:          "<SQL expression to calculate the KPI in the dialect user asks for>",
				Description:         "<short Description of the metric>",
				FormatPreset:        "<should always be 'humanize'>",
				ValidPercentOfTotal: "<true if the metrics is summable otherwise false>",
			},
		},
	}

	if isModel {
		template.Model = "<the table name>"
	} else {
		template.Table = "<the table name>"
	}

	out, err := yaml.Marshal(template)
	if err != nil {
		panic(err)
	}

	return `You are an agent whose only task is to suggest relevant business KPI metrics based on a table schema in a valid YAML file and output it in a format below:` + "\n\n" + string(out)
}

func metricsViewYAMLUserPrompt(dialect, tblName string, schema *runtimev1.StructType) string {
	prompt := fmt.Sprintf(`
Give me up to 10 suggested metrics in a YAML file.
Use %q SQL dialect for the metric expressions.
Use only COUNT, SUM, MIN, MAX, AVG as aggregation functions.
Do not use any complex aggregations.
Do not use WHERE or FILTER in metrics definition.
The table name is %q.
Here is my table schema:
`, dialect, tblName)
	for _, field := range schema.Fields {
		prompt += fmt.Sprintf("- column=%s, type=%s\n", field.Name, field.Type.Code.String())
	}
	return prompt
}

// generateMetricsViewYAMLSimple generates a simple metrics view YAML definition from a table schema.
func generateMetricsViewYAMLSimple(tblName string, isModel bool, schema *runtimev1.StructType) (string, error) {
	doc := &metricsViewYAML{
		Title: identifierToTitle(tblName),
	}

	if isModel {
		doc.Model = tblName
	} else {
		doc.Table = tblName
	}

	for _, f := range schema.Fields {
		if isTimestampType(f.Type.Code) {
			doc.TimeDimension = f.Name
			break
		}
	}

	doc.Measures = append(doc.Measures, metricsViewMeasureYAML{
		Name:                "total_records",
		Label:               "Total records",
		Expression:          "COUNT(*)",
		Description:         "",
		FormatPreset:        "humanize",
		ValidPercentOfTotal: true,
	})

	for _, f := range schema.Fields {
		if isDimensionType(f.Type.Code) {
			doc.Dimensions = append(doc.Dimensions, metricsViewDimensionYAML{
				Label:       identifierToTitle(f.Name),
				Column:      f.Name,
				Description: "",
			})
		}
		if isMeasureType(f.Type.Code) {
			doc.Measures = append(doc.Measures, metricsViewMeasureYAML{
				Name:                f.Name,
				Label:               fmt.Sprintf("Sum of %s", identifierToTitle(f.Name)),
				Expression:          fmt.Sprintf("SUM(%s)", f.Name),
				Description:         "",
				FormatPreset:        "humanize",
				ValidPercentOfTotal: true,
			})
		}
	}

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

	out, err := yaml.Marshal(doc)
	if err != nil {
		return "", err
	}

	docstring := "# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files."
	res := docstring + "\n\n" + string(out)
	return res, nil
}

// metricsViewYAML is a struct for generating a metrics view YAML file.
// We do not use the parser's structs since they are not suitable for generating pretty output YAML.
type metricsViewYAML struct {
	Title              string `yaml:"title"`
	Model              string `yaml:"model,omitempty"`
	Table              string `yaml:"table,omitempty"`
	TimeDimension      string `yaml:"timeseries"`
	Dimensions         []metricsViewDimensionYAML
	Measures           []metricsViewMeasureYAML
	AvailableTimeZones []string `yaml:"available_time_zones,omitempty"`
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

func identifierToTitle(s string) string {
	return cases.Title(language.English).String(strings.ReplaceAll(s, "_", " "))
}

func isTimestampType(t runtimev1.Type_Code) bool {
	switch t {
	case runtimev1.Type_CODE_TIMESTAMP, runtimev1.Type_CODE_TIME, runtimev1.Type_CODE_DATE:
		return true
	}
	return false
}

func isDimensionType(t runtimev1.Type_Code) bool {
	switch t {
	case runtimev1.Type_CODE_BOOL, runtimev1.Type_CODE_STRING, runtimev1.Type_CODE_BYTES, runtimev1.Type_CODE_UUID:
		return true
	}
	return false
}

func isMeasureType(t runtimev1.Type_Code) bool {
	switch t {
	case runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64:
		return true
	}
	return false
}
