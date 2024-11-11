package rduckdb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"gocloud.dev/blob"
	"gocloud.dev/gcerrors"
	"golang.org/x/sync/errgroup"
)

// syncLocalWithBackup syncs the write path with the backup location.
// This is not safe for concurrent calls.
func (d *db) syncLocalWithBackup(ctx context.Context) error {
	if !d.writeDirty || d.backup == nil {
		// optimisation to skip sync if write was already synced
		return nil
	}
	d.logger.Debug("syncing from backup")
	// Create an errgroup for background downloads with limited concurrency.
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(8)

	objects := d.backup.List(&blob.ListOptions{
		Delimiter: "/", // only list directories with a trailing slash and IsDir set to true
	})

	tblVersions := make(map[string]string)
	for {
		// Stop the loop if the ctx was cancelled
		var stop bool
		select {
		case <-ctx.Done():
			stop = true
		default:
			// don't break
		}
		if stop {
			break // can't use break inside the select
		}

		obj, err := objects.Next(ctx)
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
		d.logger.Debug("SyncWithObjectStorage: discovered table", slog.String("table", table))

		// get version of the table
		var backedUpVersion string
		err = retry(ctx, func() error {
			res, err := d.backup.ReadAll(ctx, path.Join(table, "version.txt"))
			if err != nil {
				return err
			}
			backedUpVersion = string(res)
			return nil
		})
		if err != nil {
			if gcerrors.Code(err) == gcerrors.NotFound {
				// invalid table directory
				d.logger.Debug("SyncWithObjectStorage: invalid table directory", slog.String("table", table))
				_ = d.deleteBackup(ctx, table, "")
			}
			return err
		}
		tblVersions[table] = backedUpVersion

		// check with current version
		version, exists, _ := tableVersion(d.localPath, table)
		if exists && version == backedUpVersion {
			d.logger.Debug("SyncWithObjectStorage: table is already up to date", slog.String("table", table))
			continue
		}

		tableDir := filepath.Join(d.localPath, table)
		// truncate existing table directory
		if err := os.RemoveAll(tableDir); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(tableDir, backedUpVersion), os.ModePerm); err != nil {
			return err
		}

		tblIter := d.backup.List(&blob.ListOptions{Prefix: path.Join(table, backedUpVersion)})
		// download all objects in the table and current version
		for {
			obj, err := tblIter.Next(ctx)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
			g.Go(func() error {
				return retry(ctx, func() error {
					file, err := os.Create(filepath.Join(d.localPath, obj.Key))
					if err != nil {
						return err
					}
					defer file.Close()

					rdr, err := d.backup.NewReader(ctx, obj.Key, nil)
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

	// Update table versions
	for table, version := range tblVersions {
		err = d.setTableVersion(table, version)
		if err != nil {
			return err
		}
	}

	// remove any tables that are not in backup
	entries, err := os.ReadDir(d.localPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, ok := tblVersions[entry.Name()]; ok {
			continue
		}
		err = os.RemoveAll(filepath.Join(d.localPath, entry.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}

// syncBackupWithLocal syncs the backup location with the local path for given table.
// If oldVersion is specified, it is deleted after successful sync.
func (d *db) syncBackupWithLocal(ctx context.Context, table, oldVersion string) error {
	if d.backup == nil {
		return nil
	}
	d.logger.Debug("syncing table", slog.String("table", table))
	version, exist, err := tableVersion(d.localPath, table)
	if err != nil {
		return err
	}

	if !exist {
		return fmt.Errorf("table %q not found", table)
	}

	localPath := filepath.Join(d.localPath, table, version)
	entries, err := os.ReadDir(localPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		d.logger.Debug("replicating file", slog.String("file", entry.Name()), slog.String("path", localPath))
		// no directory should exist as of now
		if entry.IsDir() {
			d.logger.Debug("found directory in path which should not exist", slog.String("file", entry.Name()), slog.String("path", localPath))
			continue
		}

		wr, err := os.Open(filepath.Join(localPath, entry.Name()))
		if err != nil {
			return err
		}

		// upload to cloud storage
		err = retry(ctx, func() error {
			return d.backup.Upload(ctx, path.Join(table, version, entry.Name()), wr, &blob.WriterOptions{
				ContentType: "application/octet-stream",
			})
		})
		_ = wr.Close()
		if err != nil {
			return err
		}
	}

	// update version.txt
	// Ideally if this fails it leaves backup in inconsistent state but for now we will rely on retries
	// ignore context cancellation errors for version.txt updates
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = retry(context.Background(), func() error {
		return d.backup.WriteAll(ctxWithTimeout, path.Join(table, "version.txt"), []byte(version), nil)
	})
	if err != nil {
		d.logger.Error("failed to update version.txt in backup", slog.Any("error", err))
	}

	// success -- remove old version
	if oldVersion != "" {
		_ = d.deleteBackup(ctx, table, oldVersion)
	}
	return err
}

// deleteBackup deletes backup.
// If table is specified, only that table is deleted.
// If table and version is specified, only that version of the table is deleted.
func (d *db) deleteBackup(ctx context.Context, table, version string) error {
	if d.backup == nil {
		return nil
	}
	if table == "" && version != "" {
		return fmt.Errorf("table must be specified if version is specified")
	}
	var prefix string
	if table != "" {
		if version != "" {
			prefix = path.Join(table, version) + "/"
		} else {
			// deleting the entire table
			prefix = table + "/"
			// delete version.txt first
			// also ignore context cancellation errors since it can leave the backup in inconsistent state
			ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err := retry(context.Background(), func() error { return d.backup.Delete(ctxWithTimeout, "version.txt") })
			if err != nil && gcerrors.Code(err) != gcerrors.NotFound {
				d.logger.Error("failed to delete version.txt in backup", slog.Any("error", err))
				return err
			}
		}
	}
	// ignore errors since version.txt is already removed

	iter := d.backup.List(&blob.ListOptions{Prefix: prefix})
	for {
		obj, err := iter.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			d.logger.Debug("failed to list object", slog.Any("error", err))
		}
		err = retry(ctx, func() error { return d.backup.Delete(ctx, obj.Key) })
		if err != nil {
			d.logger.Debug("failed to delete object", slog.String("object", obj.Key), slog.Any("error", err))
		}
	}
	return nil
}

func retry(ctx context.Context, fn func() error) error {
	var err error
	for i := 0; i < _maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err() // return on context cancellation
		case <-time.After(_retryDelay):
		}
		err = fn()
		if err == nil {
			return nil // success
		}
		if !strings.Contains(err.Error(), "stream error: stream ID") {
			break // break and return error
		}
	}
	return err
}

const (
	_maxRetries = 5
	_retryDelay = 10 * time.Second
)
