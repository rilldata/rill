package runtime

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
)

func (r *Runtime) ListFiles(ctx context.Context, instanceID, glob string) ([]drivers.DirEntry, error) {
	repo, release, err := r.Repo(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	defer release()

	return repo.ListGlob(ctx, glob, false)
}

func (r *Runtime) GetFile(ctx context.Context, instanceID, path string) (string, time.Time, error) {
	repo, release, err := r.Repo(ctx, instanceID)
	if err != nil {
		return "", time.Time{}, err
	}
	defer release()

	blob, err := repo.Get(ctx, path)
	if err != nil {
		return "", time.Time{}, err
	}

	// TODO: Could we return Stat as part of Get?
	stat, err := repo.Stat(ctx, path)
	if err != nil {
		return "", time.Time{}, err
	}

	return blob, stat.LastUpdated, nil
}

func (r *Runtime) PutFile(ctx context.Context, instanceID, path string, blob io.Reader, create, createOnly bool) error {
	repo, release, err := r.Repo(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	// TODO: Handle create, createOnly

	instCfg, err := r.InstanceConfig(ctx, instanceID)
	if err != nil {
		return err
	}
	if instCfg.AssumeModelsMaterialized || instCfg.DisableModels {
		b, err := io.ReadAll(blob)
		if err != nil {
			return err
		}
		kind, err := parser.FileResourceKind(path, b)
		if err != nil {
			return err
		}
		if kind == parser.ResourceKindModel {
			return fmt.Errorf("cannot create model file %q because models are disabled for this instance", path)
		}
		blob = bytes.NewBuffer(b) // bcz we already read it, we need to create a new reader for the next step
	}

	err = repo.Put(ctx, path, blob)
	if err != nil {
		return err
	}

	return nil
}

func (r *Runtime) MkdirAll(ctx context.Context, instanceID, path string) error {
	repo, release, err := r.Repo(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	err = repo.MkdirAll(ctx, path)
	if err != nil {
		return err
	}

	return nil
}

func (r *Runtime) DeleteFile(ctx context.Context, instanceID, path string, force bool) error {
	repo, release, err := r.Repo(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	instCfg, err := r.InstanceConfig(ctx, instanceID)
	if err != nil {
		return err
	}
	if instCfg.AssumeModelsMaterialized || instCfg.DisableModels {
		contents, err := repo.Get(ctx, path)
		if err != nil {
			return err
		}
		kind, err := parser.FileResourceKind(path, []byte(contents))
		if err != nil {
			return err
		}
		if kind == parser.ResourceKindModel {
			return fmt.Errorf("cannot delete model file %q because models are disabled for this instance", path)
		}
	}

	err = repo.Delete(ctx, path, force)
	if err != nil {
		return err
	}

	return nil
}

func (r *Runtime) RenameFile(ctx context.Context, instanceID, fromPath, toPath string) error {
	repo, release, err := r.Repo(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	instCfg, err := r.InstanceConfig(ctx, instanceID)
	if err != nil {
		return err
	}
	if instCfg.AssumeModelsMaterialized || instCfg.DisableModels {
		contents, err := repo.Get(ctx, fromPath)
		if err != nil {
			return err
		}
		kind, err := parser.FileResourceKind(fromPath, []byte(contents))
		if err != nil {
			return err
		}
		if kind == parser.ResourceKindModel {
			return fmt.Errorf("cannot rename model file %q because models are disabled for this instance", fromPath)
		}
	}

	err = repo.Rename(ctx, fromPath, toPath)
	if err != nil {
		return err
	}

	return nil
}
