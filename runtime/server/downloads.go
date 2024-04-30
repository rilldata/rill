package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func BakeQuery(qry *runtimev1.Query) (string, error) {
	if qry == nil {
		return "", errors.New("cannot bake nil query")
	}

	data, err := proto.Marshal(qry)
	if err != nil {
		return "", err
	}

	data, err = gzipCompress(data)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(data), nil
}

func UnbakeQuery(bakedQry string) (*runtimev1.Query, error) {
	data, err := base64.URLEncoding.DecodeString(bakedQry)
	if err != nil {
		return nil, err
	}

	uncompressed, err := gzipDecompress(data)
	if err != nil {
		// NOTE (2023-11-29): Backwards compatibility for when we didn't gzip baked queries. We can remove this in a few months.
		uncompressed = data
	}

	qry := &runtimev1.Query{}
	if err := proto.Unmarshal(uncompressed, qry); err != nil {
		return nil, err
	}

	return qry, nil
}

func (s *Server) Export(ctx context.Context, req *runtimev1.ExportRequest) (*runtimev1.ExportResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	cfg, err := s.runtime.InstanceConfig(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	if cfg.DownloadRowLimit != 0 {
		if req.Limit > cfg.DownloadRowLimit {
			return nil, status.Errorf(codes.InvalidArgument, "limit must be less than or equal to %d", cfg.DownloadRowLimit)
		}
	}

	if req.BakedQuery != "" {
		qry, err := UnbakeQuery(req.BakedQuery)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to parse baked query: %s", err.Error())
		}

		req.Query = qry
		req.BakedQuery = ""
	}

	tkn, err := s.generateDownloadToken(req, auth.GetClaims(ctx).Attributes())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate download token: %s", err.Error())
	}

	out := fmt.Sprintf("/v1/download?token=%s", tkn)

	return &runtimev1.ExportResponse{
		DownloadUrlPath: out,
	}, nil
}

func (s *Server) downloadHandler(w http.ResponseWriter, req *http.Request) {
	rawTkn := req.URL.Query().Get("token")
	request, attrs, err := s.parseDownloadToken(rawTkn)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse download token: %s", err.Error()), http.StatusBadRequest)
		return
	}

	cfg, err := s.runtime.InstanceConfig(req.Context(), request.InstanceId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var q runtime.Query
	switch v := request.Query.Query.(type) {
	case *runtimev1.Query_MetricsViewAggregationRequest:
		r := v.MetricsViewAggregationRequest

		var limitPtr *int64
		limit := s.resolveExportLimit(cfg, request.Limit, r.Limit)
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
			SecurityAttributes:  attrs,
		}
	case *runtimev1.Query_MetricsViewToplistRequest:
		r := v.MetricsViewToplistRequest

		mv, security, err := resolveMVAndSecurityFromAttributes(req.Context(), s.runtime, request.InstanceId, r.MetricsViewName, attrs)
		if err != nil {
			if errors.Is(err, ErrForbidden) {
				http.Error(w, "action not allowed", http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !checkFieldAccess(r.DimensionName, security) {
			http.Error(w, "action not allowed", http.StatusUnauthorized)
			return
		}

		// validate measures access
		for _, m := range r.MeasureNames {
			if !checkFieldAccess(m, security) {
				http.Error(w, "action not allowed", http.StatusUnauthorized)
				return
			}
		}

		err = validateInlineMeasures(r.InlineMeasures)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var limitPtr *int64
		limit := s.resolveExportLimit(cfg, request.Limit, r.Limit)
		if limit != 0 {
			limitPtr = &limit
		}

		q = &queries.MetricsViewToplist{
			MetricsViewName:    r.MetricsViewName,
			DimensionName:      r.DimensionName,
			MeasureNames:       r.MeasureNames,
			InlineMeasures:     r.InlineMeasures,
			TimeStart:          r.TimeStart,
			TimeEnd:            r.TimeEnd,
			Sort:               r.Sort,
			Where:              r.Where,
			Having:             r.Having,
			Limit:              limitPtr,
			MetricsView:        mv,
			ResolvedMVSecurity: security,
		}
	case *runtimev1.Query_MetricsViewRowsRequest:
		r := v.MetricsViewRowsRequest
		mv, security, err := resolveMVAndSecurityFromAttributes(req.Context(), s.runtime, request.InstanceId, r.MetricsViewName, attrs)
		if err != nil {
			if errors.Is(err, ErrForbidden) {
				http.Error(w, "action not allowed", http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var limitPtr *int64
		limit := s.resolveExportLimit(cfg, request.Limit, int64(r.Limit))
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
			MetricsView:        mv,
			ResolvedMVSecurity: security,
		}
	case *runtimev1.Query_MetricsViewTimeSeriesRequest:
		r := v.MetricsViewTimeSeriesRequest

		mv, security, err := resolveMVAndSecurityFromAttributes(req.Context(), s.runtime, request.InstanceId, r.MetricsViewName, attrs)
		if err != nil {
			if errors.Is(err, ErrForbidden) {
				http.Error(w, "action not allowed", http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = validateInlineMeasures(r.InlineMeasures)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		q = &queries.MetricsViewTimeSeries{
			MetricsViewName:    r.MetricsViewName,
			MeasureNames:       r.MeasureNames,
			InlineMeasures:     r.InlineMeasures,
			TimeStart:          r.TimeStart,
			TimeEnd:            r.TimeEnd,
			TimeGranularity:    r.TimeGranularity,
			Where:              r.Where,
			Having:             r.Having,
			TimeZone:           r.TimeZone,
			MetricsView:        mv,
			ResolvedMVSecurity: security,
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
			Limit:               s.resolveExportLimit(cfg, request.Limit, r.Limit),
			Offset:              r.Offset,
			Sort:                r.Sort,
			Filter:              r.Filter,
			Where:               r.Where,
			Having:              r.Having,
			SecurityAttributes:  attrs,
		}
	case *runtimev1.Query_TableRowsRequest:
		r := v.TableRowsRequest
		if !auth.GetClaims(req.Context()).CanInstance(r.InstanceId, auth.ReadOLAP) {
			http.Error(w, "action not allowed", http.StatusUnauthorized)
			return
		}

		q = &queries.TableHead{
			TableName: r.TableName,
			Limit:     int(r.Limit),
			Result:    nil,
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
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) resolveExportLimit(cfg drivers.InstanceConfig, base, override int64) int64 {
	res := base
	if override < res {
		res = override
	}
	if cfg.DownloadRowLimit != 0 {
		if res == 0 || res > cfg.DownloadRowLimit {
			res = cfg.DownloadRowLimit
		}
	}
	return res
}

// downloadTokenTTL determines how long a download token is valid.
const downloadTokenTTL = 1 * time.Hour

// downloadToken is the non-encrypted representation of a download token.
type downloadToken struct {
	Request    []byte         `json:"req"`
	Attributes map[string]any `json:"attrs"`
	ExpiresOn  time.Time      `json:"exp"`
}

// register downloadToken for gob encoding
func init() {
	gob.Register(downloadToken{})
}

// generateDownloadToken generates and encrypts a download token for the given request and attributes.
func (s *Server) generateDownloadToken(req *runtimev1.ExportRequest, attrs map[string]any) (string, error) {
	r, err := proto.Marshal(req)
	if err != nil {
		return "", err
	}

	r, err = gzipCompress(r)
	if err != nil {
		return "", err
	}

	tkn := downloadToken{
		Request:    r,
		Attributes: attrs,
		ExpiresOn:  time.Now().Add(downloadTokenTTL),
	}

	res, err := s.codec.Encode(tkn)
	if err != nil {
		return "", err
	}

	return res, nil
}

// parseDownloadToken decrypts and parses a download token and returns the request and attributes.
func (s *Server) parseDownloadToken(tknStr string) (*runtimev1.ExportRequest, map[string]any, error) {
	tkn := downloadToken{}
	err := s.codec.Decode(tknStr, &tkn)
	if err != nil {
		return nil, nil, err
	}

	if tkn.ExpiresOn.Before(time.Now()) {
		return nil, nil, fmt.Errorf("download token expired")
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

	return req, tkn.Attributes, nil
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
