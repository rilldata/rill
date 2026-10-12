package parser

import (
	"errors"
	"fmt"
	"strings"

	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

// EvaluateQuestionYAML is the YAML representation of an AI evaluation question.
// The `type` determines the expected shape of `criteria`:
//
//	type: noul    # criteria: mapping with `true` and `false` keys (optional)
//	type: choice  # criteria: mapping of choice to description
//	type: score   # criteria: ordered list of score levels
//
// The `instructions` and criteria values can be any YAML value; strings in them are resolved as templates at evaluation time.
type EvaluateQuestionYAML struct {
	Type         string    `yaml:"type"`
	Instructions any       `yaml:"instructions"`
	Criteria     yaml.Node `yaml:"criteria"`
}

func (e *EvaluateQuestionYAML) Proto() (*aiv1.EvaluateQuestion, error) {
	instructions, err := pbutil.ToValue(e.Instructions, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid instructions: %w", err)
	}

	switch strings.ToLower(e.Type) {
	case "noul":
		q := &aiv1.EvaluateNoulQuestion{Instructions: instructions}
		switch e.Criteria.Kind {
		case 0:
		case yaml.MappingNode:
			// Unquoted `true` and `false` keys are booleans in YAML, but the struct tags still match them on their raw text.
			tmp := &struct {
				True  any `yaml:"true"`
				False any `yaml:"false"`
			}{}
			if err := e.Criteria.Decode(tmp); err != nil {
				return nil, fmt.Errorf("invalid criteria: %w", err)
			}
			q.Criteria = &aiv1.EvaluateNoulQuestion_Criteria{}
			q.Criteria.True, err = pbutil.ToValue(tmp.True, nil)
			if err != nil {
				return nil, fmt.Errorf("invalid criteria: %w", err)
			}
			q.Criteria.False, err = pbutil.ToValue(tmp.False, nil)
			if err != nil {
				return nil, fmt.Errorf("invalid criteria: %w", err)
			}
		default:
			return nil, errors.New(`criteria for a "noul" question must be a mapping with "true" and "false" keys`)
		}
		return &aiv1.EvaluateQuestion{Question: &aiv1.EvaluateQuestion_Noul{Noul: q}}, nil
	case "choice":
		if e.Criteria.Kind != yaml.MappingNode {
			return nil, errors.New(`criteria for a "choice" question must be a mapping of choices to descriptions`)
		}
		var tmp map[string]any
		if err := e.Criteria.Decode(&tmp); err != nil {
			return nil, fmt.Errorf("invalid criteria: %w", err)
		}
		criteria := make(map[string]*structpb.Value, len(tmp))
		for k, v := range tmp {
			criteria[k], err = pbutil.ToValue(v, nil)
			if err != nil {
				return nil, fmt.Errorf("invalid criteria for choice %q: %w", k, err)
			}
		}
		return &aiv1.EvaluateQuestion{Question: &aiv1.EvaluateQuestion_Choice{Choice: &aiv1.EvaluateChoiceQuestion{
			Instructions: instructions,
			Criteria:     criteria,
		}}}, nil
	case "score":
		if e.Criteria.Kind != yaml.SequenceNode {
			return nil, errors.New(`criteria for a "score" question must be a list of score levels`)
		}
		var tmp []any
		if err := e.Criteria.Decode(&tmp); err != nil {
			return nil, fmt.Errorf("invalid criteria: %w", err)
		}
		criteria := make([]*structpb.Value, len(tmp))
		for i, v := range tmp {
			criteria[i], err = pbutil.ToValue(v, nil)
			if err != nil {
				return nil, fmt.Errorf("invalid criteria at index %d: %w", i, err)
			}
		}
		return &aiv1.EvaluateQuestion{Question: &aiv1.EvaluateQuestion_Score{Score: &aiv1.EvaluateScoreQuestion{
			Instructions: instructions,
			Criteria:     criteria,
		}}}, nil
	default:
		return nil, fmt.Errorf(`invalid eval question type %q (allowed values: noul, choice, score)`, e.Type)
	}
}
