package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/queries"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// evaluateConcurrency is the maximum number of rows evaluated in parallel.
const evaluateConcurrency = 5
const hardLimit = 10

// Evaluation evaluates AI questions for each row of a metrics view aggregation.
type Evaluation struct {
	runtime    *runtime.Runtime
	instanceID string
	req        *aiv1.EvaluateRequest
	query      *runtimev1.MetricsViewAggregationRequest
	claims     *runtime.SecurityClaims
}

type questionDimensionMeta struct {
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
}

type questionMeasureMeta struct {
	DisplayName string  `json:"display_name,omitempty"`
	Description string  `json:"description,omitempty"`
	Max         float64 `json:"max"`
	Avg         float64 `json:"avg"`
	Min         float64 `json:"min"`
}

type questionMeta struct {
	Dimensions map[string]*questionDimensionMeta `json:"dimensions"`
	Measures   map[string]*questionMeasureMeta   `json:"measures"`
}

func NewEvaluation(rt *runtime.Runtime, instanceID string, req *aiv1.EvaluateRequest, query *runtimev1.MetricsViewAggregationRequest, claims *runtime.SecurityClaims) *Evaluation {
	return &Evaluation{
		runtime:    rt,
		instanceID: instanceID,
		req:        req,
		query:      query,
		claims:     claims,
	}
}

// Execute runs the aggregation query and evaluates the questions for each row of the result.
// The answer to each question is appended to the row as a column named by evaluateColumnName.
func (e *Evaluation) Execute(ctx context.Context) (*runtimev1.MetricsViewEvaluateResponse, error) {
	res, err := e.executeQuery(ctx)
	if err != nil {
		return nil, err
	}

	qm, err := e.createQuestionMeta(ctx, res.Data)
	if err != nil {
		return nil, err
	}
	// Templates access maps by key but structs by Go field name, so we convert the meta to a map to expose its JSON names.
	metaJSON, err := json.Marshal(qm)
	if err != nil {
		return nil, err
	}
	var meta map[string]any
	if err := json.Unmarshal(metaJSON, &meta); err != nil {
		return nil, err
	}

	// TODO: Use the instance's configured AI connector once drivers other than typesafe implement Evaluate.
	conn, release, err := e.runtime.AcquireHandle(ctx, e.instanceID, "typesafe")
	if err != nil {
		return nil, err
	}
	defer release()
	aiService, ok := conn.AsAI(e.instanceID)
	if !ok {
		return nil, fmt.Errorf("connector %q is not a valid AI service", "typesafe")
	}

	data := make([]*structpb.Struct, len(res.Data))
	grp, grpCtx := errgroup.WithContext(ctx)
	grp.SetLimit(evaluateConcurrency)
	for i, row := range res.Data {
		if i >= hardLimit {
			break
		}
		grp.Go(func() error {
			evalRow, err := e.executeRow(grpCtx, aiService, row, meta)
			if err != nil {
				return err
			}
			data[i] = evalRow
			return nil
		})
	}
	if err := grp.Wait(); err != nil {
		return nil, err
	}

	// Sort labels for a deterministic schema
	labels := make([]string, 0, len(e.req.Questions))
	for label := range e.req.Questions {
		labels = append(labels, label)
	}
	sort.Strings(labels)

	schema := &runtimev1.StructType{}
	if res.Schema != nil {
		schema.Fields = append(schema.Fields, res.Schema.Fields...)
	}
	for _, label := range labels {
		schema.Fields = append(schema.Fields, &runtimev1.StructType_Field{
			Name: label,
			Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT},
		})
	}

	return &runtimev1.MetricsViewEvaluateResponse{
		Schema: schema,
		Data:   data,
	}, nil
}

func (e *Evaluation) executeRow(ctx context.Context, aiService drivers.AIService, row *structpb.Struct, meta map[string]any) (*structpb.Struct, error) {
	rowJSON, err := json.Marshal(row.AsMap())
	if err != nil {
		return nil, err
	}

	templateData := parser.TemplateData{
		// TODO: env, user & variables (ignore)
		State: map[string]any{
			"data": string(rowJSON),
			"meta": meta,
		},
	}

	// Rows are evaluated in parallel, so we resolve the templates on a copy of the request.
	req := proto.Clone(e.req).(*aiv1.EvaluateRequest)
	req.State = structpb.NewStructValue(row)
	for label, q := range req.Questions {
		var values []*structpb.Value
		switch q := q.Question.(type) {
		case *aiv1.EvaluateQuestion_Noul:
			values = append(values, q.Noul.Instructions)
			if q.Noul.Criteria != nil {
				values = append(values, q.Noul.Criteria.True, q.Noul.Criteria.False)
			}
		case *aiv1.EvaluateQuestion_Choice:
			values = append(values, q.Choice.Instructions)
			for _, v := range q.Choice.Criteria {
				values = append(values, v)
			}
		case *aiv1.EvaluateQuestion_Score:
			values = append(values, q.Score.Instructions)
			values = append(values, q.Score.Criteria...)
		default:
			return nil, fmt.Errorf("unexpected question type %T for %q", q, label)
		}

		for _, v := range values {
			if err := resolveValueTemplates(v, templateData); err != nil {
				return nil, fmt.Errorf("failed to resolve templates for question %q: %w", label, err)
			}
		}
	}

	resp, err := aiService.Evaluate(ctx, req)
	if err != nil {
		return nil, err
	}

	fields := make(map[string]*structpb.Value, len(row.Fields)+len(req.Questions))
	for k, v := range row.Fields {
		fields[k] = v
	}
	for label := range req.Questions {
		ans, ok := resp.Answers[label]
		if !ok {
			return nil, fmt.Errorf("AI service did not return an answer for %q", label)
		}

		var answer proto.Message
		switch a := ans.Answer.(type) {
		case *aiv1.EvaluateAnswer_Noul:
			answer = a.Noul
		case *aiv1.EvaluateAnswer_Choice:
			answer = a.Choice
		case *aiv1.EvaluateAnswer_Score:
			answer = a.Score
		default:
			return nil, fmt.Errorf("unexpected answer type %T for %q", ans.Answer, label)
		}

		// Go through JSON to convert the answer to a plain struct with snake_case keys.
		answerJSON, err := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}.Marshal(answer)
		if err != nil {
			return nil, err
		}
		answerVal := &structpb.Value{}
		if err := protojson.Unmarshal(answerJSON, answerVal); err != nil {
			return nil, err
		}
		fields[label] = answerVal
	}

	return &structpb.Struct{Fields: fields}, nil
}

func (e *Evaluation) executeQuery(ctx context.Context) (*runtimev1.MetricsViewAggregationResponse, error) {
	tr := e.query.TimeRange
	if e.query.TimeStart != nil || e.query.TimeEnd != nil {
		tr = &runtimev1.TimeRange{
			Start: e.query.TimeStart,
			End:   e.query.TimeEnd,
		}
	}

	q := &queries.MetricsViewAggregation{
		MetricsViewName:     e.query.MetricsView,
		Dimensions:          e.query.Dimensions,
		Measures:            e.query.Measures,
		Sort:                e.query.Sort,
		TimeRange:           tr,
		ComparisonTimeRange: e.query.ComparisonTimeRange,
		Where:               e.query.Where,
		WhereSQL:            e.query.WhereSql,
		Having:              e.query.Having,
		HavingSQL:           e.query.HavingSql,
		Filter:              e.query.Filter,
		Limit:               &e.query.Limit,
		Offset:              e.query.Offset,
		PivotOn:             e.query.PivotOn,
		SecurityClaims:      e.claims,
		Exact:               e.query.Exact,
		Aliases:             e.query.Aliases,
		FillMissing:         e.query.FillMissing,
		Rows:                e.query.Rows,
	}

	err := e.runtime.Query(ctx, e.instanceID, q, int(e.query.Priority))
	if err != nil {
		return nil, err
	}
	return q.Result, nil
}

func (e *Evaluation) createQuestionMeta(ctx context.Context, rows []*structpb.Struct) (*questionMeta, error) {
	ctrl, err := e.runtime.Controller(ctx, e.instanceID)
	if err != nil {
		return nil, err
	}
	r, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: e.query.MetricsView}, false)
	if err != nil {
		return nil, err
	}
	mv, access, err := e.runtime.ApplySecurityPolicy(ctx, e.instanceID, e.claims, r)
	if err != nil {
		return nil, err
	}
	if !access {
		return nil, fmt.Errorf("metrics view %q not found", e.query.MetricsView)
	}
	mvSpec := mv.GetMetricsView().State.ValidSpec // TODO: validate (ignore)

	qm := &questionMeta{
		Dimensions: make(map[string]*questionDimensionMeta),
		Measures:   make(map[string]*questionMeasureMeta),
	}

	for _, rd := range e.query.Dimensions {
		for _, md := range mvSpec.Dimensions {
			if rd.Name != md.Name {
				continue
			}

			qm.Dimensions[rd.Name] = &questionDimensionMeta{
				DisplayName: md.DisplayName,
				Description: md.Description,
			}
			break
		}
	}

	for _, rm := range e.query.Measures {
		for _, mm := range mvSpec.Measures {
			if rm.Name != mm.Name {
				continue
			}

			qm.Measures[rm.Name] = &questionMeasureMeta{
				DisplayName: mm.DisplayName,
				Description: mm.Description,
			}
			break
		}
	}

	calculateMeasuresMeta(rows, qm)

	return qm, nil
}

// calculateMeasuresMeta sets the min, avg and max of each measure across the evaluated rows.
// It computes them in Go since measure expressions do not support aggregate functions.
// Rows where a measure is not a number (e.g. null) are skipped for that measure.
func calculateMeasuresMeta(rows []*structpb.Struct, meta *questionMeta) {
	for name, mesMeta := range meta.Measures {
		var sum float64
		var n int
		for _, row := range rows {
			v, ok := row.Fields[name].GetKind().(*structpb.Value_NumberValue)
			if !ok {
				continue
			}
			if n == 0 || v.NumberValue < mesMeta.Min {
				mesMeta.Min = v.NumberValue
			}
			if n == 0 || v.NumberValue > mesMeta.Max {
				mesMeta.Max = v.NumberValue
			}
			sum += v.NumberValue
			n++
		}
		if n > 0 {
			mesMeta.Avg = sum / float64(n)
		}
	}
}

// resolveValueTemplates resolves templates in every string nested in the value, in place.
func resolveValueTemplates(v *structpb.Value, data parser.TemplateData) error {
	switch k := v.GetKind().(type) {
	case *structpb.Value_StringValue:
		s, err := parser.ResolveTemplate(k.StringValue, data, false)
		if err != nil {
			return err
		}
		k.StringValue = s
	case *structpb.Value_ListValue:
		for _, item := range k.ListValue.GetValues() {
			if err := resolveValueTemplates(item, data); err != nil {
				return err
			}
		}
	case *structpb.Value_StructValue:
		for _, field := range k.StructValue.GetFields() {
			if err := resolveValueTemplates(field, data); err != nil {
				return err
			}
		}
	}
	return nil
}
