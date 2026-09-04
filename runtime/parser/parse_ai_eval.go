package parser

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// aiEvalMaxCases bounds the number of cases per eval file.
// Eval runs make LLM calls per case, so this is a cost guardrail as much as a sanity limit.
const aiEvalMaxCases = 100

var aiEvalCaseNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// AIEvalYAML is the raw structure of an AIEval resource defined in YAML (does not include common fields)
type AIEvalYAML struct {
	commonYAML  `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	DisplayName string           `yaml:"display_name"`
	Notes       string           `yaml:"notes"`
	Defaults    struct {
		Agent   string `yaml:"agent"`
		Explore string `yaml:"explore"`
	} `yaml:"defaults"`
	Concurrency uint              `yaml:"concurrency"`
	Timeout     string            `yaml:"timeout"`
	CaseTimeout string            `yaml:"case_timeout"`
	Cases       []*aiEvalCaseYAML `yaml:"cases"`
}

type aiEvalCaseYAML struct {
	Name     string `yaml:"name"`
	Question string `yaml:"question"`
	Notes    string `yaml:"notes"`
	Explore  string `yaml:"explore"`
	Expect   struct {
		Answer      string   `yaml:"answer"`
		MetricsView string   `yaml:"metrics_view"`
		Measures    []string `yaml:"measures"`
		Dimensions  []string `yaml:"dimensions"`
	} `yaml:"expect"`
}

// parseAIEval parses an AI eval definition and adds the resulting resource to p.Resources.
func (p *Parser) parseAIEval(node *Node) error {
	// Parse YAML
	tmp := &AIEvalYAML{}
	err := p.decodeNodeYAML(node, true, tmp)
	if err != nil {
		return err
	}

	// Validate SQL or connector isn't set
	if node.SQL != "" {
		return fmt.Errorf("AI evals cannot have SQL")
	}
	if !node.ConnectorInferred && node.Connector != "" {
		return fmt.Errorf("AI evals cannot have a connector")
	}

	// Validate agent
	if tmp.Defaults.Agent == "" {
		tmp.Defaults.Agent = "analyst_agent"
	}
	if tmp.Defaults.Agent != "analyst_agent" {
		return fmt.Errorf(`invalid agent %q: only "analyst_agent" is currently supported`, tmp.Defaults.Agent)
	}

	// Validate concurrency
	if tmp.Concurrency > 8 {
		return fmt.Errorf("concurrency must be at most 8")
	}

	// Parse timeouts
	var timeoutSeconds, caseTimeoutSeconds uint32
	if tmp.Timeout != "" {
		timeout, err := parseDuration(tmp.Timeout)
		if err != nil {
			return fmt.Errorf(`invalid "timeout": %w`, err)
		}
		if timeout < time.Second {
			return fmt.Errorf(`invalid "timeout": must be at least 1s`)
		}
		timeoutSeconds = uint32(timeout.Seconds())
	}
	if tmp.CaseTimeout != "" {
		caseTimeout, err := parseDuration(tmp.CaseTimeout)
		if err != nil {
			return fmt.Errorf(`invalid "case_timeout": %w`, err)
		}
		if caseTimeout < time.Second {
			return fmt.Errorf(`invalid "case_timeout": must be at least 1s`)
		}
		caseTimeoutSeconds = uint32(caseTimeout.Seconds())
	}

	// Validate cases
	if len(tmp.Cases) == 0 {
		return fmt.Errorf(`missing required property "cases"`)
	}
	if len(tmp.Cases) > aiEvalMaxCases {
		return fmt.Errorf("too many cases: at most %d cases are allowed per eval file", aiEvalMaxCases)
	}
	names := make(map[string]bool, len(tmp.Cases))
	var refs []ResourceName
	// Ref the explores used as context, so typos surface at reconcile time instead of failing every case at run time.
	if tmp.Defaults.Explore != "" {
		refs = append(refs, ResourceName{Kind: ResourceKindExplore, Name: tmp.Defaults.Explore})
	}
	for i, c := range tmp.Cases {
		if c == nil {
			return fmt.Errorf("case %d is empty", i)
		}
		if c.Name == "" {
			return fmt.Errorf(`case %d: missing required property "name"`, i)
		}
		if !aiEvalCaseNameRegexp.MatchString(c.Name) {
			return fmt.Errorf("case %q: name may only contain letters, numbers, underscores and hyphens", c.Name)
		}
		if names[strings.ToLower(c.Name)] {
			return fmt.Errorf("duplicate case name %q", c.Name)
		}
		names[strings.ToLower(c.Name)] = true
		if c.Question == "" {
			return fmt.Errorf(`case %q: missing required property "question"`, c.Name)
		}
		if c.Expect.Answer == "" && c.Expect.MetricsView == "" && len(c.Expect.Measures) == 0 && len(c.Expect.Dimensions) == 0 {
			return fmt.Errorf(`case %q: "expect" must set at least one of "answer", "metrics_view", "measures" or "dimensions"`, c.Name)
		}
		if c.Expect.MetricsView != "" {
			refs = append(refs, ResourceName{Kind: ResourceKindMetricsView, Name: c.Expect.MetricsView})
		}
		if c.Explore != "" {
			refs = append(refs, ResourceName{Kind: ResourceKindExplore, Name: c.Explore})
		}
	}

	// Track the eval
	r, err := p.insertResource(ResourceKindAIEval, node.Name, node.Paths, node.Tags, append(node.Refs, refs...)...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.AIEvalSpec.DisplayName = tmp.DisplayName
	if r.AIEvalSpec.DisplayName == "" {
		r.AIEvalSpec.DisplayName = ToDisplayName(node.Name)
	}
	r.AIEvalSpec.Agent = tmp.Defaults.Agent
	r.AIEvalSpec.Notes = tmp.Notes
	r.AIEvalSpec.Explore = tmp.Defaults.Explore
	r.AIEvalSpec.TimeoutSeconds = timeoutSeconds
	r.AIEvalSpec.CaseTimeoutSeconds = caseTimeoutSeconds
	r.AIEvalSpec.Concurrency = uint32(tmp.Concurrency)
	for _, c := range tmp.Cases {
		r.AIEvalSpec.Cases = append(r.AIEvalSpec.Cases, &runtimev1.AIEvalCase{
			Name:              c.Name,
			Question:          c.Question,
			Notes:             c.Notes,
			Explore:           c.Explore,
			ExpectAnswer:      c.Expect.Answer,
			ExpectMetricsView: c.Expect.MetricsView,
			ExpectMeasures:    c.Expect.Measures,
			ExpectDimensions:  c.Expect.Dimensions,
		})
	}

	return nil
}
