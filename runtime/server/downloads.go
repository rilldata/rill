package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func (s *Server) Export(ctx context.Context, req *runtimev1.ExportRequest) (*runtimev1.ExportResponse, error) {
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadMetrics) {
		return nil, ErrForbidden
	}

	tkn, err := s.generateDownloadToken(req, auth.GetClaims(ctx, req.InstanceId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate download token: %s", err.Error())
	}

	out := fmt.Sprintf("/v1/download?token=%s", tkn)

	return &runtimev1.ExportResponse{
		DownloadUrlPath: out,
	}, nil
}

func (s *Server) ExportReport(ctx context.Context, req *runtimev1.ExportReportRequest) (*runtimev1.ExportReportResponse, error) {
	c, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get controller: %s", err.Error())
	}

	res, err := c.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindReport, Name: req.Report}, false)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get report: %s", err.Error())
	}

	r, access, err := s.runtime.ApplySecurityPolicy(ctx, req.InstanceId, auth.GetClaims(ctx, req.InstanceId), res)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if !access {
		return nil, status.Error(codes.NotFound, "resource not found")
	}

	if r.GetReport() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "resource is not a report")
	}

	rep := r.GetReport()
	t := req.ExecutionTime.AsTime()

	qry, err := queries.ProtoFromJSON(rep.Spec.QueryName, rep.Spec.QueryArgsJson, &t)
	if err != nil {
		return nil, fmt.Errorf("failed to build export request: %w", err)
	}

	// Parse contextual information from the report annotations.
	var originDashboard *runtimev1.ResourceName
	var originURL string
	if rep.Spec.Annotations != nil {
		if explore, ok := rep.Spec.Annotations["explore"]; ok && explore != "" {
			originDashboard = &runtimev1.ResourceName{Kind: runtime.ResourceKindExplore, Name: explore}
		}
		if canvas, ok := rep.Spec.Annotations["canvas"]; ok && canvas != "" {
			originDashboard = &runtimev1.ResourceName{Kind: runtime.ResourceKindCanvas, Name: canvas}
		}
		if openPath, ok := rep.Spec.Annotations["web_open_path"]; ok && openPath != "" {
			var err error
			originURL, err = url.JoinPath(req.OriginBaseUrl, openPath)
			if err != nil {
				s.logger.Error("ExportReport: failed to join origin URL path", zap.Error(err), zap.String("open_path", openPath), zap.String("origin_base_url", req.OriginBaseUrl), zap.String("report", req.Report), observability.ZapCtx(ctx))
			}
		}
	}

	// Reports delivered in "creator mode" are accessed using claims that contain AdditionalRules that grant transitive access to the report.
	// The transitive rule resolves into a RowFilter rule for the underlying metrics view that contains the filters in the query above.
	// This means that the filters will be applied twice, which can impact the accuracy of subquery filters.
	// On the frontend, this is handled by not adding the report's filters to the dashboard when it's opened.
	// On the backend however, since we know the download token is scoped to a specific query, it is easier to just create a download token that doesn't have the transitive access rule.
	// This ensures the report's filters don't get applied twice in the query.
	claims := auth.GetClaims(ctx, req.InstanceId)
	downloadClaims := &runtime.SecurityClaims{
		UserID:          claims.UserID,
		UserAttributes:  claims.UserAttributes,
		Permissions:     claims.Permissions,
		AdditionalRules: nil, // Will be populated below
		SkipChecks:      claims.SkipChecks,
	}
	for _, r := range claims.AdditionalRules {
		ta := r.GetTransitiveAccess()
		if ta == nil {
			continue
		}
		if ta.Resource.Kind == runtime.ResourceKindReport && strings.EqualFold(ta.Resource.Name, req.Report) {
			continue
		}
		downloadClaims.AdditionalRules = append(downloadClaims.AdditionalRules, r)
	}

	// Note - We are passing caller's user attributes to generateDownloadToken which may not always be the creator's attributes in case of external user's magic token. This is different from the alerts use case.
	tkn, err := s.generateDownloadToken(&runtimev1.ExportRequest{
		InstanceId:      req.InstanceId,
		Limit:           int64(rep.Spec.ExportLimit),
		Format:          rep.Spec.ExportFormat,
		Query:           qry,
		IncludeHeader:   rep.Spec.ExportIncludeHeader,
		OriginDashboard: originDashboard,
		OriginUrl:       originURL,
		ExecutionTime:   valOrNullTime(t),
	}, downloadClaims)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate download token: %s", err.Error())
	}

	out := fmt.Sprintf("/v1/download?token=%s", tkn)

	return &runtimev1.ExportReportResponse{
		DownloadUrlPath: out,
	}, nil
}

func (s *Server) downloadHandler(w http.ResponseWriter, req *http.Request) {
	rawTkn := req.URL.Query().Get("token")
	request, claims, err := s.parseDownloadToken(rawTkn)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse download token: %s", err.Error()), http.StatusBadRequest)
		return
	}

	var execTime *time.Time
	if request.ExecutionTime != nil && request.ExecutionTime.IsValid() {
		t := request.ExecutionTime.AsTime()
		execTime = &t
	}

	var q runtime.Query
	switch v := request.Query.Query.(type) {
	case *runtimev1.Query_MetricsViewAggregationRequest:
		r := v.MetricsViewAggregationRequest

		var limitPtr *int64
		limit := s.resolveExportLimit(request.Limit, r.Limit)
		if limit != 0 {
			limitPtr = &limit
		}

		tr := r.TimeRange
		if r.TimeStart != nil || r.TimeEnd != nil {
			tr = &runtimev1.TimeRange{
				Start: r.TimeStart,
				End:   r.TimeEnd,
			}
		}

		q = &queries.MetricsViewAggregation{
			MetricsViewName:     r.MetricsView,
			Dimensions:          r.Dimensions,
			Measures:            r.Measures,
			Sort:                r.Sort,
			TimeRange:           tr,
			ComparisonTimeRange: r.ComparisonTimeRange,
			Where:               r.Where,
			Having:              r.Having,
			Filter:              r.Filter,
			Limit:               limitPtr,
			Offset:              r.Offset,
			PivotOn:             r.PivotOn,
			SecurityClaims:      claims,
			Aliases:             r.Aliases,
			Exact:               r.Exact,
			Rows:                r.Rows,
			ExecutionTime:       execTime,
		}
	case *runtimev1.Query_MetricsViewToplistRequest:
		r := v.MetricsViewToplistRequest

		var limitPtr *int64
		limit := s.resolveExportLimit(request.Limit, r.Limit)
		if limit != 0 {
			limitPtr = &limit
		}

		q = &queries.MetricsViewToplist{
			MetricsViewName: r.MetricsViewName,
			DimensionName:   r.DimensionName,
			MeasureNames:    r.MeasureNames,
			TimeStart:       r.TimeStart,
			TimeEnd:         r.TimeEnd,
			Limit:           limitPtr,
			Sort:            r.Sort,
			Where:           r.Where,
			Filter:          r.Filter,
			Having:          r.Having,
			SecurityClaims:  claims,
		}
	case *runtimev1.Query_MetricsViewRowsRequest:
		r := v.MetricsViewRowsRequest
		mv, security, err := resolveMVAndSecurityFromAttributes(req.Context(), s.runtime, request.InstanceId, r.MetricsViewName, claims)
		if err != nil {
			if errors.Is(err, ErrForbidden) {
				http.Error(w, "action not allowed", http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var limitPtr *int64
		limit := s.resolveExportLimit(request.Limit, int64(r.Limit))
		if limit != 0 {
			limitPtr = &limit
		}

		q = &queries.MetricsViewRows{
			MetricsViewName:    r.MetricsViewName,
			TimeStart:          r.TimeStart,
			TimeEnd:            r.TimeEnd,
			Filter:             r.Filter,
			Where:              r.Where,
			Sort:               r.Sort,
			Limit:              limitPtr,
			TimeZone:           r.TimeZone,
			MetricsView:        mv.ValidSpec,
			ResolvedMVSecurity: security,
			Streaming:          mv.Streaming,
		}
	case *runtimev1.Query_MetricsViewTimeSeriesRequest:
		r := v.MetricsViewTimeSeriesRequest

		q = &queries.MetricsViewTimeSeries{
			MetricsViewName: r.MetricsViewName,
			MeasureNames:    r.MeasureNames,
			TimeStart:       r.TimeStart,
			TimeEnd:         r.TimeEnd,
			Where:           r.Where,
			Having:          r.Having,
			TimeGranularity: r.TimeGranularity,
			TimeZone:        r.TimeZone,
			SecurityClaims:  claims,
		}
	case *runtimev1.Query_MetricsViewComparisonRequest:
		r := v.MetricsViewComparisonRequest
		q = &queries.MetricsViewComparison{
			MetricsViewName:     r.MetricsViewName,
			DimensionName:       r.Dimension.Name,
			Measures:            r.Measures,
			ComparisonMeasures:  r.ComparisonMeasures,
			TimeRange:           r.TimeRange,
			ComparisonTimeRange: r.ComparisonTimeRange,
			Limit:               s.resolveExportLimit(request.Limit, r.Limit),
			Offset:              r.Offset,
			Sort:                r.Sort,
			Filter:              r.Filter,
			Where:               r.Where,
			Having:              r.Having,
			SecurityClaims:      claims,
			ExecutionTime:       execTime,
		}
	case *runtimev1.Query_TableRowsRequest:
		r := v.TableRowsRequest
		if !auth.GetClaims(req.Context(), r.InstanceId).Can(runtime.ReadOLAP) {
			http.Error(w, "action not allowed", http.StatusUnauthorized)
			return
		}

		q = &queries.TableHead{
			Connector:      r.Connector,
			Database:       r.Database,
			DatabaseSchema: r.DatabaseSchema,
			TableName:      r.TableName,
			Limit:          int(r.Limit),
		}
	default:
		http.Error(w, fmt.Sprintf("unsupported request type: %s", reflect.TypeOf(v).Name()), http.StatusBadRequest)
		return
	}

	err = q.Export(req.Context(), s.runtime, request.InstanceId, w, &runtime.ExportOptions{
		Format: request.Format,
		PreWriteHook: func(filename string) error {
			// Add timestamp to filename
			filename += "_" + time.Now().Format("20060102150405")

			// Write HTTP headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			switch request.Format {
			case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
				w.Header().Set("Content-Type", "text/csv")
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", filename))
			case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
				w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.xlsx\"", filename))
			case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.parquet\"", filename))
			default:
				return fmt.Errorf("unsupported format %q", request.Format.String())
			}
			return nil
		},
		IncludeHeader:   request.IncludeHeader,
		OriginDashboard: request.OriginDashboard,
		OriginURL:       request.OriginUrl,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) resolveExportLimit(base, override int64) int64 {
	res := base
	if override > 0 && override < res {
		res = override
	}
	return res
}

// downloadTokenTTL determines how long a download token is valid.
const downloadTokenTTL = 1 * time.Hour

// downloadToken is the non-encrypted representation of a download token.
type downloadToken struct {
	Request   []byte    `json:"req"`
	Claims    string    `json:"claims"`
	ExpiresOn time.Time `json:"exp"`
}

// register downloadToken for gob encoding
func init() {
	gob.Register(downloadToken{})
}

// generateDownloadToken generates and encrypts a download token for the given request and attributes.
func (s *Server) generateDownloadToken(req *runtimev1.ExportRequest, claims *runtime.SecurityClaims) (string, error) {
	r, err := proto.Marshal(req)
	if err != nil {
		return "", err
	}

	r, err = gzipCompress(r)
	if err != nil {
		return "", err
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	tkn := downloadToken{
		Request:   r,
		Claims:    string(claimsJSON),
		ExpiresOn: time.Now().Add(downloadTokenTTL),
	}

	res, err := s.codec.Encode(tkn)
	if err != nil {
		return "", err
	}

	return res, nil
}

// parseDownloadToken decrypts and parses a download token and returns the request and attributes.
func (s *Server) parseDownloadToken(tknStr string) (*runtimev1.ExportRequest, *runtime.SecurityClaims, error) {
	tkn := downloadToken{}
	err := s.codec.Decode(tknStr, &tkn)
	if err != nil {
		return nil, nil, err
	}

	if tkn.ExpiresOn.Before(time.Now()) {
		return nil, nil, fmt.Errorf("download token expired")
	}

	claims := &runtime.SecurityClaims{}
	err = json.Unmarshal([]byte(tkn.Claims), claims)
	if err != nil {
		return nil, nil, err
	}

	r, err := gzipDecompress(tkn.Request)
	if err != nil {
		return nil, nil, err
	}

	req := &runtimev1.ExportRequest{}
	err = proto.Unmarshal(r, req)
	if err != nil {
		return nil, nil, err
	}

	return req, claims, nil
}

// gzipCompress compress the input bytes using gzip.
func gzipCompress(v []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err := w.Write(v)
	if err != nil {
		_ = w.Close()
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// gzipDecompress decompresses the input bytes using gzip.
func gzipDecompress(v []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(v))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}
