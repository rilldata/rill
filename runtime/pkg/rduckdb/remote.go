package rduckdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"gocloud.dev/gcerrors"
	"golang.org/x/sync/errgroup"
)

// pullFromRemote updates local data with the latest data from remote.
// This is not safe for concurrent calls.
func (d *db) pullFromRemote(ctx context.Context, updateCatalog bool) error {
	if !d.localDirty || d.remote == nil {
		// optimisation to skip sync if write was already synced
		if !updateCatalog {
			// cleanup of older versions of table
			_ = d.iterateLocalTables(true, func(name string, meta *tableMeta) error {
				return nil
			})
		}
		return nil
	}
	d.logger.Debug("syncing from remote", observability.ZapCtx(ctx))
	// Create an errgroup for background downloads with limited concurrency.
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(8)

	objects := d.remote.List(&blob.ListOptions{
		Delimiter: "/", // only list directories with a trailing slash and IsDir set to true
	})

	remoteTables := make(map[string]*tableMeta)
	for {
		// Stop the loop if the ctx was cancelled
		var stop bool
		select {
		case <-gctx.Done():
			stop = true
		default:
			// don't break
		}
		if stop {
			break // can't use break inside the select
		}

		obj, err := objects.Next(gctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		if !obj.IsDir {
			continue
		}

		table := strings.TrimSuffix(obj.Key, "/")
		d.logger.Debug("SyncWithObjectStorage: discovered table", zap.String("table", table), observability.ZapCtx(ctx))

		// get version of the table
		var b []byte
		err = retry(gctx, func() error {
			res, err := d.remote.ReadAll(gctx, path.Join(table, "meta.json"))
			if err != nil {
				return err
			}
			b = res
			return nil
		})
		if err != nil {
			if gcerrors.Code(err) == gcerrors.NotFound {
				// invalid table directory
				d.logger.Debug("SyncWithObjectStorage: invalid table directory", zap.String("table", table), observability.ZapCtx(ctx))
				continue
			}
			return err
		}
		remoteMeta := &tableMeta{}
		err = json.Unmarshal(b, remoteMeta)
		if err != nil {
			d.logger.Debug("SyncWithObjectStorage: failed to unmarshal table metadata", zap.String("table", table), zap.Error(err), observability.ZapCtx(ctx))
			continue
		}
		remoteTables[table] = remoteMeta

		// check if table is locally present
		meta, _ := d.tableMeta(table)
		if meta != nil && meta.Version == remoteMeta.Version {
			d.logger.Debug("SyncWithObjectStorage: local table is in sync with remote", zap.String("table", table), observability.ZapCtx(ctx))
			continue
		}
		if err := d.initLocalTable(table, remoteMeta.Version); err != nil {
			return err
		}

		tblIter := d.remote.List(&blob.ListOptions{Prefix: path.Join(table, remoteMeta.Version)})
		// download all objects in the table and current version
		for {
			obj, err := tblIter.Next(gctx)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
			g.Go(func() error {
				return retry(gctx, func() error {
					file, err := os.Create(filepath.Join(d.localPath, obj.Key))
					if err != nil {
						return err
					}
					defer file.Close()

					rdr, err := d.remote.NewReader(gctx, obj.Key, nil)
					if err != nil {
						return err
					}
					defer rdr.Close()

					_, err = io.Copy(file, rdr)
					return err
				})
			})
		}
	}

	// Wait for all outstanding downloads to complete
	err := g.Wait()
	if err != nil {
		return err
	}

	// Update table versions(updates even if local is same as remote)
	for table, meta := range remoteTables {
		err = d.writeTableMeta(table, meta)
		if err != nil {
			return err
		}
	}

	if !updateCatalog {
		// delete all local tables which are not present in remote
		_ = d.iterateLocalTables(true, func(name string, meta *tableMeta) error {
			if _, ok := remoteTables[name]; !ok {
				return d.deleteLocalTableFiles(name, "")
			}
			return nil
		})
		return nil
	}

	// iterate over all remote tables and update catalog
	for table, remoteMeta := range remoteTables {
		meta, err := d.catalog.tableMeta(table)
		if err != nil {
			if errors.Is(err, errNotFound) {
				// table not found in catalog
				d.catalog.addTableVersion(table, remoteMeta, true)
			}
			return err
		}
		// table is present in catalog but has version mismatch
		if meta.Version != remoteMeta.Version {
			d.catalog.addTableVersion(table, remoteMeta, true)
		}
	}

	// iterate over local entries and remove if not present in remote
	_ = d.iterateLocalTables(false, func(name string, meta *tableMeta) error {
		if _, ok := remoteTables[name]; ok {
			// table is present in remote
			return nil
		}
		// check if table is present in catalog
		_, err := d.catalog.tableMeta(name)
		if err != nil {
			return d.deleteLocalTableFiles(name, "")
		}
		// remove table from catalog
		d.catalog.removeTable(name)
		return nil
	})
	return nil
}

// pushToRemote syncs the remote location with the local path for given table.
// If oldVersion is specified, it is deleted after successful sync.
func (d *db) pushToRemote(ctx context.Context, table string, oldMeta, meta *tableMeta) error {
	if d.remote == nil {
		return nil
	}
	if meta.Type == "TABLE" {
		localPath := d.localTableDir(table, meta.Version)
		entries, err := os.ReadDir(localPath)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			d.logger.Debug("replicating file", zap.String("file", entry.Name()), zap.String("path", localPath), observability.ZapCtx(ctx))
			// no directory should exist as of now
			if entry.IsDir() {
				d.logger.Debug("found directory in path which should not exist", zap.String("file", entry.Name()), zap.String("path", localPath), observability.ZapCtx(ctx))
				continue
			}

			wr, err := os.Open(filepath.Join(localPath, entry.Name()))
			if err != nil {
				return err
			}

			// upload to cloud storage
			err = retry(ctx, func() error {
				return d.remote.Upload(ctx, path.Join(table, meta.Version, entry.Name()), wr, &blob.WriterOptions{
					ContentType: "application/octet-stream",
				})
			})
			_ = wr.Close()
			if err != nil {
				return err
			}
		}
	}

	// update table meta
	// todo :: also use etag to avoid concurrent writer conflicts
	d.localDirty = true
	m, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal table metadata: %w", err)
	}
	err = retry(ctx, func() error {
		return d.remote.WriteAll(ctx, path.Join(table, "meta.json"), m, nil)
	})
	if err != nil {
		d.logger.Error("failed to update meta.json in remote", zap.String("table", table), zap.Error(err), observability.ZapCtx(ctx))
	}

	// success -- remove old version
	if oldMeta != nil {
		_ = d.deleteRemote(ctx, table, oldMeta.Version)
	}
	return err
}

// deleteRemote deletes remote.
// If table is specified, only that table is deleted.
// If table and version is specified, only that version of the table is deleted.
func (d *db) deleteRemote(ctx context.Context, table, version string) error {
	if d.remote == nil {
		return nil
	}
	if table == "" && version != "" {
		return fmt.Errorf("table must be specified if version is specified")
	}
	d.logger.Debug("deleting remote", zap.String("table", table), zap.String("version", version), observability.ZapCtx(ctx))
	var prefix string
	if table != "" {
		if version != "" {
			prefix = path.Join(table, version) + "/"
		} else {
			prefix = table + "/"
			// delete meta.json first
			err := retry(ctx, func() error { return d.remote.Delete(ctx, "meta.json") })
			if err != nil && gcerrors.Code(err) != gcerrors.NotFound {
				d.logger.Error("failed to delete meta.json in remote", zap.String("table", table), zap.Error(err), observability.ZapCtx(ctx))
				return err
			}
		}
	}
	// ignore errors since meta.json is already removed

	iter := d.remote.List(&blob.ListOptions{Prefix: prefix})
	for {
		obj, err := iter.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			d.logger.Debug("failed to list object", zap.String("table", table), zap.Error(err), observability.ZapCtx(ctx))
		}
		err = retry(ctx, func() error { return d.remote.Delete(ctx, obj.Key) })
		if err != nil {
			d.logger.Debug("failed to delete object", zap.String("table", table), zap.String("object", obj.Key), zap.Error(err), observability.ZapCtx(ctx))
		}
	}
	return nil
}

func retry(ctx context.Context, fn func() error) error {
	var err error
	for i := 0; i < _maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil // success
		}
		if !strings.Contains(err.Error(), "stream error: stream ID") {
			break // break and return error
		}

		select {
		case <-ctx.Done():
			return ctx.Err() // return on context cancellation
		case <-time.After(_retryDelay):
		}
	}
	return err
}

const (
	_maxRetries = 5
	_retryDelay = 10 * time.Second
)
