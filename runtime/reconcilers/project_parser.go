package reconcilers

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	parserpkg "github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/arrayutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrParserHasParseErrors = errors.New("encountered parse errors")

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindProjectParser, newProjectParser)
}

type ProjectParserReconciler struct {
	C *runtime.Controller
}

func newProjectParser(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	return &ProjectParserReconciler{C: c}, nil
}

func (r *ProjectParserReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *ProjectParserReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetProjectParser()
	b := to.GetProjectParser()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *ProjectParserReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetProjectParser()
	b := to.GetProjectParser()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *ProjectParserReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetProjectParser().State = &runtimev1.ProjectParserState{}
	return nil
}

func (r *ProjectParserReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	// Get ProjectParser resource
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	pp := self.GetProjectParser()
	if pp == nil {
		return runtime.ReconcileResult{Err: errors.New("not a project parser")}
	}

	// Reset watching to false (in case of a crash during a previous watch)
	if pp.State.Watching {
		pp.State.Watching = false
		if err = r.C.UpdateState(ctx, n, self); err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Does not support renames
	if self.Meta.RenamedFrom != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("project parser cannot be renamed")}
	}

	// If deleted, remove all resources created by self
	if self.Meta.DeletedOn != nil {
		r.C.Lock(ctx)
		defer r.C.Unlock(ctx)

		resources, err := r.C.List(ctx, "", "", false)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}

		for _, resource := range resources {
			if equalResourceName(resource.Meta.Owner, self.Meta.Name) {
				err := r.C.Delete(ctx, resource.Meta.Name)
				if err != nil {
					return runtime.ReconcileResult{Err: err}
				}
			}
		}

		return runtime.ReconcileResult{}
	}

	// Get and sync repo
	repo, release, err := r.C.Runtime.Repo(ctx, r.C.InstanceID)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to access repo: %w", err)}
	}
	defer release()
	// Pull the latest changes
	// on rill developer do not pull latest changes, all pulls should be user triggered
	err = repo.Pull(ctx, &drivers.PullOptions{UserTriggered: !r.C.Runtime.AllowHostAccess()})
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to pull repo: %w", err)}
	}

	// Update commit sha and timestamp
	hash, err := repo.CommitHash(ctx)
	if err != nil {
		// Not worth failing the reconcile for this. On error, it'll just set CurrentCommitSha to "".
		r.C.Logger.Error("failed to get commit hash", zap.String("error", err.Error()), observability.ZapCtx(ctx))
	}
	ts, err := repo.CommitTimestamp(ctx)
	if err != nil {
		r.C.Logger.Error("failed to get commit timestamp", zap.String("error", err.Error()), observability.ZapCtx(ctx))
	}
	if pp.State.CurrentCommitSha != hash {
		pp.State.CurrentCommitSha = hash
		pp.State.CurrentCommitOn = nil
		if !ts.IsZero() {
			pp.State.CurrentCommitOn = timestamppb.New(ts)
		}
		err = r.C.UpdateState(ctx, n, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Get instance
	inst, err := r.C.Runtime.Instance(ctx, r.C.InstanceID)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to find instance: %w", err)}
	}
	instCfg, err := inst.Config()
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to get instance config: %w", err)}
	}

	// Parse the project
	// NOTE: Explicitly passing inst.OLAPConnector instead of inst.ResolveOLAPConnector() since the parser expects the base name to use if not overridden in rill.yaml.
	parser, err := parserpkg.Parse(ctx, repo, r.C.InstanceID, inst.Environment, inst.OLAPConnector)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to parse: %w", err)}
	}

	// Do the actual reconciliation of parsed resources and catalog resources
	err = r.reconcileParser(ctx, inst, self, parser, nil, nil)

	// If err is not for parse errors, always return. Otherwise, only return it if we're not watching for changes.
	if err != nil && !errors.Is(err, ErrParserHasParseErrors) {
		return runtime.ReconcileResult{Err: err}
	}
	if !instCfg.WatchRepo {
		return runtime.ReconcileResult{Err: err}
	}

	// Set watching to true and add a defer to ensure it's set to false on exit
	pp.State.Watching = true
	if err = r.C.UpdateState(ctx, n, self); err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	defer func() {
		pp.State.Watching = false
		if err = r.C.UpdateState(ctx, n, self); err != nil {
			r.C.Logger.Error("failed to update watch state", zap.Any("error", err), observability.ZapCtx(ctx))
		}
	}()

	// Start a watcher that incrementally reparses the project.
	// This is a blocking and long-running call, which is supported by the controller.
	// If pp.Spec is changed, the controller will cancel the context and call Reconcile again.
	var reparseErr error
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	err = repo.Watch(ctx, func(events []drivers.WatchEvent) {
		// Get changed paths that are not directories
		changedPaths := make([]string, 0, len(events))
		hasDuplicates := false
		for _, e := range events {
			if e.Dir {
				continue
			}
			if parser.IsSkippable(e.Path) {
				// We do not get events for files in deleted/renamed directories.
				// So we need to manually find paths we're tracking in the directory and add them to changedPaths.
				//
				// Note that e.Dir is always false for deletes, so we don't actually know if the path was a directory.
				// Calling TrackedPathsInDir is safe even if the given path isn't a directory.
				//
				// NOTE: This is nested under IsSkippable as an optimization because IsSkippable is true for directories.
				// This is pretty hacky and should be refactored (probably more fundamentally in the watcher itself).
				if e.Type == runtimev1.FileEvent_FILE_EVENT_DELETE {
					ps := parser.TrackedPathsInDir(e.Path)
					if len(ps) > 0 {
						changedPaths = append(changedPaths, ps...)
						hasDuplicates = true
					}
					continue
				}

				continue
			}
			changedPaths = append(changedPaths, e.Path)
		}

		// Small optimization to avoid deduplicating if we know we didn't append to it.
		if hasDuplicates {
			changedPaths = arrayutil.Dedupe(changedPaths)
		}

		if len(changedPaths) == 0 {
			return
		}

		// If reparsing fails, we cancel repo.Watch and exit.
		// NOTE: Parse errors are not returned here (they're stored in p.Errors). Errors returned from Reparse are mainly file system errors.
		diff, err := parser.Reparse(ctx, changedPaths)
		if err == nil {
			err = r.reconcileParser(ctx, inst, self, parser, diff, changedPaths)
		}
		if err != nil && !errors.Is(err, ErrParserHasParseErrors) {
			if reparseErr == nil { // In case a callback is somehow invoked after cancel() is called in a previous callback
				reparseErr = err
				cancel()
			}
			return
		}
	})
	if reparseErr != nil {
		err = fmt.Errorf("re-parse failed: %w", reparseErr)
	} else if err != nil && !errors.Is(err, ctx.Err()) {
		err = fmt.Errorf("watch failed: %w", err)
	}

	// If the watch failed, we return and ask the controller to retry immediately.
	if !errors.Is(err, ctx.Err()) { // context cancellations are used for manual triggers and graceful shutdowns, so not an error.
		r.C.Logger.Error("Stopped watching for file changes, retrying...", zap.String("error", err.Error()), observability.ZapCtx(ctx))
	} else {
		r.C.Logger.Debug("Stopped watching for file changes, retrying...", observability.ZapCtx(ctx))
	}
	return runtime.ReconcileResult{Err: err, Retrigger: time.Now()}
}

func (r *ProjectParserReconciler) ResolveTransitiveAccess(ctx context.Context, claims *runtime.SecurityClaims, res *runtimev1.Resource) ([]*runtimev1.SecurityRule, error) {
	if res.GetProjectParser() == nil {
		return nil, fmt.Errorf("not a project parser resource")
	}
	return []*runtimev1.SecurityRule{{Rule: runtime.SelfAllowRuleAccess(res)}}, nil
}

// reconcileParser reconciles a parser's output with the current resources in the catalog.
func (r *ProjectParserReconciler) reconcileParser(ctx context.Context, inst *drivers.Instance, self *runtimev1.Resource, parser *parserpkg.Parser, diff *parserpkg.Diff, changedPaths []string) error {
	// Update parse errors
	pp := self.GetProjectParser()
	pp.State.ParseErrors = parser.Errors
	err := r.C.UpdateState(ctx, self.Meta.Name, self)
	if err != nil {
		return err
	}

	// Log parse errors
	if diff == nil {
		// This handles a very specific case - when opening the application on an uninitialized directory, we do not want to print an error for "rill.yaml not found".
		// But if the user subsequently in the session, after initializing the project, breaks rill.yaml, then we DO want to log the error.
		// So we rely on StateVersion == 1 on the first call to the reconciler.
		// (The UpdateState calls above do not mutate `self`, which is a cloned object, so the starting StateVersion is preserved here. Also quite hacky.)
		skipRillYAMLErr := inst.IgnoreInitialInvalidProjectError && self.Meta.StateVersion == 1

		for _, e := range parser.Errors {
			if skipRillYAMLErr && e.FilePath == "/rill.yaml" {
				continue
			}
			r.C.Logger.Warn("Parser error", zap.String("path", e.FilePath), zap.String("error", e.Message), observability.ZapCtx(ctx))
		}
	} else if diff.Skipped {
		r.C.Logger.Warn("Not parsing changed paths due to missing or broken rill.yaml", observability.ZapCtx(ctx))
	} else {
		for _, e := range parser.Errors {
			if slices.Contains(changedPaths, e.FilePath) {
				r.C.Logger.Warn("Parser error", zap.String("path", e.FilePath), zap.String("error", e.Message), observability.ZapCtx(ctx))
			}
		}
	}

	// Set an error without returning to mark if there are parse errors (if not, force error to nil in case there previously were parse errors)
	var parseErrsErr error
	if len(parser.Errors) > 0 {
		parseErrsErr = ErrParserHasParseErrors
	}
	err = r.C.UpdateError(ctx, self.Meta.Name, parseErrsErr)
	if err != nil {
		return err
	}

	// If RillYAML is missing, don't reconcile anything
	if parser.RillYAML == nil {
		return parseErrsErr
	}

	// not setting restartController=true when diff is actually nil prevents infinite restarts
	updateConfig := diff == nil || diff.ModifiedDotEnv || diff.Reloaded
	if updateConfig {
		restartController := diff != nil
		err := r.reconcileProjectConfig(ctx, parser, restartController)
		if err != nil {
			return err
		}
		if restartController {
			return nil
		}
	}

	// Reconcile resources.
	// The lock serves to delay the controller from triggering reconciliation until all resources have been updated.
	// By delaying reconciliation until all resources have been updated, we don't need to worry about making changes in DAG order here.
	r.C.Lock(ctx)
	defer r.C.Unlock(ctx)
	if diff != nil {
		err = r.reconcileResourcesDiff(ctx, inst, self, parser, diff)
		if err != nil {
			return err
		}
		return parseErrsErr // Keep the parseErrsErr in this case
	}

	err = r.reconcileResources(ctx, inst, self, parser)
	if err != nil {
		return err
	}
	return parseErrsErr // Keep the parseErrsErr in this case
}

// reconcileProjectConfig updates instance config derived from rill.yaml and .env
func (r *ProjectParserReconciler) reconcileProjectConfig(ctx context.Context, parser *parserpkg.Parser, restartController bool) error {
	return r.C.Runtime.UpdateInstanceWithRillYAML(ctx, r.C.InstanceID, parser, restartController)
}

// reconcileResources creates, updates and deletes resources as necessary to match the parser's output with the current resources in the catalog.
func (r *ProjectParserReconciler) reconcileResources(ctx context.Context, inst *drivers.Instance, self *runtimev1.Resource, parser *parserpkg.Parser) error {
	// Gather resources to delete so we can check for renames.
	var deleteResources []*runtimev1.Resource

	// Pass over all existing resources in the catalog.
	resources, err := r.C.List(ctx, "", "", false)
	if err != nil {
		return err
	}
	seen := make(map[parserpkg.ResourceName]bool, len(resources))
	for _, rr := range resources {
		// Skip if the resource was not created by the parser.
		// If a code file is added for a currently ad-hoc resource, the putParserResourceDef call for it will fail.
		if !equalResourceName(rr.Meta.Owner, self.Meta.Name) {
			continue
		}

		n := runtime.ResourceNameToParser(rr.Meta.Name).Normalized()
		def, ok := parser.Resources[n]

		// If the existing resource is in the parser output, update it.
		// NOTE: putParserResourceDef renames if the casing of the name has changed.
		if ok {
			seen[n] = true
			err = r.putParserResourceDef(ctx, inst, self, def, rr)
			if err != nil {
				return err
			}
			continue
		}

		// If the existing resource is not in the parser output, delete it
		deleteResources = append(deleteResources, rr)
	}

	// Insert resources for the parser outputs that were not seen when passing over the existing resources
	for _, def := range parser.Resources {
		if seen[def.Name.Normalized()] {
			continue
		}

		// Rename if possible
		renamed := false
		for idx, rr := range deleteResources {
			if rr == nil {
				// Already renamed
				continue
			}
			renamed, err = r.attemptRename(ctx, inst, self, def, rr)
			if err != nil {
				return err
			}
			if renamed {
				deleteResources[idx] = nil
				break
			}
		}
		if renamed {
			continue
		}

		// Insert resource
		err = r.putParserResourceDef(ctx, inst, self, def, nil)
		if err != nil {
			return err
		}
	}

	// Delete resources that did not get renamed
	for _, rr := range deleteResources {
		// The ones that got renamed were set to nil
		if rr == nil {
			continue
		}

		err = r.C.Delete(ctx, rr.Meta.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

// reconcileResourcesDiff is similar to reconcileResources, but uses a diff from parser.Reparse instead of doing a full comparison of all resources.
func (r *ProjectParserReconciler) reconcileResourcesDiff(ctx context.Context, inst *drivers.Instance, self *runtimev1.Resource, parser *parserpkg.Parser, diff *parserpkg.Diff) error {
	// Gather resource to delete so we can check for renames.
	deleteResources := make([]*runtimev1.ResourceName, 0, len(diff.Deleted))
	for _, n := range diff.Deleted {
		deleteResources = append(deleteResources, runtime.ResourceNameFromParser(n))
	}

	// Updates
	for _, n := range diff.Modified {
		existing, err := r.C.Get(ctx, runtime.ResourceNameFromParser(n), false)
		if err != nil {
			return err
		}
		def := parser.Resources[n.Normalized()]
		err = r.putParserResourceDef(ctx, inst, self, def, existing)
		if err != nil {
			return err
		}
	}

	// Inserts
	for _, n := range diff.Added {
		def := parser.Resources[n.Normalized()]

		// Rename if possible
		renamed := false
		for idx, rn := range deleteResources {
			if rn == nil {
				// Already renamed
				continue
			}

			rr, err := r.C.Get(ctx, rn, false)
			if err != nil {
				return err
			}

			renamed, err = r.attemptRename(ctx, inst, self, def, rr)
			if err != nil {
				return err
			}
			if renamed {
				deleteResources[idx] = nil
				break
			}
		}
		if renamed {
			continue
		}

		// Insert resource
		err := r.putParserResourceDef(ctx, inst, self, def, nil)
		if err != nil {
			return err
		}
	}

	// Deletes
	for _, rn := range deleteResources {
		// The ones that got renamed were set to nil
		if rn == nil {
			continue
		}

		err := r.C.Delete(ctx, rn)
		if err != nil {
			return err
		}
	}

	return nil
}

// putParserResourceDef creates or updates a resource in the catalog based on a parser resource definition.
// It does an insert if existing is nil, otherwise it does an update.
// If existing is not nil, it compares values and only updates meta/spec values if they have changed (ensuring stable resource version numbers).
func (r *ProjectParserReconciler) putParserResourceDef(ctx context.Context, inst *drivers.Instance, self *runtimev1.Resource, def *parserpkg.Resource, existing *runtimev1.Resource) error {
	// Apply defaults
	def, err := applySpecDefaults(inst, def)
	if err != nil {
		return err
	}

	// Make resource spec to insert/update.
	// res should be nil if no spec changes are needed.
	res := newResourceIfModified(def, existing)

	// Make refs for the resource meta
	refs := make([]*runtimev1.ResourceName, 0, len(def.Refs))
	for _, r := range def.Refs {
		refs = append(refs, runtime.ResourceNameFromParser(r))
	}

	// Create and return if not updating
	n := runtime.ResourceNameFromParser(def.Name)
	if existing == nil {
		return r.C.Create(ctx, n, refs, self.Meta.Name, def.Paths, false, res)
	}

	// Handle changed name and/or path
	if n.Kind == existing.Meta.Name.Kind && n.Name != existing.Meta.Name.Name {
		// The name may have changed to a different case (e.g. aAa -> Aaa).
		// Note that this also updates the paths (updating them separately with UpdateMeta would be considered a mutation of a renamed resource, which requires falling back to a less optimal reconciliation).
		err := r.C.UpdateName(ctx, existing.Meta.Name, n, self.Meta.Name, def.Paths)
		if err != nil {
			return err
		}
	} else if !slices.Equal(existing.Meta.FilePaths, def.Paths) {
		// The path may have been changed. Usually this case is covered in the UpdateName case above because changing a file path usually changes the name.
		err := r.C.UpdateMeta(ctx, n, existing.Meta.Refs, self.Meta.Name, def.Paths)
		if err != nil {
			return err
		}
	}

	// Update meta if refs changed
	if !equalResourceNames(existing.Meta.Refs, refs) {
		err := r.C.UpdateMeta(ctx, n, refs, self.Meta.Name, def.Paths)
		if err != nil {
			return err
		}
	}

	// Update spec if it changed
	if res != nil {
		err := r.C.UpdateMeta(ctx, n, refs, self.Meta.Name, def.Paths)
		if err != nil {
			return err
		}
		err = r.C.UpdateSpec(ctx, n, res)
		if err != nil {
			return err
		}
	}

	return nil
}

// attemptRename renames an existing resource if its spec matches a parser resource definition.
// It returns false if no rename was done.
// In addition to renaming, it also updates the resource's meta to match the parser resource definition.
func (r *ProjectParserReconciler) attemptRename(ctx context.Context, inst *drivers.Instance, self *runtimev1.Resource, def *parserpkg.Resource, existing *runtimev1.Resource) (bool, error) {
	newName := runtime.ResourceNameFromParser(def.Name)
	if existing.Meta.Name.Kind != newName.Kind {
		return false, nil
	}

	// Check refs are the same (note: refs are always sorted)
	if len(existing.Meta.Refs) != len(def.Refs) {
		return false, nil
	}
	for i, n := range existing.Meta.Refs {
		if runtime.ResourceNameToParser(n) != def.Refs[i] {
			return false, nil
		}
	}

	// Apply defaults before comparing specs
	def, err := applySpecDefaults(inst, def)
	if err != nil {
		return false, err
	}

	// Check spec is the same

	switch def.Name.Kind {
	case parserpkg.ResourceKindSource, parserpkg.ResourceKindModel, parserpkg.ResourceKindMetricsView, parserpkg.ResourceKindMigration:
		if newResourceIfModified(def, existing) != nil {
			return false, nil
		}
	default:
		// NOTE: No panic because we don't need to support renames for all resource kinds.
		// If renaming is not supported, it will just do a delete + insert instead.
		return false, nil
	}

	// NOTE: Not comparing owner and paths since changing those are allowed when renaming.

	// Run rename
	err = r.C.UpdateName(ctx, existing.Meta.Name, newName, self.Meta.Name, def.Paths)
	if err != nil {
		return false, err
	}

	return true, nil
}

// applySpecDefaults applies instance-level default properties to a resource spec.
func applySpecDefaults(_ *drivers.Instance, def *parserpkg.Resource) (*parserpkg.Resource, error) {
	// There are no defaults to apply at the moment but it can be used to set defaults in the future.
	return def, nil
}

func equalResourceName(a, b *runtimev1.ResourceName) bool {
	return a != nil && b != nil && a.Kind == b.Kind && strings.EqualFold(a.Name, b.Name)
}

func equalResourceNames(a, b []*runtimev1.ResourceName) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if !equalResourceName(v, b[i]) {
			return false
		}
	}
	return true
}

// newResourceIfModified creates a new resource if the given parser definition is different from the existing resource.
func newResourceIfModified(def *parserpkg.Resource, existing *runtimev1.Resource) *runtimev1.Resource {
	switch def.Name.Kind {
	case parserpkg.ResourceKindSource:
		if existing == nil || !proto.Equal(existing.GetSource().Spec, def.SourceSpec) {
			return &runtimev1.Resource{Resource: &runtimev1.Resource_Source{Source: &runtimev1.Source{Spec: def.SourceSpec}}}
		}
	case parserpkg.ResourceKindModel:
		if existing != nil { // Copy over the ephemeral trigger properties from the existing resource.
			def.ModelSpec.Trigger = existing.GetModel().Spec.Trigger
			def.ModelSpec.TriggerFull = existing.GetModel().Spec.TriggerFull
			def.ModelSpec.TriggerPartitions = existing.GetModel().Spec.TriggerPartitions
		}
		if existing == nil || !proto.Equal(existing.GetModel().Spec, def.ModelSpec) {
			return &runtimev1.Resource{Resource: &runtimev1.Resource_Model{Model: &runtimev1.Model{Spec: def.ModelSpec}}}
		}
	case parserpkg.ResourceKindMetricsView:
		if existing == nil || !proto.Equal(existing.GetMetricsView().Spec, def.MetricsViewSpec) {
			return &runtimev1.Resource{Resource: &runtimev1.Resource_MetricsView{MetricsView: &runtimev1.MetricsView{Spec: def.MetricsViewSpec}}}
		}
	case parserpkg.ResourceKindExplore:
		if existing == nil || !proto.Equal(existing.GetExplore().Spec, def.ExploreSpec) {
			return &runtimev1.Resource{Resource: &runtimev1.Resource_Explore{Explore: &runtimev1.Explore{Spec: def.ExploreSpec}}}
		}
	case parserpkg.ResourceKindMigration:
		if existing == nil || !proto.Equal(existing.GetMigration().Spec, def.MigrationSpec) {
			return &runtimev1.Resource{Resource: &runtimev1.Resource_Migration{Migration: &runtimev1.Migration{Spec: def.MigrationSpec}}}
		}
	case parserpkg.ResourceKindReport:
		if existing != nil { // Copy over the ephemeral trigger property from the existing resource.
			def.ReportSpec.Trigger = existing.GetReport().Spec.Trigger
		}
		if existing == nil || !proto.Equal(existing.GetReport().Spec, def.ReportSpec) {
			return &runtimev1.Resource{Resource: &runtimev1.Resource_Report{Report: &runtimev1.Report{Spec: def.ReportSpec}}}
		}
	case parserpkg.ResourceKindAlert:
		if existing != nil { // Copy over the ephemeral trigger property from the existing resource.
			def.AlertSpec.Trigger = existing.GetAlert().Spec.Trigger
		}
		if existing == nil || !proto.Equal(existing.GetAlert().Spec, def.AlertSpec) {
			return &runtimev1.Resource{Resource: &runtimev1.Resource_Alert{Alert: &runtimev1.Alert{Spec: def.AlertSpec}}}
		}
	case parserpkg.ResourceKindTheme:
		if existing == nil || !proto.Equal(existing.GetTheme().Spec, def.ThemeSpec) {
			return &runtimev1.Resource{Resource: &runtimev1.Resource_Theme{Theme: &runtimev1.Theme{Spec: def.ThemeSpec}}}
		}
	case parserpkg.ResourceKindComponent:
		if existing == nil || !proto.Equal(existing.GetComponent().Spec, def.ComponentSpec) {
			return &runtimev1.Resource{Resource: &runtimev1.Resource_Component{Component: &runtimev1.Component{Spec: def.ComponentSpec}}}
		}
	case parserpkg.ResourceKindCanvas:
		if existing == nil || !proto.Equal(existing.GetCanvas().Spec, def.CanvasSpec) {
			return &runtimev1.Resource{Resource: &runtimev1.Resource_Canvas{Canvas: &runtimev1.Canvas{Spec: def.CanvasSpec}}}
		}
	case parserpkg.ResourceKindAPI:
		if existing == nil || !proto.Equal(existing.GetApi().Spec, def.APISpec) {
			return &runtimev1.Resource{Resource: &runtimev1.Resource_Api{Api: &runtimev1.API{Spec: def.APISpec}}}
		}
	case parserpkg.ResourceKindConnector:
		if existing == nil || !proto.Equal(existing.GetConnector().Spec, def.ConnectorSpec) {
			return &runtimev1.Resource{Resource: &runtimev1.Resource_Connector{Connector: &runtimev1.ConnectorV2{Spec: def.ConnectorSpec}}}
		}
	default:
		panic(fmt.Errorf("unknown resource type %q", def.Name.Kind))
	}

	return nil
}
