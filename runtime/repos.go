package runtime

import (
	"context"
	"io"
	"time"
)

func (r *Runtime) ListFiles(ctx context.Context, instanceID string, glob string) ([]string, error) {
	repo, err := r.Repo(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	return repo.ListRecursive(ctx, instanceID, glob)
}

func (r *Runtime) GetFile(ctx context.Context, instanceID string, path string) (string, time.Time, error) {
	repo, err := r.Repo(ctx, instanceID)
	if err != nil {
		return "", time.Time{}, err
	}

	blob, err := repo.Get(ctx, instanceID, path)
	if err != nil {
		return "", time.Time{}, err
	}

	// TODO: Could we return Stat as part of Get?
	stat, err := repo.Stat(ctx, instanceID, path)
	if err != nil {
		return "", time.Time{}, err
	}

	return blob, stat.LastUpdated, nil
}

func (r *Runtime) PutFile(ctx context.Context, instanceID string, path string, blob string, create bool, createOnly bool) error {
	repo, err := r.Repo(ctx, instanceID)
	if err != nil {
		return err
	}

	// TODO: Handle create, createOnly

	err = repo.PutBlob(ctx, instanceID, path, blob)
	if err != nil {
		return err
	}

	return nil
}

// TODO: Merge with PutFile. A string can easily be converted to a reader.
func (r *Runtime) PutFileReader(ctx context.Context, instanceID string, path string, blob io.Reader, create bool, createOnly bool) error {
	repo, err := r.Repo(ctx, instanceID)
	if err != nil {
		return err
	}

	// TODO: Handle create, createOnly

	err = repo.PutReader(ctx, instanceID, path, blob)
	if err != nil {
		return err
	}

	return nil
}

func (r *Runtime) DeleteFile(ctx context.Context, instanceID string, path string) error {
	repo, err := r.Repo(ctx, instanceID)
	if err != nil {
		return err
	}

	err = repo.Delete(ctx, instanceID, path)
	if err != nil {
		return err
	}

	return nil
}

func (r *Runtime) RenameFile(ctx context.Context, instanceID string, fromPath string, toPath string) error {
	repo, err := r.Repo(ctx, instanceID)
	if err != nil {
		return err
	}

	err = repo.Rename(ctx, instanceID, fromPath, toPath)
	if err != nil {
		return err
	}

	return nil
}
