package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestExportWorkbookHandler(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"m1.sql":    "SELECT 'US' AS country",
			"mv1.yaml": `
type: metrics_view
version: 1
model: m1
dimensions:
- column: country
measures:
- name: count
  expression: COUNT(*)
`,
			"c1.yaml": `
type: canvas
rows:
- items:
  - kpi:
      metrics_view: mv1
      measure: count
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 0, 0)

	srv, err := NewServer(context.Background(), &Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	call := func(claims *runtime.SecurityClaims, body []byte) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body)).WithContext(auth.WithClaims(context.Background(), claims))
		req.SetPathValue("instance_id", instanceID)
		httputil.Handler(srv.exportWorkbookHandler).ServeHTTP(rec, req)
		return rec
	}

	workbookRequest := func(canvas string) []byte {
		rows, err := structpb.NewList([]any{"US", 1})
		require.NoError(t, err)
		data, err := protojson.Marshal(&runtimev1.CanvasWorkbookExportRequest{
			Canvas: canvas,
			Meta:   &runtimev1.WorkbookMeta{Title: "Countries"},
			DashboardSheets: []*runtimev1.WorkbookDashboardSheet{{
				Title: "Countries",
				Blocks: []*runtimev1.WorkbookBlock{{
					Component: "c1--component-0-0",
					Title:     "Count by country",
					Geometry:  &runtimev1.WorkbookGeometry{WidthPx: 400, HeightPx: 300},
					Content: &runtimev1.WorkbookBlock_Chart{Chart: &runtimev1.WorkbookChartBlock{
						DataSheetId:    "d1",
						Type:           runtimev1.WorkbookChartType_WORKBOOK_CHART_TYPE_COLUMN,
						CategoryColumn: "country",
						Series:         []*runtimev1.WorkbookChartSeries{{Column: "count"}},
					}},
				}},
			}},
			DataSheets: []*runtimev1.WorkbookDataSheet{{
				Id:    "d1",
				Title: "Count by country",
				Columns: []*runtimev1.WorkbookColumn{
					{Name: "country", DisplayName: "Country", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_STRING},
					{Name: "count", DisplayName: "Count", Type: runtimev1.WorkbookColumnType_WORKBOOK_COLUMN_TYPE_NUMBER},
				},
				Rows: []*structpb.ListValue{rows},
			}},
		})
		require.NoError(t, err)
		return data
	}

	reader := &runtime.SecurityClaims{Permissions: []runtime.Permission{runtime.ReadMetrics}}

	t.Run("renders a workbook", func(t *testing.T) {
		rec := call(reader, workbookRequest("c1"))
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		require.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", rec.Header().Get("Content-Type"))
		require.Regexp(t, `^attachment; filename="Countries_\d{14}\.xlsx"$`, rec.Header().Get("Content-Disposition"))
		f, err := excelize.OpenReader(bytes.NewReader(rec.Body.Bytes()))
		require.NoError(t, err)
		defer f.Close()
		require.Equal(t, []string{"Countries", "Data - Count by country"}, f.GetSheetList())
		v, err := f.GetCellValue("Data - Count by country", "A2")
		require.NoError(t, err)
		require.Equal(t, "US", v)
	})

	t.Run("requires the read metrics permission", func(t *testing.T) {
		rec := call(&runtime.SecurityClaims{Permissions: []runtime.Permission{runtime.ReadAPI}}, workbookRequest("c1"))
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("rejects an unknown canvas", func(t *testing.T) {
		rec := call(reader, workbookRequest("nope"))
		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("rejects a malformed body", func(t *testing.T) {
		rec := call(reader, []byte(`{"canvas": `))
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("rejects a request that fails validation", func(t *testing.T) {
		rec := call(reader, []byte(`{"canvas": "c1"}`))
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "at least one dashboard sheet")
	})
}
