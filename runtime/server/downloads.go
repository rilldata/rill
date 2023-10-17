package server

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
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

	return base64.URLEncoding.EncodeToString(data), nil
}

func UnbakeQuery(bakedQry string) (*runtimev1.Query, error) {
	data, err := base64.URLEncoding.DecodeString(bakedQry)
	if err != nil {
		return nil, err
	}

	qry := &runtimev1.Query{}
	if err := proto.Unmarshal(data, qry); err != nil {
		return nil, err
	}

	return qry, nil
}

func (s *Server) Export(ctx context.Context, req *runtimev1.ExportRequest) (*runtimev1.ExportResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	if s.opts.DownloadRowLimit != nil {
		if req.Limit > *s.opts.DownloadRowLimit {
			return nil, status.Errorf(codes.InvalidArgument, "limit must be less than or equal to %d", *s.opts.DownloadRowLimit)
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

	var q runtime.Query
	switch v := request.Query.Query.(type) {
	case *runtimev1.Query_MetricsViewAggregationRequest:
		r := v.MetricsViewAggregationRequest
		mv, security, err := resolveMVAndSecurityFromAttributes(req.Context(), s.runtime, request.InstanceId, r.MetricsView, attrs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, dim := range r.Dimensions {
			if dim.Name == mv.TimeDimension {
				// checkFieldAccess doesn't currently check the time dimension
				continue
			}
			if !checkFieldAccess(dim.Name, security) {
				http.Error(w, "action not allowed", http.StatusUnauthorized)
				return
			}
		}

		for _, m := range r.Measures {
			if m.BuiltinMeasure != runtimev1.BuiltinMeasure_BUILTIN_MEASURE_UNSPECIFIED {
				continue
			}
			if !checkFieldAccess(m.Name, security) {
				http.Error(w, "action not allowed", http.StatusUnauthorized)
				return
			}
		}

		var limitPtr *int64
		limit := s.resolveExportLimit(request.Limit, r.Limit)
		if limit != 0 {
			limitPtr = &limit
		}

		q = &queries.MetricsViewAggregation{
			MetricsViewName:    r.MetricsView,
			Dimensions:         r.Dimensions,
			Measures:           r.Measures,
			Sort:               r.Sort,
			TimeStart:          r.TimeStart,
			TimeEnd:            r.TimeEnd,
			Filter:             r.Filter,
			Limit:              limitPtr,
			Offset:             r.Offset,
			MetricsView:        mv,
			ResolvedMVSecurity: security,
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
		limit := s.resolveExportLimit(request.Limit, r.Limit)
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
			Filter:             r.Filter,
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
		limit := s.resolveExportLimit(request.Limit, int64(r.Limit))
		if limit != 0 {
			limitPtr = &limit
		}

		q = &queries.MetricsViewRows{
			MetricsViewName:    r.MetricsViewName,
			TimeStart:          r.TimeStart,
			TimeEnd:            r.TimeEnd,
			Filter:             r.Filter,
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
			Filter:             r.Filter,
			TimeZone:           r.TimeZone,
			MetricsView:        mv,
			ResolvedMVSecurity: security,
		}
	case *runtimev1.Query_MetricsViewComparisonToplistRequest:
		r := v.MetricsViewComparisonToplistRequest

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

		q = &queries.MetricsViewComparisonToplist{
			MetricsViewName:     r.MetricsViewName,
			DimensionName:       r.DimensionName,
			MeasureNames:        r.MeasureNames,
			InlineMeasures:      r.InlineMeasures,
			BaseTimeRange:       r.BaseTimeRange,
			ComparisonTimeRange: r.ComparisonTimeRange,
			Limit:               s.resolveExportLimit(request.Limit, r.Limit),
			Offset:              r.Offset,
			Sort:                r.Sort,
			Filter:              r.Filter,
			MetricsView:         mv,
			ResolvedMVSecurity:  security,
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

func (s *Server) resolveExportLimit(base, override int64) int64 {
	res := base
	if override < res {
		res = override
	}
	if s.opts.DownloadRowLimit != nil {
		if res == 0 || res > *s.opts.DownloadRowLimit {
			res = *s.opts.DownloadRowLimit
		}
	}
	return res
}

// downloadTokenTTL determines how long a download token is valid.
const downloadTokenTTL = 1 * time.Hour

// downloadTokenJSON is the non-encrypted JSON representation of a download token.
type downloadTokenJSON struct {
	Request    []byte         `json:"req"`
	Attributes map[string]any `json:"attrs"`
	ExpiresOn  time.Time      `json:"exp"`
}

// generateDownloadToken generates and encrypts a download token for the given request and attributes.
func (s *Server) generateDownloadToken(req *runtimev1.ExportRequest, attrs map[string]any) (string, error) {
	r, err := proto.Marshal(req)
	if err != nil {
		return "", err
	}

	tknJSON := downloadTokenJSON{
		Request:    r,
		Attributes: attrs,
		ExpiresOn:  time.Now().Add(downloadTokenTTL),
	}

	res, err := s.codec.Encode(tknJSON)
	if err != nil {
		return "", err
	}

	return res, nil
}

// parseDownloadToken decrypts and parses a download token and returns the request and attributes.
func (s *Server) parseDownloadToken(tkn string) (*runtimev1.ExportRequest, map[string]any, error) {
	tknJSON := downloadTokenJSON{}
	err := s.codec.Decode(tkn, &tknJSON)
	if err != nil {
		return nil, nil, err
	}

	if tknJSON.ExpiresOn.Before(time.Now()) {
		return nil, nil, fmt.Errorf("download token expired")
	}

	req := &runtimev1.ExportRequest{}
	err = proto.Unmarshal(tknJSON.Request, req)
	if err != nil {
		return nil, nil, err
	}

	return req, tknJSON.Attributes, nil
}
