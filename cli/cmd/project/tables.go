package project

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

// TableInfo represents information about a database table
type TableInfo struct {
	Name         string `json:"name" csv:"NAME" header:"NAME"`
	RowCount     string `json:"row_count" csv:"ROW_COUNT" header:"ROW COUNT"`
	ColumnCount  string `json:"column_count" csv:"COLUMN_COUNT" header:"COLUMN COUNT"`
	DatabaseSize string `json:"database_size" csv:"DATABASE_SIZE" header:"DATABASE SIZE"`
}

func TablesCmd(ch *cmdutil.Helper) *cobra.Command {
	var project, path string
	var local bool

	tablesCmd := &cobra.Command{
		Use:   "tables [<project-name>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Get information about tables in a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				project = args[0]
			}

			if !local && !cmd.Flags().Changed("project") && len(args) == 0 && ch.Interactive {
				var err error
				project, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
				}
			}

			rt, instanceID, err := ch.OpenRuntimeClient(cmd.Context(), ch.Org, project, local)
			if err != nil {
				return fmt.Errorf("failed to connect to runtime: %w", err)
			}
			defer rt.Close()

			res, err := rt.ConnectorServiceClient().OLAPListTables(cmd.Context(), &runtimev1.OLAPListTablesRequest{
				InstanceId: instanceID,
				Connector:  "", // Uses default OLAP connector
			})
			if err != nil {
				return fmt.Errorf("failed to list tables: %w", err)
			}

			if len(res.Tables) == 0 {
				ch.PrintfWarn("No tables found\n")
				return nil
			}

			var tableInfos []TableInfo

			for _, table := range res.Tables {
				var dbSize string
				if table.PhysicalSizeBytes == -1 {
					dbSize = "unknown"
				} else {
					dbSize = ch.FormatBytes(table.PhysicalSizeBytes)
				}

				// Get table information for column count
				var columnCount string
				tableRes, err := rt.ConnectorServiceClient().OLAPGetTable(cmd.Context(), &runtimev1.OLAPGetTableRequest{
					InstanceId: instanceID,
					Table:      table.Name,
					Connector:  "", // Uses default OLAP connector
				})
				if err != nil {
					columnCount = "error"
				} else if tableRes.Schema != nil {
					columnCount = ch.FormatNumber(int64(len(tableRes.Schema.Fields)))
				} else {
					columnCount = "0"
				}

				// Get row count using SQL query
				var rowCount string
				countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", drivers.DialectDuckDB.EscapeIdentifier(table.Name))
				queryRes, err := rt.RuntimeServiceClient.QueryResolver(cmd.Context(), &runtimev1.QueryResolverRequest{
					InstanceId:         instanceID,
					Resolver:           "sql",
					ResolverProperties: must(structpb.NewStruct(map[string]any{"sql": countQuery})),
				})
				if err != nil {
					rowCount = "error"
				} else if len(queryRes.Data) > 0 && len(queryRes.Data[0].Fields) > 0 {
					// Extract the count value from the first row, first column (should be only one column from COUNT(*))
					for _, countValue := range queryRes.Data[0].Fields {
						if countValue != nil {
							rowCount = fmt.Sprint(countValue.AsInterface())
						} else {
							rowCount = "unknown"
						}
						break
					}
				} else {
					rowCount = "unknown"
				}

				tableInfos = append(tableInfos, TableInfo{
					Name:         table.Name,
					RowCount:     rowCount,
					ColumnCount:  columnCount,
					DatabaseSize: dbSize,
				})
			}

			ch.PrintData(tableInfos)

			return nil
		},
	}

	tablesCmd.Flags().StringVar(&project, "project", "", "Project name")
	tablesCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	tablesCmd.Flags().BoolVar(&local, "local", false, "Target local runtime instead of Rill Cloud")

	return tablesCmd
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
