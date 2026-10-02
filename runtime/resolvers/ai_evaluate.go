package resolvers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// aiEvaluateSingleQuestionLabel is the label used for the answer when the single `question` property is used.
const aiEvaluateSingleQuestionLabel = "answer"

// aiEvaluateSchema is the schema of the result returned by the ai_evaluate resolver.
// The answer column's struct type is left unset because its fields depend on the answer type (noul, choice or score),
// so each row only contains the fields relevant to its answer type.
// This may cause issues for consumers that rely on a fixed schema for struct columns (e.g. exports or jsonval.ToValue).
var aiEvaluateSchema = &runtimev1.StructType{
	Fields: []*runtimev1.StructType_Field{
		{Name: "label", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
		{Name: "type", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
		{Name: "answer", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT}},
	},
}

func init() {
	runtime.RegisterResolverInitializer("ai_evaluate", newAIEvaluate)
}

type aiEvaluateProps struct {
	Connector string                    `mapstructure:"connector"`
	Question  map[string]any            `mapstructure:"question"`
	Questions map[string]map[string]any `mapstructure:"questions"`
}

type aiEvaluateArgs struct {
	State any `mapstructure:"state"`
}

func newAIEvaluate(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	// Parse props
	props := &aiEvaluateProps{}
	if err := mapstructureutil.WeakDecode(opts.Properties, props); err != nil {
		return nil, err
	}

	// Parse args
	args := &aiEvaluateArgs{}
	if err := mapstructureutil.WeakDecode(opts.Args, args); err != nil {
		return nil, err
	}

	// Normalize the single question into the questions map
	rawQuestions := props.Questions
	if props.Question != nil {
		if props.Questions != nil {
			return nil, errors.New("only one of 'question' or 'questions' can be set")
		}
		rawQuestions = map[string]map[string]any{aiEvaluateSingleQuestionLabel: props.Question}
	}
	if len(rawQuestions) == 0 {
		return nil, errors.New("one of 'question' or 'questions' must be set")
	}

	// Convert the questions to protos.
	// We go through JSON since mapstructure doesn't support the oneof fields in the protos.
	questions := make(map[string]*aiv1.EvaluateQuestion, len(rawQuestions))
	for label, raw := range rawQuestions {
		data, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal question %q: %w", label, err)
		}
		q := &aiv1.EvaluateQuestion{}
		if err := protojson.Unmarshal(data, q); err != nil {
			return nil, fmt.Errorf("invalid question %q: %w", label, err)
		}
		questions[label] = q
	}

	state, err := structpb.NewValue(args.State)
	if err != nil {
		return nil, fmt.Errorf("invalid state: %w", err)
	}

	return &aiEvaluate{
		runtime:    opts.Runtime,
		instanceID: opts.InstanceID,
		connector:  props.Connector,
		questions:  questions,
		state:      state,
	}, nil
}

type aiEvaluate struct {
	runtime    *runtime.Runtime
	instanceID string
	connector  string
	questions  map[string]*aiv1.EvaluateQuestion
	state      *structpb.Value
}

// Close implements runtime.Resolver.
func (t *aiEvaluate) Close() error {
	return nil
}

// CacheKey implements runtime.Resolver.
func (t *aiEvaluate) CacheKey(ctx context.Context) ([]byte, bool, error) {
	// TODO: should be based on question and state?
	return nil, false, nil
}

// Refs implements runtime.Resolver.
func (t *aiEvaluate) Refs() []*runtimev1.ResourceName {
	// TODO: should reference the dimension/measure that possibly defined it?
	return nil
}

// Validate implements runtime.Resolver.
func (t *aiEvaluate) Validate(ctx context.Context) error {
	return nil
}

// ResolveInteractive implements runtime.Resolver.
func (t *aiEvaluate) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	connector := t.connector
	if connector == "" {
		inst, err := t.runtime.Instance(ctx, t.instanceID)
		if err != nil {
			return nil, err
		}
		connector = inst.ResolveAIConnector()
	}

	conn, release, err := t.runtime.AcquireHandle(ctx, t.instanceID, connector)
	if err != nil {
		return nil, err
	}
	defer release()

	ai, ok := conn.AsAI(t.instanceID)
	if !ok {
		return nil, fmt.Errorf("connector %q is not a valid AI service", connector)
	}

	resp, err := ai.Evaluate(ctx, &aiv1.EvaluateRequest{
		State:     t.state,
		Questions: t.questions,
	})
	if err != nil {
		return nil, err
	}

	// Sort labels for deterministic output
	labels := make([]string, 0, len(t.questions))
	for label := range t.questions {
		labels = append(labels, label)
	}
	sort.Strings(labels)

	rows := make([]map[string]any, 0, len(labels))
	for _, label := range labels {
		ans, ok := resp.Answers[label]
		if !ok {
			return nil, fmt.Errorf("AI service did not return an answer for %q", label)
		}

		var typ string
		var answer map[string]any
		switch a := ans.Answer.(type) {
		case *aiv1.EvaluateAnswer_Noul:
			typ = "noul"
			answer = map[string]any{
				"noul": a.Noul.Noul,
			}
		case *aiv1.EvaluateAnswer_Choice:
			typ = "choice"
			answer = map[string]any{
				"choice":        a.Choice.Choice,
				"probabilities": a.Choice.Probabilities,
				"confidence":    a.Choice.Confidence,
			}
		case *aiv1.EvaluateAnswer_Score:
			typ = "score"
			answer = map[string]any{
				"score":         a.Score.Score,
				"legend":        a.Score.Legend,
				"probabilities": a.Score.Probabilities,
				"confidence":    a.Score.Confidence,
			}
		default:
			return nil, fmt.Errorf("unexpected answer type %T for %q", ans.Answer, label)
		}

		rows = append(rows, map[string]any{
			"label":  label,
			"type":   typ,
			"answer": answer,
		})
	}

	return runtime.NewMapsResolverResult(rows, aiEvaluateSchema), nil
}

// ResolveExport implements runtime.Resolver.
func (t *aiEvaluate) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("ai_evaluate resolver does not support export")
}

// InferRequiredSecurityRules implements runtime.Resolver.
func (t *aiEvaluate) InferRequiredSecurityRules() ([]*runtimev1.SecurityRule, error) {
	return nil, nil
}
