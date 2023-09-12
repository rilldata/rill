package runtime

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog"
	"google.golang.org/protobuf/proto"
)

func (r *Runtime) ListCatalogEntries(ctx context.Context, instanceID string, t drivers.ObjectType) ([]*drivers.CatalogEntry, error) {
	cat, err := r.NewCatalogService(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer cat.Close()

	return cat.FindEntries(ctx, t)
}

func (r *Runtime) GetCatalogEntry(ctx context.Context, instanceID, name string) (*drivers.CatalogEntry, error) {
	cat, err := r.NewCatalogService(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer cat.Close()

	e, err := cat.FindEntry(ctx, name)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (r *Runtime) Reconcile(ctx context.Context, instanceID string, changedPaths, forcedPaths []string, dry, strict bool) (*catalog.ReconcileResult, error) {
	cat, err := r.NewCatalogService(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer cat.Close()

	err = cat.Repo.Sync(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := cat.Reconcile(ctx, catalog.ReconcileConfig{
		DryRun:            dry,
		Strict:            strict,
		ChangedPaths:      changedPaths,
		ForcedPaths:       forcedPaths,
		SafeSourceRefresh: r.opts.SafeSourceRefresh,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (r *Runtime) RefreshSource(ctx context.Context, instanceID, name string) error {
	cat, err := r.NewCatalogService(ctx, instanceID)
	if err != nil {
		return err
	}
	defer cat.Close()

	err = cat.Repo.Sync(ctx)
	if err != nil {
		return err
	}

	path, ok := cat.Meta.NameToPath[name]
	if !ok {
		return fmt.Errorf("artifact not found for source")
	}

	resp, err := cat.Reconcile(ctx, catalog.ReconcileConfig{
		ChangedPaths:      []string{path},
		ForcedPaths:       []string{path},
		Strict:            true,
		SafeSourceRefresh: r.opts.SafeSourceRefresh,
	})
	if err != nil {
		return err
	}
	if len(resp.Errors) > 0 {
		return errors.New(resp.Errors[0].Message)
	}

	return nil
}

func (r *Runtime) SyncExistingTables(ctx context.Context, instanceID string) error {
	// TODO: move to using reconcile

	// Get OLAP
	olap, release, err := r.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	// Get catalog
	cat, err := r.NewCatalogService(ctx, instanceID)
	if err != nil {
		return err
	}
	defer cat.Close()

	// Get full catalog
	objs, err := cat.FindEntries(ctx, drivers.ObjectTypeUnspecified)
	if err != nil {
		return err
	}

	// Get information schema
	tables, err := olap.InformationSchema().All(ctx)
	if err != nil {
		return err
	}

	// Index objects for lookup
	objMap := make(map[string]*drivers.CatalogEntry)
	objSeen := make(map[string]bool)
	for _, obj := range objs {
		objMap[obj.Name] = obj
		objSeen[obj.Name] = false
	}

	// Process tables in information schema
	added := 0
	updated := 0
	for _, t := range tables {
		obj, ok := objMap[t.Name]

		// Track that the object still exists
		if ok {
			objSeen[t.Name] = true
		}

		// Create or update in catalog if relevant
		if ok && obj.Type == drivers.ObjectTypeTable && !obj.GetTable().Managed {
			// If the table has already been synced, update the schema if it has changed
			tbl := obj.GetTable()
			if !proto.Equal(t.Schema, tbl.Schema) {
				tbl.Schema = t.Schema
				err := cat.Catalog.UpdateEntry(ctx, obj)
				if err != nil {
					return err
				}
				updated++
			}
		} else if !ok {
			// If we haven't seen this table before, add it
			err := cat.Catalog.CreateEntry(ctx, &drivers.CatalogEntry{
				Name: t.Name,
				Type: drivers.ObjectTypeTable,
				Object: &runtimev1.Table{
					Name:    t.Name,
					Schema:  t.Schema,
					Managed: false,
				},
			})
			if err != nil {
				return err
			}
			added++
		}
		// Defensively do nothing in all other cases
	}

	// Remove non-managed tables not found in information schema
	removed := 0
	for name, seen := range objSeen {
		obj := objMap[name]
		if !seen && obj.Type == drivers.ObjectTypeTable && !obj.GetTable().Managed {
			err := cat.Catalog.DeleteEntry(ctx, name)
			if err != nil {
				return err
			}
			removed++
		}
	}

	// Done
	return nil
}
