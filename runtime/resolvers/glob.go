package resolvers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"github.com/rilldata/rill/runtime/pkg/typepb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

func init() {
	runtime.RegisterResolverInitializer("glob", newGlob)
}

// globPartitionType is the type of partitioning for a "glob" resolver.
type globPartitionType string

const (
	// globPartitionTypeUnspecified is the empty partition type.
	globPartitionTypeUnspecified globPartitionType = ""

	// globPartitionTypeFile disables partition analysis, effectively treating each file as a separate partition.
	globPartitionTypeFile globPartitionType = "file"

	// globPartitionTypeDirectory treats each directory that directly contains files as a partition.
	globPartitionTypeDirectory globPartitionType = "directory"

	// globPartitionTypeHive extracts partitions from the path using Hive-style partitioning.
	globPartitionTypeHive globPartitionType = "hive"
)

// globResolver is a resolver that lists objects matching a glob pattern in an object store.
type globResolver struct {
	runtime      *runtime.Runtime
	instanceID   string
	props        *globProps
	bucketURI    *globutil.URL
	tmpTableName string
}

// globProps declares the properties for a "glob" resolver.
type globProps struct {
	// Connector is the object store connector to target.
	Connector string `mapstructure:"connector"`
	// Path is the glob pattern to match.
	Path string `mapstructure:"path"`
	// Partition defines if and how to group the files that match the glob into partitions.
	Partition globPartitionType `mapstructure:"partition"`
	// RollupFiles is a flag to roll up and include the files in each partition in the output.
	RollupFiles bool `mapstructure:"rollup_files"`
	// TransformSQL is an optional SQL statement to transform the results.
	// The SQL statement should be a DuckDB SQL statement that queries a table templated into the query with "{{ .table }}".
	TransformSQL string `mapstructure:"transform_sql"`
}

// globArgs declares the arguments for a "glob" resolver.
type globArgs struct {
	// State to make available for template resolution in the props.
	State map[string]any `mapstructure:"state"`
}

func newGlob(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	args := &globArgs{}
	if err := mapstructureutil.WeakDecode(opts.Args, args); err != nil {
		return nil, err
	}
	if args.State == nil {
		args.State = map[string]any{}
	}

	tmpTableName, err := randomString("glob", 8)
	if err != nil {
		return nil, fmt.Errorf("glob resolver: failed to generate random name: %w", err)
	}

	inst, err := opts.Runtime.Instance(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	propsMap, err := parser.ResolveTemplateRecursively(opts.Properties, parser.TemplateData{
		Environment: inst.Environment,
		User:        map[string]any{},
		Variables:   inst.ResolveVariables(false),
		State:       args.State,
		ExtraProps: map[string]any{
			"table": tmpTableName,
		},
	}, true)
	if err != nil {
		return nil, fmt.Errorf("glob resolver: failed to resolve templating: %w", err)
	}

	props := &globProps{}
	if err := mapstructureutil.WeakDecode(propsMap, props); err != nil {
		return nil, err
	}

	props.Path = strings.TrimSpace(props.Path)

	// set props to span attributes
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		span.SetAttributes(
			attribute.String("connector", props.Connector),
			attribute.String("path", props.Path),
			attribute.String("partition", string(props.Partition)),
			attribute.Bool("rollup_files", props.RollupFiles),
			attribute.String("transform_sql", props.TransformSQL),
		)
	}

	// Parse the bucket URI without the path (e.g. for "s3://bucket/path", it is "s3://bucket")
	bucketURI, err := globutil.ParseBucketURL(props.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bucket path %q: %w", props.Path, err)
	}
	bucketURI.Path = ""

	// If connector is not specified outright, infer it from the path (e.g. for "s3://bucket/path", the connector becomes "s3").
	if props.Connector == "" {
		if bucketURI.Scheme == "gs" {
			props.Connector = "gcs"
		} else {
			props.Connector = bucketURI.Scheme
		}
	}

	return &globResolver{
		runtime:      opts.Runtime,
		instanceID:   opts.InstanceID,
		props:        props,
		bucketURI:    bucketURI,
		tmpTableName: tmpTableName,
	}, nil
}

func (r *globResolver) Close() error {
	return nil
}

func (r *globResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	return nil, false, nil
}

func (r *globResolver) Refs() []*runtimev1.ResourceName {
	return nil
}

func (r *globResolver) Validate(ctx context.Context) error {
	return nil
}

func (r *globResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	h, release, err := r.runtime.AcquireHandle(ctx, r.instanceID, r.props.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	store, ok := h.AsObjectStore()
	if !ok {
		return nil, fmt.Errorf("connector %q is not an object store", r.props.Connector)
	}

	entries, err := store.ListObjects(ctx, r.props.Path)
	if err != nil {
		return nil, err
	}

	var rows []map[string]any
	switch r.props.Partition {
	case globPartitionTypeUnspecified, globPartitionTypeFile:
		rows = r.buildFilesResult(entries)
	case globPartitionTypeDirectory:
		rows = r.buildPartitionedResult(entries, false)
	case globPartitionTypeHive:
		rows = r.buildPartitionedResult(entries, true)
	default:
		return nil, fmt.Errorf("unknown glob partition type %q", r.props.Partition)
	}

	if r.props.TransformSQL != "" {
		rows, err = r.transformResult(ctx, rows, r.props.TransformSQL)
		if err != nil {
			return nil, err
		}
	}

	return runtime.NewMapsResolverResult(rows, r.makeResultSchema(rows)), nil
}

func (r *globResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return errors.New("not implemented")
}

func (r *globResolver) InferRequiredSecurityRules() ([]*runtimev1.SecurityRule, error) {
	return nil, errors.New("security rule inference not implemented")
}

// buildUnpartitioned builds a result consisting of one row per file.
// Each row is a map with the keys "uri", "path", and "updated_on".
func (r *globResolver) buildFilesResult(entries []drivers.ObjectStoreEntry) []map[string]any {
	rows := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir {
			continue
		}

		uri := &globutil.URL{
			Scheme: r.bucketURI.Scheme,
			Host:   r.bucketURI.Host,
			Path:   entry.Path,
		}

		rows = append(rows, map[string]any{
			"uri":        uri.String(),
			"path":       entry.Path,
			"updated_on": entry.UpdatedOn,
		})
	}
	return rows
}

// hivePartitionRegex is a regex that matches Hive-style partition values in a path.
var hivePartitionRegex = regexp.MustCompile(`/([^/\?]+)=([^/\n\?]*)`)

// buildPartitioned builds a result consisting of one row per partition.
// It groups the files by directory.
// Each row is a map with the keys "uri", "path", "updated_on", "files" if RollupFiles is true, and the Hive partition columns if requested.
func (r *globResolver) buildPartitionedResult(entries []drivers.ObjectStoreEntry, parseHivePartitions bool) []map[string]any {
	// Group the entries by directory
	rows := make(map[string]map[string]any)
	for _, entry := range entries {
		if entry.IsDir {
			continue
		}

		dir := path.Dir(entry.Path)

		row := rows[dir]
		if row == nil {
			// Init a new row
			uri := &globutil.URL{
				Scheme: r.bucketURI.Scheme,
				Host:   r.bucketURI.Host,
				Path:   dir,
			}
			row = map[string]any{
				"uri":        uri.String(),
				"path":       dir,
				"updated_on": entry.UpdatedOn,
			}
			if r.props.RollupFiles {
				row["files"] = []string{entry.Path}
			}
			rows[dir] = row

			// Extract and add Hive partition values
			if parseHivePartitions {
				for _, match := range hivePartitionRegex.FindAllStringSubmatch(dir, -1) {
					row[match[1]] = match[2]
				}
			}
		} else {
			// Add to files slice of the existing row
			if r.props.RollupFiles {
				files := row["files"].([]string)
				row["files"] = append(files, entry.Path)
			}

			// Bump updated_on of the existing row if it's newer
			updatedOn := row["updated_on"].(time.Time)
			if entry.UpdatedOn.After(updatedOn) {
				row["updated_on"] = entry.UpdatedOn
			}
		}
	}

	// Convert the map to a slice and sort ascending by path
	result := maps.Values(rows)
	slices.SortFunc(result, func(a, b map[string]any) int {
		return strings.Compare(a["path"].(string), b["path"].(string))
	})

	return result
}

func (r *globResolver) makeResultSchema(rows []map[string]any) *runtimev1.StructType {
	if len(rows) == 0 {
		return &runtimev1.StructType{}
	}
	row := rows[0]
	return typepb.InferFromValue(row).StructType
}

func (r *globResolver) transformResult(ctx context.Context, rows []map[string]any, sql string) ([]map[string]any, error) {
	olap, release, err := r.runtime.OLAP(ctx, r.instanceID, "duckdb")
	if err != nil {
		return nil, fmt.Errorf("failed to acquire duckdb for glob transform: %w", err)
	}
	defer release()

	jsonFile, err := r.writeTempNDJSONFile(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to write temp NDJSON file for glob transform: %w", err)
	}
	defer os.Remove(jsonFile)

	var result []map[string]any
	err = olap.WithConnection(ctx, 0, func(wrappedCtx context.Context, ensuredCtx context.Context) error {
		// Load the JSON file into a temporary table
		err = olap.Exec(wrappedCtx, &drivers.Statement{
			Query: fmt.Sprintf("CREATE TEMPORARY TABLE %s AS (SELECT * FROM read_ndjson_auto(%s))", olap.Dialect().EscapeIdentifier(r.tmpTableName), olap.Dialect().EscapeStringValue(jsonFile)),
		})
		if err != nil {
			return fmt.Errorf("failed to stage underlying data for pivot: %w", err)
		}

		// Defer cleanup of the temporary table
		defer func() {
			err = olap.Exec(ensuredCtx, &drivers.Statement{
				Query: fmt.Sprintf("DROP TABLE %s", olap.Dialect().EscapeIdentifier(r.tmpTableName)),
			})
			if err != nil {
				l, err2 := r.runtime.InstanceLogger(ctx, r.instanceID)
				if err2 == nil {
					l.Error("duckdb: failed to cleanup temporary table for glob transform", zap.Error(err))
				}
			}
		}()

		// Execute the transform SQL
		rows, err := olap.Query(wrappedCtx, &drivers.Statement{
			Query: sql,
		})
		if err != nil {
			return fmt.Errorf("failed to execute transform SQL for glob: %w", err)
		}
		for rows.Next() {
			row := make(map[string]any)
			if err := rows.MapScan(row); err != nil {
				return err
			}
			result = append(result, row)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *globResolver) writeTempNDJSONFile(rows []map[string]any) (string, error) {
	tempDir, err := r.runtime.TempDir(r.instanceID)
	if err != nil {
		return "", err
	}
	f, err := os.CreateTemp(tempDir, "glob_result_*.ndjson")
	if err != nil {
		return "", err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, row := range rows {
		if err := enc.Encode(row); err != nil {
			return "", err
		}
	}

	return f.Name(), nil
}

func randomString(prefix string, n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b), nil
}
