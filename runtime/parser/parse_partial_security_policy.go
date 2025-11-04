package parser

import (
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"gopkg.in/yaml.v3"
)

type SecurityPolicyYAML struct {
	Access    string `yaml:"access"`
	RowFilter string `yaml:"row_filter"`
	Include   []*struct {
		Condition string    `yaml:"if"`
		Names     yaml.Node // []string or "*" (will be parsed with parseNamesYAML)
	}
	Exclude []*struct {
		Condition string    `yaml:"if"`
		Names     yaml.Node // []string or "*" (will be parsed with parseNamesYAML)
	}
	Rules []*SecurityRuleYAML `yaml:"rules"`
}

func (p *SecurityPolicyYAML) Proto() ([]*runtimev1.SecurityRule, error) {
	var rules []*runtimev1.SecurityRule
	if p == nil {
		return rules, nil
	}

	if p.Access != "" {
		tmp, err := ResolveTemplate(p.Access, validationTemplateData, false)
		if err != nil {
			return nil, fmt.Errorf(`invalid 'security': 'access' templating is not valid: %w`, err)
		}
		_, err = EvaluateBoolExpression(tmp)
		if err != nil {
			return nil, fmt.Errorf(`invalid 'security': 'access' expression error: %w`, err)
		}

		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_Access{
				Access: &runtimev1.SecurityRuleAccess{
					ConditionExpression: p.Access,
					Allow:               true,
				},
			},
		})
	} else {
		// If "security:" is present, but "access:" is not, default to deny all
		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_Access{
				Access: &runtimev1.SecurityRuleAccess{
					Allow: false,
				},
			},
		})
	}

	if p.RowFilter != "" {
		_, err := ResolveTemplate(p.RowFilter, validationTemplateData, false)
		if err != nil {
			return nil, fmt.Errorf(`invalid 'security': 'row_filter' templating is not valid: %w`, err)
		}

		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_RowFilter{
				RowFilter: &runtimev1.SecurityRuleRowFilter{
					Sql: p.RowFilter,
				},
			},
		})
	}

	for _, inc := range p.Include {
		if inc == nil {
			continue
		}

		tmp, err := ResolveTemplate(inc.Condition, validationTemplateData, false)
		if err != nil {
			return nil, fmt.Errorf(`invalid 'security': 'if' condition templating is not valid: %w`, err)
		}
		_, err = EvaluateBoolExpression(tmp)
		if err != nil {
			return nil, fmt.Errorf(`invalid 'security': 'if' condition expression error: %w`, err)
		}

		names, all, err := parseNamesYAML(inc.Names)
		if err != nil {
			return nil, fmt.Errorf(`invalid 'security': 'include' names: %w`, err)
		}

		if all && len(names) > 0 {
			return nil, fmt.Errorf(`invalid 'security': 'include' cannot have both 'all: true' and specific 'names' fields`)
		} else if !all && len(names) == 0 {
			return nil, fmt.Errorf(`invalid 'security': 'include' must have 'all: true' or a valid 'names' list`)
		}

		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_FieldAccess{
				FieldAccess: &runtimev1.SecurityRuleFieldAccess{
					ConditionExpression: inc.Condition,
					Allow:               true,
					Fields:              names,
					AllFields:           all,
				},
			},
		})
	}

	if len(p.Include) == 0 && len(p.Exclude) > 0 {
		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_FieldAccess{
				FieldAccess: &runtimev1.SecurityRuleFieldAccess{
					Allow:     true,
					AllFields: true,
				},
			},
		})
	}

	for _, exc := range p.Exclude {
		if exc == nil {
			continue
		}

		tmp, err := ResolveTemplate(exc.Condition, validationTemplateData, false)
		if err != nil {
			return nil, fmt.Errorf(`invalid 'security': 'if' condition templating is not valid: %w`, err)
		}
		_, err = EvaluateBoolExpression(tmp)
		if err != nil {
			return nil, fmt.Errorf(`invalid 'security': 'if' condition expression error: %w`, err)
		}

		names, all, err := parseNamesYAML(exc.Names)
		if err != nil {
			return nil, fmt.Errorf(`invalid 'security': 'exclude' names: %w`, err)
		}

		if all && len(names) > 0 {
			return nil, fmt.Errorf(`invalid 'security': 'exclude' cannot have both 'all: true' and specific 'names' fields`)
		} else if !all && len(names) == 0 {
			return nil, fmt.Errorf(`invalid 'security': 'exclude' must have 'all: true' or a valid 'names' list`)
		}

		rules = append(rules, &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_FieldAccess{
				FieldAccess: &runtimev1.SecurityRuleFieldAccess{
					ConditionExpression: exc.Condition,
					Allow:               false,
					Fields:              names,
					AllFields:           all,
				},
			},
		})
	}

	for _, r := range p.Rules {
		if r == nil {
			continue
		}

		rule, err := r.Proto()
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

type SecurityRuleYAML struct {
	Type   string
	Action string
	If     string
	Names  []string
	All    bool
	SQL    string
}

func (r *SecurityRuleYAML) Proto() (*runtimev1.SecurityRule, error) {
	condition := r.If
	if condition != "" {
		tmp, err := ResolveTemplate(condition, validationTemplateData, false)
		if err != nil {
			return nil, fmt.Errorf(`invalid 'if': templating is not valid: %w`, err)
		}
		_, err = EvaluateBoolExpression(tmp)
		if err != nil {
			return nil, fmt.Errorf(`invalid 'if': expression error: %w`, err)
		}
	}

	var allow *bool
	switch r.Action {
	case "allow":
		tmp := true
		allow = &tmp
	case "deny":
		tmp := false
		allow = &tmp
	default:
		if r.Action != "" {
			return nil, fmt.Errorf("invalid security rule action %q", r.Action)
		}
	}

	switch r.Type {
	case "access":
		if allow == nil {
			return nil, fmt.Errorf("invalid security rule of type %q: must specify an action", r.Type)
		}
		return &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_Access{
				Access: &runtimev1.SecurityRuleAccess{
					ConditionExpression: condition,
					Allow:               *allow,
				},
			},
		}, nil
	case "field_access":
		if allow == nil {
			return nil, fmt.Errorf("invalid security rule of type %q: must specify an action", r.Type)
		}

		if r.All && len(r.Names) > 0 {
			return nil, fmt.Errorf(`invalid security rule of type %q: cannot have both 'all: true' and specific 'names' fields`, r.Type)
		} else if !r.All && len(r.Names) == 0 {
			return nil, fmt.Errorf(`invalid security rule of type %q: must have 'all: true' or a valid 'names' list`, r.Type)
		}

		return &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_FieldAccess{
				FieldAccess: &runtimev1.SecurityRuleFieldAccess{
					ConditionExpression: condition,
					Allow:               *allow,
					Fields:              r.Names,
					AllFields:           r.All,
				},
			},
		}, nil
	case "row_filter":
		if allow != nil {
			return nil, fmt.Errorf("invalid security rule of type %q: cannot specify an action", r.Type)
		}
		if r.SQL == "" {
			return nil, fmt.Errorf("invalid security rule of type %q: must provide a 'sql' property", r.Type)
		}
		return &runtimev1.SecurityRule{
			Rule: &runtimev1.SecurityRule_RowFilter{
				RowFilter: &runtimev1.SecurityRuleRowFilter{
					ConditionExpression: condition,
					Sql:                 r.SQL,
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("invalid security rule type %q", r.Type)
	}
}
