package server

import (
	"context"
	"errors"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"github.com/rilldata/rill/runtime/server/auth"
)

func (s *Server) ProjectStorage(ctx context.Context, req *runtimev1.ProjectStorageRequest) (*runtimev1.ProjectStorageResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	// Require admin permissions (currently indicatsed by ReadInstance)
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadInstance) {
		return nil, ErrForbidden
	}

	res, _, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID: req.InstanceId,
		Resolver:   "project_storage",
		Claims:     claims,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	type row struct {
		Connector     string `mapstructure:"connector"`
		Driver        string `mapstructure:"driver"`
		IsDefaultOLAP bool   `mapstructure:"is_default_olap"`
		Managed       bool   `mapstructure:"managed"`
		SizeBytes     int64  `mapstructure:"size_bytes"`
		Error         string `mapstructure:"error"`
	}

	var entries []*runtimev1.ProjectStorageEntry
	managedSizeBytes := int64(-1)
	defaultOLAPSizeBytes := int64(-1)
	for {
		m, err := res.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		var r row
		if err := mapstructureutil.WeakDecode(m, &r); err != nil {
			return nil, err
		}

		entries = append(entries, &runtimev1.ProjectStorageEntry{
			Connector:     r.Connector,
			Driver:        r.Driver,
			IsDefaultOlap: r.IsDefaultOLAP,
			Managed:       r.Managed,
			SizeBytes:     r.SizeBytes,
			Error:         r.Error,
		})

		if r.Managed && r.SizeBytes > 0 {
			if managedSizeBytes < 0 {
				managedSizeBytes = r.SizeBytes
			} else {
				managedSizeBytes += r.SizeBytes
			}
		}
		if r.IsDefaultOLAP {
			defaultOLAPSizeBytes = r.SizeBytes
		}
	}

	return &runtimev1.ProjectStorageResponse{
		Entries:              entries,
		ManagedSizeBytes:     managedSizeBytes,
		DefaultOlapSizeBytes: defaultOLAPSizeBytes,
	}, nil
}
