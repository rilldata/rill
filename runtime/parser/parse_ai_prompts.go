package parser

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"gopkg.in/yaml.v3"
)

// Limits for the `ai_prompts` list.
// The label is a short button text, so it is kept well below the length of a full prompt.
const (
	maxAIPrompts       = 8
	maxAIPromptLabel   = 40
	maxAIPromptLength  = 500
	aiPromptLabelWords = 6
)

// AIPromptYAML represents an entry in an `ai_prompts` list.
// It accepts either a plain string (the prompt, with the label derived from it) or a mapping with `label` and `prompt`.
// Example:
//
//	ai_prompts:
//	  - Which campaigns drove the biggest change in impressions this week?
//	  - label: CTR outliers
//	    prompt: Which publishers have a click-through rate far above or below the average?
type AIPromptYAML struct {
	Label  string `yaml:"label"`
	Prompt string `yaml:"prompt"`
}

func (y *AIPromptYAML) UnmarshalYAML(v *yaml.Node) error {
	switch v.Kind {
	case yaml.ScalarNode:
		y.Prompt = v.Value
		return nil
	case yaml.MappingNode:
		type plain AIPromptYAML
		var tmp plain
		if err := v.Decode(&tmp); err != nil {
			return err
		}
		*y = AIPromptYAML(tmp)
		return nil
	default:
		return errors.New("an AI prompt must be a string or a mapping with `label` and `prompt`")
	}
}

// parseAIPrompts validates an `ai_prompts` list and converts it to protobufs.
// It derives a label from the prompt when none is provided.
func parseAIPrompts(prompts []AIPromptYAML) ([]*runtimev1.AIPrompt, error) {
	if len(prompts) == 0 {
		return nil, nil
	}
	if len(prompts) > maxAIPrompts {
		return nil, fmt.Errorf("ai_prompts can have at most %d entries", maxAIPrompts)
	}

	res := make([]*runtimev1.AIPrompt, 0, len(prompts))
	for i, p := range prompts {
		prompt := strings.TrimSpace(p.Prompt)
		if prompt == "" {
			return nil, fmt.Errorf("ai_prompts entry %d must have a non-empty prompt", i+1)
		}
		if utf8.RuneCountInString(prompt) > maxAIPromptLength {
			return nil, fmt.Errorf("ai_prompts entry %d exceeds the maximum prompt length of %d characters", i+1, maxAIPromptLength)
		}
		label := strings.TrimSpace(p.Label)
		if label == "" {
			label = deriveAIPromptLabel(prompt)
		} else if utf8.RuneCountInString(label) > maxAIPromptLabel {
			return nil, fmt.Errorf("ai_prompts entry %d exceeds the maximum label length of %d characters", i+1, maxAIPromptLabel)
		}
		res = append(res, &runtimev1.AIPrompt{Label: label, Prompt: prompt})
	}
	return res, nil
}

// deriveAIPromptLabel builds a short label from a prompt by taking its first few words and trimming trailing punctuation.
func deriveAIPromptLabel(prompt string) string {
	words := strings.Fields(prompt)
	if len(words) > aiPromptLabelWords {
		words = words[:aiPromptLabelWords]
	}
	label := strings.Join(words, " ")
	label = strings.TrimRight(label, " ,.;:!?")
	if utf8.RuneCountInString(label) > maxAIPromptLabel {
		runes := []rune(label)
		label = strings.TrimRight(string(runes[:maxAIPromptLabel-1]), " ,.;:!?") + "…"
	}
	if len(words) < len(strings.Fields(prompt)) && !strings.HasSuffix(label, "…") {
		label += "…"
	}
	return label
}
