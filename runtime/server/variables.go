package server

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/server/auth"
)

func (s *Server) AnalyzeVariables(ctx context.Context, req *runtimev1.AnalyzeVariablesRequest) (*runtimev1.AnalyzeVariablesResponse, error) {
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadInstance) {
		return nil, ErrForbidden
	}

	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	va := &variableAnalyzer{
		analyzedVars: make(map[string]*analyzedVariable),
		inst:         inst,
	}

	// Analyze all sources
	resources, err := ctrl.List(ctx, runtime.ResourceKindSource, "", false)
	if err != nil {
		return nil, err
	}

	for _, r := range resources {
		vars := make(map[string]string)
		err := parser.AnalyzeTemplateRecursively(r.GetSource().Spec.Properties.AsMap(), vars)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze source %q: %w", r.Meta.Name.Name, err)
		}
		va.trackVariablesForResource(vars, parser.ResourceName{Kind: parser.ResourceKindSource, Name: r.Meta.Name.Name})
	}

	// Analyze all models
	resources, err = ctrl.List(ctx, runtime.ResourceKindModel, "", false)
	if err != nil {
		return nil, err
	}

	for _, r := range resources {
		vars := make(map[string]string)
		err := parser.AnalyzeTemplateRecursively(r.GetModel().Spec.InputProperties.AsMap(), vars)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze model input properties %q: %w", r.Meta.Name.Name, err)
		}

		err = parser.AnalyzeTemplateRecursively(r.GetModel().Spec.OutputProperties.AsMap(), vars)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze model output properties %q: %w", r.Meta.Name.Name, err)
		}

		va.trackVariablesForResource(vars, parser.ResourceName{Kind: parser.ResourceKindModel, Name: r.Meta.Name.Name})
	}

	// Analyze all connectors
	resources, err = ctrl.List(ctx, runtime.ResourceKindConnector, "", false)
	if err != nil {
		return nil, err
	}

	for _, r := range resources {
		vars := make(map[string]string)
		err := parser.AnalyzeTemplateRecursively(r.GetConnector().Spec.Properties, vars)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze connector %q: %w", r.Meta.Name.Name, err)
		}
		va.trackVariablesForResource(vars, parser.ResourceName{Kind: parser.ResourceKindConnector, Name: r.Meta.Name.Name})
	}

	// Result
	var analyzedVars []*runtimev1.AnalyzedVariable
	for _, analyzedVar := range va.analyzedVars {
		av := &runtimev1.AnalyzedVariable{
			Name:         analyzedVar.Name,
			DefaultValue: analyzedVar.DefaultValue,
			UsedBy:       make([]*runtimev1.ResourceName, 0, len(analyzedVar.UsedBy)),
		}
		for r := range analyzedVar.UsedBy {
			av.UsedBy = append(av.UsedBy, runtime.ResourceNameFromParser(r))
		}
		analyzedVars = append(analyzedVars, av)
	}

	return &runtimev1.AnalyzeVariablesResponse{
		Variables: analyzedVars,
	}, nil
}

type variableAnalyzer struct {
	analyzedVars map[string]*analyzedVariable
	inst         *drivers.Instance
}

type analyzedVariable struct {
	Name         string
	DefaultValue string
	UsedBy       map[parser.ResourceName]any
}

func (va *variableAnalyzer) trackVariablesForResource(variables map[string]string, r parser.ResourceName) {
	for variable := range variables {
		variable = strings.TrimPrefix(variable, "vars.")
		analyzedVar, ok := va.analyzedVars[variable]
		if ok {
			// Variable is also used by another resource
			analyzedVar.UsedBy[r] = nil
			continue
		}

		analyzedVar = &analyzedVariable{
			Name:   variable,
			UsedBy: map[parser.ResourceName]any{r: nil},
		}
		if def, ok := va.inst.ProjectVariables[variable]; ok {
			analyzedVar.DefaultValue = def
		}
		va.analyzedVars[variable] = analyzedVar
	}
}
