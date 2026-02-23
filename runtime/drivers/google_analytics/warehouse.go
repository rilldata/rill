package google_analytics

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/api/analyticsdata/v1beta"
)

var tracer = otel.Tracer("github.com/rilldata/rill/runtime/drivers/google_analytics")

// pageSize is the number of rows requested per GA4 API call.
// The GA4 API supports up to 250,000 rows per request.
const pageSize int64 = 100000

var _ drivers.Warehouse = &connection{}

// Preset report definitions: each maps a report type to its dimensions and metrics.
type reportPreset struct {
	Dimensions []string
	Metrics    []string
}

var reportPresets = map[string]reportPreset{
	"traffic": {
		Dimensions: []string{"date", "sessionSource", "sessionMedium", "sessionCampaignName"},
		Metrics:    []string{"sessions", "activeUsers", "newUsers", "bounceRate"},
	},
	"pages": {
		Dimensions: []string{"date", "pagePath", "pageTitle"},
		Metrics:    []string{"screenPageViews", "activeUsers", "averageSessionDuration"},
	},
	"demographics": {
		Dimensions: []string{"date", "country", "city", "language"},
		Metrics:    []string{"activeUsers", "sessions", "newUsers"},
	},
	"events": {
		Dimensions: []string{"date", "eventName"},
		Metrics:    []string{"eventCount", "activeUsers", "eventCountPerUser"},
	},
}

type sourceProperties struct {
	ReportType string `mapstructure:"report_type"`
	Dimensions string `mapstructure:"dimensions"`
	Metrics    string `mapstructure:"metrics"`
	StartDate  string `mapstructure:"start_date"`
	EndDate    string `mapstructure:"end_date"`
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	if conf.ReportType == "" {
		return nil, fmt.Errorf("property 'report_type' is mandatory for connector \"google_analytics\"")
	}
	if conf.StartDate == "" {
		return nil, fmt.Errorf("property 'start_date' is mandatory for connector \"google_analytics\"")
	}
	if conf.EndDate == "" {
		conf.EndDate = "today"
	}
	return conf, nil
}

// resolveDimensionsAndMetrics resolves preset report types or parses custom comma-separated values.
func resolveDimensionsAndMetrics(src *sourceProperties) ([]string, []string, error) {
	if src.ReportType == "custom" {
		if src.Dimensions == "" {
			return nil, nil, fmt.Errorf("property 'dimensions' is required when report_type is 'custom'")
		}
		if src.Metrics == "" {
			return nil, nil, fmt.Errorf("property 'metrics' is required when report_type is 'custom'")
		}
		dims := splitAndTrim(src.Dimensions)
		mets := splitAndTrim(src.Metrics)
		return dims, mets, nil
	}

	preset, ok := reportPresets[src.ReportType]
	if !ok {
		return nil, nil, fmt.Errorf("unknown report_type %q; valid types: traffic, pages, demographics, events, custom", src.ReportType)
	}
	return preset.Dimensions, preset.Metrics, nil
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// QueryAsFiles implements drivers.Warehouse.
func (c *connection) QueryAsFiles(ctx context.Context, props map[string]any) (outIt drivers.FileIterator, outErr error) {
	ctx, span := tracer.Start(ctx, "connection.QueryAsFiles")
	defer func() {
		if outErr != nil {
			span.SetStatus(codes.Error, outErr.Error())
		}
		span.End()
	}()

	srcProps, err := parseSourceProperties(props)
	if err != nil {
		return nil, err
	}

	dims, mets, err := resolveDimensionsAndMetrics(srcProps)
	if err != nil {
		return nil, err
	}

	// Build GA4 API request
	gaDims := make([]*analyticsdata.Dimension, len(dims))
	for i, d := range dims {
		gaDims[i] = &analyticsdata.Dimension{Name: d}
	}

	gaMets := make([]*analyticsdata.Metric, len(mets))
	for i, m := range mets {
		gaMets[i] = &analyticsdata.Metric{Name: m}
	}

	dateRange := &analyticsdata.DateRange{
		StartDate: srcProps.StartDate,
		EndDate:   srcProps.EndDate,
	}

	// Determine property ID: source-level override or connection config
	propertyID := c.config.PropertyID
	if pid, ok := props["property_id"].(string); ok && pid != "" {
		propertyID = pid
	}
	if propertyID == "" {
		return nil, fmt.Errorf("property 'property_id' is required for connector \"google_analytics\"")
	}
	gaProperty := fmt.Sprintf("properties/%s", propertyID)

	svc, err := c.createService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GA4 service: %w", err)
	}

	// Fetch all pages and write each to a temp CSV file
	var tempFiles []string
	var offset int64
	header := append(dims, mets...)

	for {
		req := &analyticsdata.RunReportRequest{
			Dimensions: gaDims,
			Metrics:    gaMets,
			DateRanges: []*analyticsdata.DateRange{dateRange},
			Limit:      pageSize,
			Offset:     offset,
		}

		resp, err := svc.Properties.RunReport(gaProperty, req).Context(ctx).Do()
		if err != nil {
			// Clean up any files written so far
			for _, f := range tempFiles {
				os.Remove(f)
			}
			return nil, fmt.Errorf("GA4 RunReport failed: %w", err)
		}

		if len(resp.Rows) == 0 {
			break
		}

		filePath, err := writeRowsToCSV(header, resp.Rows)
		if err != nil {
			for _, f := range tempFiles {
				os.Remove(f)
			}
			return nil, fmt.Errorf("failed to write CSV: %w", err)
		}
		tempFiles = append(tempFiles, filePath)

		offset += int64(len(resp.Rows))
		if offset >= resp.RowCount {
			break
		}
	}

	if len(tempFiles) == 0 {
		return &fileIterator{done: true}, nil
	}

	return &fileIterator{files: tempFiles}, nil
}

// writeRowsToCSV writes GA4 report rows to a temporary CSV file.
func writeRowsToCSV(header []string, rows []*analyticsdata.Row) (string, error) {
	f, err := os.CreateTemp("", "rill-ga4-*.csv")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Write header
	if err := w.Write(header); err != nil {
		os.Remove(f.Name())
		return "", err
	}

	// Write data rows
	for _, row := range rows {
		record := make([]string, 0, len(row.DimensionValues)+len(row.MetricValues))
		for _, dv := range row.DimensionValues {
			record = append(record, dv.Value)
		}
		for _, mv := range row.MetricValues {
			record = append(record, mv.Value)
		}
		if err := w.Write(record); err != nil {
			os.Remove(f.Name())
			return "", err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		os.Remove(f.Name())
		return "", err
	}

	return f.Name(), nil
}
