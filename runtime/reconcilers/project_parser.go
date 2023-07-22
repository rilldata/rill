package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	compilerv1 "github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/proto"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindProjectParser, newProjectParser)
}

type ProjectParserReconciler struct {
	C *runtime.Controller
}

func newProjectParser(c *runtime.Controller) runtime.Reconciler {
	return &ProjectParserReconciler{C: c}
}

func (r *ProjectParserReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *ProjectParserReconciler) Reconcile(ctx context.Context, s *runtime.Signal) error {
	// Get ProjectParser resource
	owner, err := r.C.Get(ctx, s.Name)
	if err != nil {
		return err
	}
	pp := owner.GetProjectParser()

	// If deleted, remove all resources created by owner
	if owner.Meta.Deleted {
		r.C.Lock()
		defer r.C.Unlock()

		resources, err := r.C.List(ctx)
		if err != nil {
			return err
		}

		for _, resource := range resources {
			if equalResourceName(resource.Meta.Owner, owner.Meta.Name) {
				err := r.C.Delete(ctx, resource.Meta.Name)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	// Check pp.Spec.Compiler
	if pp.Spec.Compiler != compilerv1.Version {
		return fmt.Errorf("unsupported compiler %q", pp.Spec.Compiler)
	}

	// Get and sync repo
	repo, err := r.C.Runtime.Repo(ctx, r.C.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to access repo: %w", err)
	}
	err = repo.Sync(ctx, r.C.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to sync repo: %w", err)
	}

	// Update commit sha
	hash, err := repo.CommitHash(ctx, r.C.InstanceID)
	if err != nil {
		// Not worth failing the reconcile for this. On error, it'll just set CurrentCommitSha to "".
		r.C.Logger.Error("failed to get commit hash", slog.String("err", err.Error()))
	}
	if pp.State.CurrentCommitSha != hash {
		pp.State.CurrentCommitSha = hash
		err = r.C.UpdateState(ctx, s.Name, owner) // TODO: Pointer relationship between owner and pp makes this hard to follow
		if err != nil {
			return err
		}
	}

	// Parse the project
	parser, err := compilerv1.Parse(ctx, repo, r.C.InstanceID, pp.Spec.DuckdbConnectors)
	if err != nil {
		return fmt.Errorf("failed to parse repo: %w", err)
	}

	// Do the actual reconciliation of parsed resources and catalog resources
	err = r.reconcileParser(ctx, owner, parser, nil)
	if err != nil {
		return err
	}

	// Exit if not watching
	if !pp.Spec.Watch {
		return nil
	}

	// Start a watcher that incrementally reparses the project.
	// This is a blocking and long-running call, which is supported by the controller.
	// If pp.Spec is changed, the controller will cancel the context and call Reconcile again.
	var reparseErr error
	ctx, cancel := context.WithCancel(ctx)
	err = repo.Watch(ctx, r.C.InstanceID, func(events []drivers.WatchEvent) {
		// Get changed paths that are not directories
		changedPaths := make([]string, 0, len(events))
		for _, e := range events {
			if !e.Dir {
				changedPaths = append(changedPaths, e.Path)
			}
		}

		// If reparsing fails, we cancel repo.Watch and exit.
		// NOTE: Parse errors are not returned here (they're stored in p.Errors). Errors returned from Reparse are mainly file system errors.
		diff, err := parser.Reparse(ctx, changedPaths)
		if err == nil {
			err = r.reconcileParser(ctx, owner, parser, diff)
		}
		if err != nil {
			reparseErr = err
			cancel()
			return
		}
	})
	if reparseErr != nil {
		err = fmt.Errorf("re-parse failed: %w", err)
	} else if err != nil {
		if errors.Is(err, ctx.Err()) {
			// The controller cancelled the context. It means pp.Spec was changed. Will be rescheduled.
			return err
		}
		err = fmt.Errorf("watch failed: %w", err)
	}

	// If the watch failed, we return without rescheduling.
	// TODO: Should we have some kind of retry?
	r.C.Logger.Error("stopped watching for file changes", slog.String("err", err.Error()))
	return err
}

// reconcileParser reconciles a parser's output with the current resources in the catalog.
func (r *ProjectParserReconciler) reconcileParser(ctx context.Context, owner *runtimev1.Resource, parser *compilerv1.Parser, diff *compilerv1.Diff) error {
	// Update state from rill.yaml
	if diff == nil || diff.ModifiedRillYAML {
		err := r.reconcileRillYAML(ctx, parser)
		if err != nil {
			return err
		}
	}

	// Update parse errors
	pp := owner.GetProjectParser()
	pp.State.ParseErrors = parser.Errors
	err := r.C.UpdateState(ctx, owner.Meta.Name, owner)
	if err != nil {
		return err
	}

	// Reconcile resources.
	// The lock serves to delay the controller from triggering reconciliation until all resources have been updated.
	// By delaying reconciliation until all resources have been updated, we don't need to worry about making changes in DAG order here.
	r.C.Lock()
	defer r.C.Unlock()
	if diff != nil {
		return r.reconcileResourcesDiff(ctx, owner, parser, diff)
	}
	return r.reconcileResources(ctx, owner, parser)
}

// reconcileRillYAML updates instance config derived from rill.yaml
func (r *ProjectParserReconciler) reconcileRillYAML(ctx context.Context, parser *compilerv1.Parser) error {
	inst, err := r.C.Runtime.FindInstance(ctx, r.C.InstanceID)
	if err != nil {
		return err
	}

	vars := make(map[string]string)
	for _, v := range parser.RillYAML.Variables {
		vars[v.Name] = v.Default
	}

	inst.ProjectVariables = vars
	err = r.C.Runtime.EditInstance(ctx, inst)
	if err != nil {
		return err
	}

	return nil
}

// reconcileResources creates, updates and deletes resources as necessary to match the parser's output with the current resources in the catalog.
func (r *ProjectParserReconciler) reconcileResources(ctx context.Context, owner *runtimev1.Resource, parser *compilerv1.Parser) error {
	// Pass over all existing resources in the catalog.
	resources, err := r.C.List(ctx)
	if err != nil {
		return err
	}
	seen := make(map[compilerv1.ResourceName]bool, len(resources))
	for _, rr := range resources {
		n := resourceNameToCompiler(rr.Meta.Name).Normalized()
		def, ok := parser.Resources[n]

		// If the existing resource is in the parser output, update it.
		if ok {
			seen[n] = true
			err = r.putParserResourceDef(ctx, owner, def, rr)
			if err != nil {
				return err
			}
			continue
		}

		// If the existing resource is not in the parser output, delete it, but only if it was previously created by owner.
		if equalResourceName(rr.Meta.Owner, owner.Meta.Name) {
			err = r.C.Delete(ctx, rr.Meta.Name)
			if err != nil {
				return err
			}
		}
	}

	// Insert resources for the parser outputs that were not seen when passing over the existing resources
	for _, def := range parser.Resources {
		if seen[def.Name.Normalized()] {
			continue
		}

		err = r.putParserResourceDef(ctx, owner, def, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// reconcileResourcesDiff is similar to reconcileResources, but uses a diff from parser.Reparse instead of doing a full comparison of all resources.
func (r *ProjectParserReconciler) reconcileResourcesDiff(ctx context.Context, owner *runtimev1.Resource, parser *compilerv1.Parser, diff *compilerv1.Diff) error {
	// Deletes
	for _, n := range diff.Deleted {
		err := r.C.Delete(ctx, resourceNameFromCompiler(n))
		if err != nil {
			return err
		}
	}

	// Updates
	for _, n := range diff.Modified {
		def := parser.Resources[n]
		existing, err := r.C.Get(ctx, resourceNameFromCompiler(n))
		if err != nil {
			return err
		}
		err = r.putParserResourceDef(ctx, owner, def, existing)
		if err != nil {
			return err
		}
	}

	// Inserts
	for _, n := range diff.Added {
		def := parser.Resources[n]
		err := r.putParserResourceDef(ctx, owner, def, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// putParserResourceDef creates or updates a resource in the catalog based on a parser resource definition.
// It does an insert if existing is nil, otherwise it does an update.
// If existing is not nil, it compares values and only updates meta/spec values if they have changed (ensuring stable resource version numbers).
func (r *ProjectParserReconciler) putParserResourceDef(ctx context.Context, owner *runtimev1.Resource, def *compilerv1.Resource, existing *runtimev1.Resource) error {
	// Make resource spec to insert/update.
	// res should be nil if no spec changes are needed.
	var res *runtimev1.Resource
	switch def.Name.Kind {
	case compilerv1.ResourceKindSource:
		if existing == nil || !equalSourceSpec(existing.GetSource().Spec, def.SourceSpec) {
			res = &runtimev1.Resource{Resource: &runtimev1.Resource_Source{Source: &runtimev1.SourceV2{Spec: def.SourceSpec}}}
		}
	case compilerv1.ResourceKindModel:
		if existing == nil || !equalModelSpec(existing.GetModel().Spec, def.ModelSpec) {
			res = &runtimev1.Resource{Resource: &runtimev1.Resource_Model{Model: &runtimev1.ModelV2{Spec: def.ModelSpec}}}
		}
	case compilerv1.ResourceKindMetricsView:
		if existing == nil || !equalMetricsViewSpec(existing.GetMetricsView().Spec, def.MetricsViewSpec) {
			res = &runtimev1.Resource{Resource: &runtimev1.Resource_MetricsView{MetricsView: &runtimev1.MetricsViewV2{Spec: def.MetricsViewSpec}}}
		}
	case compilerv1.ResourceKindMigration:
		if existing == nil || !equalMigrationSpec(existing.GetMigration().Spec, def.MigrationSpec) {
			res = &runtimev1.Resource{Resource: &runtimev1.Resource_Migration{Migration: &runtimev1.Migration{Spec: def.MigrationSpec}}}
		}
	default:
		panic(fmt.Errorf("unknown resource kind %q", def.Name.Kind))
	}

	// Make refs for the resource meta
	refs := make([]*runtimev1.ResourceName, 0, len(def.Refs))
	for _, r := range def.Refs {
		refs = append(refs, resourceNameFromCompiler(r))
	}

	// Create and return if not updating
	n := resourceNameFromCompiler(def.Name)
	if existing == nil {
		return r.C.Create(ctx, n, refs, owner.Meta.Name, def.Paths, res)
	}

	// Update meta if refs or file paths changed
	if !slices.Equal(existing.Meta.FilePaths, def.Paths) || !slices.Equal(existing.Meta.Refs, refs) {
		err := r.C.UpdateMeta(ctx, n, refs, owner.Meta.Name, def.Paths)
		if err != nil {
			return err
		}
	}

	// Update spec if it changed
	if res != nil {
		err := r.C.UpdateSpec(ctx, n, refs, owner.Meta.Name, def.Paths, res)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceNameFromCompiler(name compilerv1.ResourceName) *runtimev1.ResourceName {
	switch name.Kind {
	case compilerv1.ResourceKindSource:
		return &runtimev1.ResourceName{Kind: runtime.ResourceKindSource, Name: name.Name}
	case compilerv1.ResourceKindModel:
		return &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: name.Name}
	case compilerv1.ResourceKindMetricsView:
		return &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: name.Name}
	case compilerv1.ResourceKindMigration:
		return &runtimev1.ResourceName{Kind: runtime.ResourceKindMigration, Name: name.Name}
	default:
		panic(fmt.Errorf("unknown resource kind %q", name.Kind))
	}
}

func resourceNameToCompiler(name *runtimev1.ResourceName) compilerv1.ResourceName {
	switch name.Kind {
	case runtime.ResourceKindSource:
		return compilerv1.ResourceName{Kind: compilerv1.ResourceKindSource, Name: name.Name}
	case runtime.ResourceKindModel:
		return compilerv1.ResourceName{Kind: compilerv1.ResourceKindModel, Name: name.Name}
	case runtime.ResourceKindMetricsView:
		return compilerv1.ResourceName{Kind: compilerv1.ResourceKindMetricsView, Name: name.Name}
	case runtime.ResourceKindMigration:
		return compilerv1.ResourceName{Kind: compilerv1.ResourceKindMigration, Name: name.Name}
	default:
		panic(fmt.Errorf("unknown resource kind %q", name.Kind))
	}
}

func equalResourceName(a, b *runtimev1.ResourceName) bool {
	return a.Kind == b.Kind && strings.EqualFold(a.Name, b.Name)
}

func equalSourceSpec(a, b *runtimev1.SourceSpec) bool {
	return proto.Equal(a, b)
}

func equalModelSpec(a, b *runtimev1.ModelSpec) bool {
	return proto.Equal(a, b)
}

func equalMetricsViewSpec(a, b *runtimev1.MetricsViewSpec) bool {
	return proto.Equal(a, b)
}

func equalMigrationSpec(a, b *runtimev1.MigrationSpec) bool {
	return proto.Equal(a, b)
}
