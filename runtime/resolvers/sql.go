package resolvers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	compilerv1 "github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/reconcilers"
	"google.golang.org/protobuf/types/known/structpb"
)

func init() {
	runtime.RegisterAPIResolverInitializer("SQL", New)
}

type SQLResolver struct {
	resolvedSQL string
	deps        []*runtimev1.ResourceName
	olap        drivers.OLAPStore
	releaseFunc func()
}

func New(ctx context.Context, opts *runtime.APIResolverOptions) (runtime.APIResolver, error) {
	sql := opts.API.Spec.ResolverProperties.Fields["sql"].GetStringValue()
	if sql == "" {
		return nil, errors.New("no sql query found for sql resolver")
	}
	resolvedSQL, deps, err := resolveSQLAndDeps(ctx, sql, opts)
	if err != nil {
		return nil, err
	}
	olap, release, err := opts.Runtime.OLAP(ctx, opts.InstanceID)
	if err != nil {
		return nil, err
	}

	return &SQLResolver{
		resolvedSQL: resolvedSQL,
		deps:        deps,
		olap:        olap,
		releaseFunc: release,
	}, nil
}

// Key that can be used for caching
func (r *SQLResolver) Key() string {
	return r.resolvedSQL
}

// Deps referenced by the query
func (r *SQLResolver) Deps() []*runtimev1.ResourceName {
	return r.deps
}

// Validate the query without running any "expensive" operations
func (r *SQLResolver) Validate(ctx context.Context) error {
	_, err := r.olap.Execute(ctx, &drivers.Statement{
		Query:  r.resolvedSQL,
		DryRun: true,
	})
	return err
}

// ResolveInteractive Resolve for interactive use (e.g. API requests or alerts)
func (r *SQLResolver) ResolveInteractive(ctx context.Context, priority int) (runtime.Result, error) {
	res, err := r.olap.Execute(ctx, &drivers.Statement{
		Query:    r.resolvedSQL,
		Priority: priority,
	})
	if err != nil {
		return runtime.Result{}, err
	}
	var out []map[string]interface{}
	for res.Rows.Next() {
		row := make(map[string]interface{})
		err = res.Rows.MapScan(row)
		if err != nil {
			return runtime.Result{}, err
		}
		out = append(out, row)
	}

	b, err := json.Marshal(out)
	if err != nil {
		return runtime.Result{}, err
	}

	return runtime.Result{
		Rows:  b,
		Cache: r.olap.Dialect() == drivers.DialectDuckDB,
	}, nil
}

// ResolveExport Resolve for export to a file (e.g. downloads or reports)
func (r *SQLResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ExportOptions) error {
	switch r.olap.Dialect() {
	case drivers.DialectDuckDB:
		if opts.Format == runtimev1.ExportFormat_EXPORT_FORMAT_CSV || opts.Format == runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET {
			filename := r.generateFilename()
			if err := queries.DuckDBCopyExport(ctx, w, opts, r.resolvedSQL, nil, filename, r.olap, opts.Format); err != nil {
				return err
			}
		} else {
			if err := r.generalExport(ctx, w, opts); err != nil {
				return err
			}
		}
	case drivers.DialectDruid:
		if err := r.generalExport(ctx, w, opts); err != nil {
			return err
		}
	case drivers.DialectClickHouse:
		if err := r.generalExport(ctx, w, opts); err != nil {
			return err
		}
	default:
		return fmt.Errorf("not available for dialect '%s'", r.olap.Dialect())
	}

	return nil
}

func (r *SQLResolver) generalExport(ctx context.Context, w io.Writer, opts *runtime.ExportOptions) error {
	res, err := r.olap.Execute(ctx, &drivers.Statement{
		Query:    r.resolvedSQL,
		Priority: opts.Priority,
	})
	if err != nil {
		return err
	}

	meta := make([]*runtimev1.MetricsViewColumn, len(res.Schema.Fields))
	for i, f := range res.Schema.Fields {
		meta[i] = &runtimev1.MetricsViewColumn{
			Name: f.Name,
			Type: f.Type.Code.String(),
		}
	}

	var data []*structpb.Struct
	for res.Rows.Next() {
		row := make(map[string]interface{})
		err = res.Rows.MapScan(row)
		if err != nil {
			return err
		}
		curr, err := structpb.NewStruct(row)
		if err != nil {
			return err
		}
		data = append(data, curr)
	}

	if opts.PreWriteHook != nil {
		err = opts.PreWriteHook(r.generateFilename())
		if err != nil {
			return err
		}
	}

	switch opts.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED:
		return fmt.Errorf("unspecified format")
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		return queries.WriteCSV(meta, data, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		return queries.WriteXLSX(meta, data, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		return queries.WriteParquet(meta, data, w)
	}

	return nil
}

func (r *SQLResolver) Close() error {
	r.releaseFunc()
	return nil
}

func (r *SQLResolver) generateFilename() string {
	return "api_export_" + time.Now().Format("2006-01-02T15-04-05.000Z")
}

func resolveSQLAndDeps(ctx context.Context, sqlTemplate string, opts *runtime.APIResolverOptions) (string, []*runtimev1.ResourceName, error) {
	dialect, err := getDialect(ctx, opts.Runtime, opts.InstanceID)
	if err != nil {
		return "", nil, err
	}
	var deps []*runtimev1.ResourceName

	sql, err := compilerv1.ResolveTemplate(sqlTemplate, compilerv1.TemplateData{
		User:       opts.UserAttributes,
		ExtraProps: opts.Args,
		Self: compilerv1.TemplateResource{
			Meta:  &runtimev1.ResourceMeta{}, // TODO: Fill in with actual metadata
			Spec:  opts.API.Spec,
			State: opts.API.State,
		},
		Resolve: func(ref compilerv1.ResourceName) (string, error) {
			return reconcilers.SafeSQLName(ref.Name), nil
		},
		Lookup: func(name compilerv1.ResourceName) (compilerv1.TemplateResource, error) {
			// TODO: Implement this, do we need to support this?
			return compilerv1.TemplateResource{}, nil
		},
	})
	if err != nil {
		return "", nil, err
	}

	if dialect == drivers.DialectDuckDB {
		ast, err := duckdbsql.Parse(sql)
		if err != nil {
			return "", nil, err
		}
		for _, t := range ast.GetTableRefs() {
			if !t.LocalAlias && t.Name != "" && t.Function == "" && len(t.Paths) == 0 {
				deps = append(deps, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: t.Name})
			}
		}
	} else {
		meta, err := compilerv1.AnalyzeTemplate(sqlTemplate)
		if err != nil {
			return "", nil, err
		}
		for _, ref := range meta.Refs {
			deps = append(deps, &runtimev1.ResourceName{Kind: ref.Kind.String(), Name: ref.Name})
		}
	}

	return sql, deps, nil
}

func getDialect(ctx context.Context, r *runtime.Runtime, instanceID string) (drivers.Dialect, error) {
	i, err := r.Instance(ctx, instanceID)
	if err != nil {
		return drivers.DialectUnspecified, err
	}
	dialect := connectorToDialect(i.ResolveOLAPConnector())
	return dialect, nil
}

func connectorToDialect(connector string) drivers.Dialect {
	switch connector {
	case "duckdb":
		return drivers.DialectDuckDB
	case "druid":
		return drivers.DialectDruid
	case "clickhouse":
		return drivers.DialectClickHouse
	default:
		return drivers.DialectUnspecified
	}
}
