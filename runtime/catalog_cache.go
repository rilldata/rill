package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/dag"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// catalogCache is a catalog proxy that caches and edits resources in-memory. It enables rapid reads and writes to the catalog.
// It writes changes to resources to the underlying store when flush() or close() is called.
// It only reads resources from the underlying store on initialization, making the overall workload against the store write-heavy.
//
// catalogCache additionally provides various indexes of the resources: a DAG, map of soft-deleted resources, map of renamed resources.
// It is not thread-safe, but it protects against split-brain scenarios by erroring if the underlying store is mutated by another catalog cache.
// The catalogCache treats resource names as case insensitive.
type catalogCache struct {
	ctrl    *Controller
	store   drivers.CatalogStore
	release func()
	version int64

	resources map[string]map[string]*runtimev1.Resource
	dirty     map[string]*runtimev1.ResourceName
	stored    map[string]bool
	dag       dag.DAG[string, *runtimev1.ResourceName]
	cyclic    map[string]*runtimev1.ResourceName
	renamed   map[string]*runtimev1.ResourceName
	deleted   map[string]*runtimev1.ResourceName

	events      map[string]catalogEvent
	hasEvents   bool
	hasEventsCh chan struct{}
}

// newCatalogCache initializes and warms a new catalog cache.
// It resets ephemeral fields to the following defaults: reconcile_status=idle, renamed_from=nil, reconcile_on=nil.
func newCatalogCache(ctx context.Context, ctrl *Controller, instanceID string) (*catalogCache, error) {
	store, release, err := ctrl.Runtime.Catalog(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	v, err := store.NextControllerVersion(ctx)
	if err != nil {
		return nil, err
	}

	c := &catalogCache{
		ctrl:        ctrl,
		store:       store,
		release:     release,
		version:     v,
		resources:   make(map[string]map[string]*runtimev1.Resource),
		dirty:       make(map[string]*runtimev1.ResourceName),
		stored:      make(map[string]bool),
		dag:         dag.New(nameStr),
		cyclic:      make(map[string]*runtimev1.ResourceName),
		renamed:     make(map[string]*runtimev1.ResourceName),
		deleted:     make(map[string]*runtimev1.ResourceName),
		events:      make(map[string]catalogEvent),
		hasEventsCh: make(chan struct{}, 1),
	}

	rs, err := store.FindResources(ctx)
	if err != nil {
		return nil, err
	}
	for _, r := range rs {
		r2 := resourceFromDriver(r)
		c.link(r2)
		c.stored[nameStr(r2.Meta.Name)] = true
	}

	return c, nil
}

// close closes the catalog cache, flushing any changes to the underlying store.
func (c *catalogCache) close(ctx context.Context) error {
	err := c.flush(ctx)
	c.release()
	return err
}

// flush flushes changes to the underlying store.
// Unlike other catalog functions, it is safe to call flush concurrently with calls to get and list (i.e. under a read lock).
// However, it is not safe to call flush concurrently with other calls to flush.
func (c *catalogCache) flush(ctx context.Context) error {
	for s, n := range c.dirty {
		r, err := c.get(n, true, false)
		if err != nil {
			if !errors.Is(err, drivers.ErrResourceNotFound) {
				return fmt.Errorf("flush: unexpected error from get: %w", err)
			}

			// Resource should be deleted from store
			err = c.store.DeleteResource(ctx, c.version, n.Kind, n.Name)
			if err != nil {
				return err
			}
			delete(c.dirty, s)
			delete(c.stored, s)
			continue
		}

		// Resource should be saved in store
		s := nameStr(r.Meta.Name)
		if c.stored[s] {
			// Updating
			err = c.store.UpdateResource(ctx, c.version, resourceToDriver(r))
			if err != nil {
				return err
			}
		} else {
			// Creating
			err = c.store.CreateResource(ctx, c.version, resourceToDriver(r))
			if err != nil {
				return err
			}
			c.stored[s] = true
		}

		delete(c.dirty, s)
	}

	return nil
}

// get returns a resource from the catalog.
// Unlike other catalog functions, it is safe to call get concurrently with calls to list and flush (i.e. under a read lock).
func (c *catalogCache) get(n *runtimev1.ResourceName, withDeleted, clone bool) (*runtimev1.Resource, error) {
	rs := c.resources[n.Kind]
	if rs == nil {
		return nil, drivers.ErrResourceNotFound
	}
	r, ok := rs[strings.ToLower(n.Name)]
	if !ok {
		return nil, drivers.ErrResourceNotFound
	}
	if r.Meta.DeletedOn != nil && !withDeleted {
		return nil, drivers.ErrResourceNotFound
	}
	if clone {
		return c.clone(r), nil
	}
	return r, nil
}

// list returns a list of resources in the catalog.
// It optionally supports filtering by kind, path, and soft-deleted status.
// The returned list is not sorted.
// The returned list is always safe to manipulate (e.g. sort/filter), but the resource pointers must not be edited unless clone=true.
// Unlike other catalog functions, it is safe to call list concurrently with calls to get and flush (i.e. under a read lock).
func (c *catalogCache) list(kind, path string, withDeleted, clone bool) ([]*runtimev1.Resource, error) {
	// Estimate number of resources to list
	n := 0
	if path != "" {
		n = 1
	} else if kind == "" {
		for _, rs := range c.resources {
			n += len(rs)
		}
	} else {
		n = len(c.resources[kind])
	}

	// Alloc slice for resources
	res := make([]*runtimev1.Resource, 0, n)

	// Find resources matching the filters and append to slice
	for k, rs := range c.resources {
		// Skip if doesn't match kind filter
		if kind != "" && k != kind {
			continue
		}

		for _, r := range rs {
			// Skip if doesn't match withDeleted filter
			if !withDeleted && r.Meta.DeletedOn != nil {
				continue
			}

			// Skip if doesn't match path filter
			if path != "" {
				found := false
				for _, p := range r.Meta.FilePaths {
					if p == path {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// Append to res
			if clone {
				r = c.clone(r)
			}
			res = append(res, r)
		}
	}

	return res, nil
}

// create creates a resource in the catalog.
// It will error if a resource with the same name already exists.
// If a soft-deleted resource exists with the same name, it will be overwritten (no longer deleted).
// The passed resource should only have its spec populated. The meta and state fields will be populated by this function.
func (c *catalogCache) create(name *runtimev1.ResourceName, refs []*runtimev1.ResourceName, owner *runtimev1.ResourceName, paths []string, hidden bool, r *runtimev1.Resource) error {
	existing, _ := c.get(name, true, false)
	if existing != nil {
		if existing.Meta.DeletedOn == nil {
			return drivers.ErrResourceAlreadyExists
		}
		c.unlink(existing) // If creating a resource that's currently soft-deleted, it'll be like the previous delete never happened.
	}
	r.Meta = &runtimev1.ResourceMeta{
		Name:            name,
		Refs:            refs,
		Owner:           owner,
		FilePaths:       paths,
		Hidden:          hidden,
		Version:         1,
		SpecVersion:     1,
		StateVersion:    1,
		CreatedOn:       timestamppb.Now(),
		SpecUpdatedOn:   timestamppb.Now(),
		StateUpdatedOn:  timestamppb.Now(),
		ReconcileStatus: runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE,
	}
	if existing != nil {
		r.Meta.Version = existing.Meta.Version + 1
		r.Meta.SpecVersion = existing.Meta.SpecVersion + 1
	}
	err := c.ctrl.reconciler(name.Kind).AssignSpec(r, r) // Self-assign spec through reconciler to ensure spec and kind have matching types
	if err != nil {
		return err
	}
	err = c.ctrl.reconciler(name.Kind).ResetState(r)
	if err != nil {
		return err
	}
	c.link(r)
	c.dirty[nameStr(r.Meta.Name)] = r.Meta.Name
	c.addEvent(name, r, runtimev1.ResourceEvent_RESOURCE_EVENT_WRITE)
	return nil
}

// rename renames a resource in the catalog and sets the r.Meta.RenamedFrom field.
func (c *catalogCache) rename(name, newName *runtimev1.ResourceName) error {
	r, err := c.get(name, false, false)
	if err != nil {
		return err
	}
	c.unlink(r)
	r.Meta.RenamedFrom = r.Meta.Name
	r.Meta.Name = newName
	r.Meta.Version++
	r.Meta.SpecVersion++
	r.Meta.SpecUpdatedOn = timestamppb.Now()
	c.link(r)
	c.dirty[nameStr(r.Meta.RenamedFrom)] = r.Meta.RenamedFrom
	c.dirty[nameStr(r.Meta.Name)] = r.Meta.Name
	c.addEvent(r.Meta.RenamedFrom, nil, runtimev1.ResourceEvent_RESOURCE_EVENT_DELETE)
	c.addEvent(r.Meta.Name, r, runtimev1.ResourceEvent_RESOURCE_EVENT_WRITE)
	return nil
}

// clearRenamedFrom clears the r.Meta.RenamedFrom field without bumping version numbers.
func (c *catalogCache) clearRenamedFrom(name *runtimev1.ResourceName) error {
	r, err := c.get(name, false, false)
	if err != nil {
		return err
	}
	if r.Meta.RenamedFrom == nil {
		return nil
	}
	c.unlink(r)
	r.Meta.RenamedFrom = nil
	c.link(r)
	c.dirty[nameStr(r.Meta.Name)] = r.Meta.Name
	c.addEvent(r.Meta.Name, r, runtimev1.ResourceEvent_RESOURCE_EVENT_WRITE)
	return nil
}

// updateMeta updates the meta fields of a resource.
func (c *catalogCache) updateMeta(name *runtimev1.ResourceName, refs []*runtimev1.ResourceName, owner *runtimev1.ResourceName, paths []string) error {
	r, err := c.get(name, false, false)
	if err != nil {
		return err
	}
	c.unlink(r)
	r.Meta.Refs = refs
	r.Meta.Owner = owner
	r.Meta.FilePaths = paths
	r.Meta.Version++
	r.Meta.SpecVersion++
	r.Meta.SpecUpdatedOn = timestamppb.Now()
	c.link(r)
	c.dirty[nameStr(r.Meta.Name)] = r.Meta.Name
	c.addEvent(r.Meta.Name, r, runtimev1.ResourceEvent_RESOURCE_EVENT_WRITE)
	return nil
}

// updateSpec updates the spec field of a resource.
// It uses the spec from the passed resource and disregards its other fields.
func (c *catalogCache) updateSpec(name *runtimev1.ResourceName, from *runtimev1.Resource) error {
	r, err := c.get(name, false, false)
	if err != nil {
		return err
	}
	// NOTE: No need to unlink/link because no indexed fields are edited.
	err = c.ctrl.reconciler(name.Kind).AssignSpec(from, r)
	if err != nil {
		return err
	}
	r.Meta.Version++
	r.Meta.SpecVersion++
	r.Meta.SpecUpdatedOn = timestamppb.Now()
	c.dirty[nameStr(r.Meta.Name)] = r.Meta.Name
	c.addEvent(r.Meta.Name, r, runtimev1.ResourceEvent_RESOURCE_EVENT_WRITE)
	return nil
}

// updateState updates the state field of a resource.
// It uses the state from the passed resource and disregards its other fields.
func (c *catalogCache) updateState(name *runtimev1.ResourceName, from *runtimev1.Resource) error {
	r, err := c.get(name, true, false)
	if err != nil {
		return err
	}
	// NOTE: No need to unlink/link because no indexed fields are edited.
	err = c.ctrl.reconciler(name.Kind).AssignState(from, r)
	if err != nil {
		return err
	}
	r.Meta.Version++
	r.Meta.StateVersion++
	r.Meta.StateUpdatedOn = timestamppb.Now()
	c.dirty[nameStr(r.Meta.Name)] = r.Meta.Name
	c.addEvent(r.Meta.Name, r, runtimev1.ResourceEvent_RESOURCE_EVENT_WRITE)
	return nil
}

// updateError updates the reconcile_error field of a resource.
func (c *catalogCache) updateError(name *runtimev1.ResourceName, reconcileErr error) error {
	r, err := c.get(name, true, false)
	if err != nil {
		return err
	}
	var errStr string
	if reconcileErr != nil {
		errStr = reconcileErr.Error()
	}
	if r.Meta.ReconcileError == errStr {
		// Since bumping the state version usually invalidates derived things, we don't want to do it redundantly.
		return nil
	}
	// NOTE: No need to unlink/link because no indexed fields are edited.
	r.Meta.ReconcileError = errStr
	r.Meta.Version++
	r.Meta.StateVersion++
	r.Meta.StateUpdatedOn = timestamppb.Now()
	c.dirty[nameStr(r.Meta.Name)] = r.Meta.Name
	c.addEvent(r.Meta.Name, r, runtimev1.ResourceEvent_RESOURCE_EVENT_WRITE)
	return nil
}

// updateDeleted sets the deleted_on field of a resource (a soft delete).
// Afterwards, the resource can still be accessed by passing withDeleted to the getters.
func (c *catalogCache) updateDeleted(name *runtimev1.ResourceName) error {
	r, err := c.get(name, false, false)
	if err != nil {
		return err
	}
	c.unlink(r)
	r.Meta.DeletedOn = timestamppb.Now()
	r.Meta.Version++
	r.Meta.SpecVersion++
	r.Meta.SpecUpdatedOn = timestamppb.Now()
	c.link(r)
	c.dirty[nameStr(r.Meta.Name)] = r.Meta.Name
	c.addEvent(r.Meta.Name, r, runtimev1.ResourceEvent_RESOURCE_EVENT_WRITE)
	return nil
}

// updateStatus updates the ephemeral status fields on a resource.
// The values of these fields are reset next time a catalog cache is created.
func (c *catalogCache) updateStatus(name *runtimev1.ResourceName, status runtimev1.ReconcileStatus, reconcileOn time.Time) error {
	r, err := c.get(name, true, false)
	if err != nil {
		return err
	}
	r.Meta.ReconcileStatus = status
	if reconcileOn.IsZero() {
		r.Meta.ReconcileOn = nil
	} else {
		r.Meta.ReconcileOn = timestamppb.New(reconcileOn)
	}
	c.addEvent(r.Meta.Name, r, runtimev1.ResourceEvent_RESOURCE_EVENT_WRITE)
	return nil
}

// delete permanently deletes a resource from the catalog (a hard delete).
// Afterwards, the resource can no longer be accessed.
func (c *catalogCache) delete(name *runtimev1.ResourceName) error {
	r, err := c.get(name, true, false)
	if err != nil {
		return err
	}
	c.unlink(r)
	c.dirty[nameStr(r.Meta.Name)] = r.Meta.Name
	c.addEvent(r.Meta.Name, nil, runtimev1.ResourceEvent_RESOURCE_EVENT_DELETE)
	return nil
}

// link adds a resource to the cache and its indexes.
func (c *catalogCache) link(r *runtimev1.Resource) {
	if c.resources[r.Meta.Name.Kind] == nil {
		c.resources[r.Meta.Name.Kind] = make(map[string]*runtimev1.Resource)
	}
	c.resources[r.Meta.Name.Kind][strings.ToLower(r.Meta.Name.Name)] = r

	s := nameStr(r.Meta.Name)

	if r.Meta.DeletedOn == nil {
		ok := c.dag.Add(r.Meta.Name, r.Meta.Refs...)
		if !ok {
			c.cyclic[s] = r.Meta.Name
		}
	} else {
		c.deleted[s] = r.Meta.Name
	}

	if r.Meta.RenamedFrom != nil {
		c.renamed[s] = r.Meta.Name
	}
}

// unlink reverses a previous call to link.
func (c *catalogCache) unlink(r *runtimev1.Resource) {
	s := nameStr(r.Meta.Name)

	if r.Meta.RenamedFrom != nil {
		delete(c.renamed, s)
	}

	if r.Meta.DeletedOn == nil {
		if _, ok := c.cyclic[s]; ok {
			delete(c.cyclic, s)
		} else {
			c.dag.Remove(r.Meta.Name)
		}
	} else {
		delete(c.deleted, s)
	}

	delete(c.resources[r.Meta.Name.Kind], strings.ToLower(r.Meta.Name.Name))
}

// clone clones a resource such that it is safe to mutate without affecting a cached resource.
func (c *catalogCache) clone(r *runtimev1.Resource) *runtimev1.Resource {
	return proto.Clone(r).(*runtimev1.Resource)
}

// isCyclic returns true if the resource has cyclic references.
// It must be checked before accessing the resource in c.dag (since cyclic resources are not added to the DAG).
func (c *catalogCache) isCyclic(n *runtimev1.ResourceName) bool {
	_, ok := c.cyclic[nameStr(n)]
	return ok
}

// retryCyclicRefs attempts to re-link resources into the DAG that were previously rejected due to cyclic references.
// It returns a list of resource names that were successfully linked into the DAG.
func (c *catalogCache) retryCyclicRefs() []*runtimev1.ResourceName {
	var res []*runtimev1.ResourceName
	for s, n := range c.cyclic {
		r, err := c.get(n, false, false)
		if err != nil {
			panic(err)
		}
		ok := c.dag.Add(n, r.Meta.Refs...)
		if ok {
			delete(c.cyclic, s)
			res = append(res, n)
		}
	}
	return res
}

// catalogEvent represents a change to a resource.
type catalogEvent struct {
	event    runtimev1.ResourceEvent
	name     *runtimev1.ResourceName
	resource *runtimev1.Resource // will be nil if the event is a deletion
}

// addEvent tracks when a resource has been changed.
func (c *catalogCache) addEvent(n *runtimev1.ResourceName, r *runtimev1.Resource, e runtimev1.ResourceEvent) {
	c.events[nameStr(n)] = catalogEvent{
		event:    e,
		name:     n,
		resource: r,
	}
	if !c.hasEvents {
		c.hasEvents = true
		c.hasEventsCh <- struct{}{}
	}
}

// clearEvents clears all buffered events.
// It should be called after consuming c.events.
// It initializes a new event buffer on each call, so it is safe to use a copy of c.events after calling this method.
func (c *catalogCache) resetEvents() {
	c.hasEvents = false
	c.events = make(map[string]catalogEvent)
}

// nameStr returns a normalized string representation of a resource name.
func nameStr(r *runtimev1.ResourceName) string {
	return fmt.Sprintf("%s/%s", r.Kind, strings.ToLower(r.Name))
}

// resourceFromDriver converts a drivers.Resource to a proto resource.
func resourceFromDriver(r drivers.Resource) *runtimev1.Resource {
	res := &runtimev1.Resource{}
	err := proto.Unmarshal(r.Data, res)
	if err != nil {
		panic(err)
	}

	// Reset ephemeral fields.
	res.Meta.ReconcileStatus = runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE
	res.Meta.ReconcileError = ""
	res.Meta.ReconcileOn = nil
	res.Meta.RenamedFrom = nil

	return res
}

// resourceToDriver converts a proto resource to a drivers.Resource.
func resourceToDriver(r *runtimev1.Resource) drivers.Resource {
	data, err := proto.Marshal(r)
	if err != nil {
		panic(err)
	}

	return drivers.Resource{
		Kind: r.Meta.Name.Kind,
		Name: r.Meta.Name.Name,
		Data: data,
	}
}
