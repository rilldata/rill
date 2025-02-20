package rduckdb

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
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
	mu                sync.Mutex
	tables            map[string]*table
	snapshots         map[int]*snapshot
	currentSnapshotID int

	removeVersionFunc  func(string, string)
	removeSnapshotFunc func(int)

	logger *zap.Logger
}

// newCatalog creates a new catalog.
// The removeSnapshotFunc func will be called exactly once for each snapshot ID when it is no longer the current snapshot and is no longer held by any readers.
// The removeVersionFunc func will be called exactly once for each table version when it is no longer the current version and is no longer used by any active snapshots.
func newCatalog(removeVersionFunc func(string, string), removeSnapshotFunc func(int), tables []*tableMeta, logger *zap.Logger) *catalog {
	c := &catalog{
		tables:             make(map[string]*table),
		snapshots:          make(map[int]*snapshot),
		removeVersionFunc:  removeVersionFunc,
		removeSnapshotFunc: removeSnapshotFunc,
		logger:             logger,
	}
	for _, meta := range tables {
		c.addTableVersion(meta.Name, meta, false)
	}
	c.incrementSnapshotUnsafe()
	return c
}

func (c *catalog) tableMeta(name string) (*tableMeta, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	t, ok := c.tables[name]
	if !ok || t.deleted {
		return nil, errNotFound
	}
	meta, ok := t.versionMeta[t.currentVersion]
	if !ok {
		panic(fmt.Errorf("internal error: meta for table %q and version %q not found", name, t.currentVersion))
	}
	return meta, nil
}

// addTableVersion registers a new version of a table.
// If the table name has not been seen before, it is added to the catalog.
func (c *catalog) addTableVersion(name string, meta *tableMeta, incrementSnapshot bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

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
		c.releaseVersion(t, oldVersion)
	}

	if incrementSnapshot {
		c.incrementSnapshotUnsafe()
	}
}

// removeTable removes a table from the catalog.
// If the table is currently used by a snapshot, it will stay in the catalog but marked with deleted=true.
// When the last snapshot referencing the table is released, the table will be removed completely.
func (c *catalog) removeTable(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	t, ok := c.tables[name]
	if !ok {
		c.logger.Debug("table not found in rduckdb catalog", zap.String("name", name))
		return
	}

	oldVersion := t.currentVersion
	t.deleted = true
	t.currentVersion = ""

	c.releaseVersion(t, oldVersion)
	c.incrementSnapshotUnsafe()
}

// listTables returns tableMeta for all active tables present in the catalog.
func (c *catalog) listTables() []*tableMeta {
	c.mu.Lock()
	defer c.mu.Unlock()

	tables := make([]*tableMeta, 0)
	for _, t := range c.tables {
		if t.deleted {
			continue
		}
		meta, ok := t.versionMeta[t.currentVersion]
		if !ok {
			c.logger.Error("internal error: meta for table not found in catalog", zap.String("name", t.name), zap.String("version", t.currentVersion))
		}
		tables = append(tables, meta)
	}
	return tables
}

// incrementSnapshotUnsafe increments the current snapshot.
// It ensures that the currentSnapshotID always has at least one reference.
func (c *catalog) incrementSnapshotUnsafe() {
	// Increment snapshot ID
	c.currentSnapshotID++

	// Acquire new current snapshot
	c.acquireSnapshotUnsafe()

	// Release previous current snapshot
	if c.currentSnapshotID > 1 {
		c.releaseSnapshotUnsafe(c.snapshots[c.currentSnapshotID-1])
	}
}

// acquireSnapshot acquires a snapshot of the current table versions.
func (c *catalog) acquireSnapshot() *snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.acquireSnapshotUnsafe()
}

func (c *catalog) acquireSnapshotUnsafe() *snapshot {
	s, ok := c.snapshots[c.currentSnapshotID]
	if ok {
		s.referenceCount++
		return s
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
			panic(fmt.Errorf("internal error: meta for table %q version %q not found in catalog", t.name, t.currentVersion))
		}
		s.tables = append(s.tables, meta)
		c.acquireVersion(t, t.currentVersion)
	}
	c.snapshots[c.currentSnapshotID] = s
	return s
}

// releaseSnapshot releases a snapshot of table versions.
func (c *catalog) releaseSnapshot(s *snapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.releaseSnapshotUnsafe(s)
}

func (c *catalog) releaseSnapshotUnsafe(s *snapshot) {
	s.referenceCount--
	if s.referenceCount > 0 || s.id == c.currentSnapshotID {
		return
	}

	for _, meta := range s.tables {
		t, ok := c.tables[meta.Name]
		if !ok {
			panic(fmt.Errorf("internal error: table %q not found in catalog", meta.Name))
		}
		c.releaseVersion(t, meta.Version)
	}
	// delete the older snapshot
	delete(c.snapshots, s.id)
	c.removeSnapshotFunc(s.id)
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
func (c *catalog) releaseVersion(t *table, version string) {
	referenceCount, ok := t.versionReferenceCounts[version]
	if !ok {
		panic(fmt.Errorf("internal error: version %q of table %q not found in catalog", t.currentVersion, t.name))
	}
	referenceCount--
	if referenceCount > 0 {
		t.versionReferenceCounts[version] = referenceCount
		return
	}

	delete(t.versionReferenceCounts, version)
	if t.deleted && len(t.versionReferenceCounts) == 0 {
		delete(c.tables, t.name)
	}
	if t.currentVersion == version {
		// do not remove the current version
		return
	}
	c.removeVersionFunc(t.name, version)
}
