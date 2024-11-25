/*
Example init logic:
- Sync remote files with the local cache
- Create a catalog
- Traverse the local files and call addTableVersion for table
Example write logic:
- Call addTableVersion after adding a new table version
- Call removeTable when deleting a table
Example read logic:
- Call acquireSnapshot when starting a read
- If it doesn't already exist, create a schema for the snapshot ID with views for all the table version in the snapshot
- Call releaseSnapshot when done reading the snapshot
Example removeFunc logic:
- Detach the version
- Remove the version file
- If there are no files left in it, remove the table folder
*/
package rduckdb

import (
	"context"
	"fmt"

	"golang.org/x/sync/semaphore"
)

// Represents one table and its versions currently present in the local cache.
type table struct {
	name                   string
	deleted                bool
	currentVersion         string
	versionReferenceCounts map[string]int
	versionMeta            map[string]*tableMeta
}

// Represents a snapshot of table versions.
// The table versions referenced by the snapshot are guaranteed to exist for as long as the snapshot is acquired.
type snapshot struct {
	id             int
	referenceCount int
	tables         []*tableMeta
	// if snapshot is ready to be served then ready will be marked true
	ready bool
}

// Represents a catalog of available table versions.
// It is thread-safe and supports acquiring a snapshot of table versions which will not be mutated or removed for as long as the snapshot is held.
type catalog struct {
	sem               *semaphore.Weighted
	tables            map[string]*table
	snapshots         map[int]*snapshot
	currentSnapshotID int

	removeVersionFunc  func(context.Context, string, string) error
	removeSnapshotFunc func(context.Context, int) error
}

// newCatalog creates a new catalog.
// The removeSnapshotFunc func will be called exactly once for each snapshot ID when it is no longer the current snapshot and is no longer held by any readers.
// The removeVersionFunc func will be called exactly once for each table version when it is no longer the current version and is no longer used by any active snapshots.
func newCatalog(removeVersionFunc func(context.Context, string, string) error, removeSnapshotFunc func(context.Context, int) error) *catalog {
	return &catalog{
		sem:                semaphore.NewWeighted(1),
		tables:             make(map[string]*table),
		snapshots:          make(map[int]*snapshot),
		removeVersionFunc:  removeVersionFunc,
		removeSnapshotFunc: removeSnapshotFunc,
	}
}

func (c *catalog) tableMeta(ctx context.Context, name string) (*tableMeta, error) {
	err := c.sem.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	defer c.sem.Release(1)

	t, ok := c.tables[name]
	if !ok || t.deleted {
		return nil, errNotFound
	}
	meta, ok := t.versionMeta[t.currentVersion]
	if !ok {
		return nil, fmt.Errorf("internal error: meta for version %q not found", t.currentVersion)
	}
	return meta, nil
}

// addTableVersion registers a new version of a table.
// If the table name has not been seen before, it is added to the catalog.
func (c *catalog) addTableVersion(ctx context.Context, name string, meta *tableMeta) error {
	err := c.sem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer c.sem.Release(1)

	t, ok := c.tables[name]
	if !ok {
		t = &table{
			name:                   name,
			versionReferenceCounts: make(map[string]int),
			versionMeta:            make(map[string]*tableMeta),
		}
		c.tables[name] = t
	}

	oldVersion := t.currentVersion
	t.deleted = false // In case the table was deleted previously, but a snapshot still references it.
	t.currentVersion = meta.Version
	t.versionMeta[meta.Version] = meta
	c.acquireVersion(t, t.currentVersion)
	if oldVersion != "" {
		_ = c.releaseVersion(ctx, t, oldVersion)
	}

	c.currentSnapshotID++
	return nil
}

// removeTable removes a table from the catalog.
// If the table is currently used by a snapshot, it will stay in the catalog but marked with deleted=true.
// When the last snapshot referencing the table is released, the table will be removed completely.
func (c *catalog) removeTable(ctx context.Context, name string) error {
	err := c.sem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer c.sem.Release(1)

	t, ok := c.tables[name]
	if !ok {
		return fmt.Errorf("table %q not found", name)
	}

	oldVersion := t.currentVersion
	t.deleted = true
	t.currentVersion = ""
	return c.releaseVersion(ctx, t, oldVersion)
}

// listTables returns tableMeta for all active tables present in the catalog.
func (c *catalog) listTables(ctx context.Context) ([]*tableMeta, error) {
	err := c.sem.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	defer c.sem.Release(1)

	tables := make([]*tableMeta, 0)
	for _, t := range c.tables {
		if t.deleted {
			continue
		}
		meta, ok := t.versionMeta[t.currentVersion]
		if !ok {
			return nil, fmt.Errorf("internal error: meta for version %q not found", t.currentVersion)
		}
		tables = append(tables, meta)
	}
	return tables, nil
}

// acquireSnapshot acquires a snapshot of the current table versions.
func (c *catalog) acquireSnapshot(ctx context.Context) (*snapshot, error) {
	err := c.sem.Acquire(ctx, 1)
	if err != nil {
		return nil, err
	}
	defer c.sem.Release(1)

	s, ok := c.snapshots[c.currentSnapshotID]
	if ok {
		s.referenceCount++
		return s, nil
	}
	// first acquire
	s = &snapshot{
		id:             c.currentSnapshotID,
		referenceCount: 1,
		tables:         make([]*tableMeta, 0),
	}
	for _, t := range c.tables {
		if t.deleted {
			continue
		}

		meta, ok := t.versionMeta[t.currentVersion]
		if !ok {
			return nil, fmt.Errorf("internal error: meta for version %q not found", t.currentVersion)
		}
		s.tables = append(s.tables, meta)
		c.acquireVersion(t, t.currentVersion)
	}
	c.snapshots[c.currentSnapshotID] = s
	return s, nil
}

// releaseSnapshot releases a snapshot of table versions.
func (c *catalog) releaseSnapshot(ctx context.Context, s *snapshot) error {
	err := c.sem.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	defer c.sem.Release(1)

	s.referenceCount--
	if s.referenceCount > 0 {
		return nil
	}

	for _, meta := range s.tables {
		t, ok := c.tables[meta.Name]
		if !ok {
			return fmt.Errorf("internal error: table %q not found", meta.Name)
		}
		if err := c.releaseVersion(ctx, t, meta.Version); err != nil {
			return err
		}
	}

	delete(c.snapshots, s.id)
	return c.removeSnapshotFunc(ctx, s.id)
}

// acquireVersion increments the reference count of a table version.
// It must be called while holding the catalog mutex.
func (c *catalog) acquireVersion(t *table, version string) {
	referenceCount := t.versionReferenceCounts[version]
	referenceCount++
	t.versionReferenceCounts[version] = referenceCount
}

// releaseVersion decrements the reference count of a table version.
// If the reference count reaches zero and the version is no longer the current version, it is removec.
func (c *catalog) releaseVersion(ctx context.Context, t *table, version string) error {
	referenceCount, ok := t.versionReferenceCounts[version]
	if !ok {
		return fmt.Errorf("version %q of table %q not found", version, t.name)
	}
	referenceCount--
	if referenceCount > 0 {
		t.versionReferenceCounts[version] = referenceCount
		return nil
	}

	delete(t.versionReferenceCounts, version)
	if t.deleted && len(t.versionReferenceCounts) == 0 {
		delete(c.tables, t.name)
	}

	return c.removeVersionFunc(ctx, t.name, version)
}
