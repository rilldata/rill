package project

import (
	"fmt"
	"strconv"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/spf13/cobra"
)

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

			ch.Printf("  DATABASE NAME (%d)                       ROW COUNT      COLUMN COUNT   DATABASE SIZE\n", len(res.Tables))
			ch.Printf(" --------------------------------------- -------------- -------------- ---------------\n")

			for _, table := range res.Tables {
				var dbSize string
				if table.PhysicalSizeBytes == -1 {
					dbSize = "unknown"
				} else {
					dbSize = formatBytes(table.PhysicalSizeBytes)
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
					columnCount = strconv.Itoa(len(tableRes.Schema.Fields))
				} else {
					columnCount = "0"
				}

				// Get row count using SQL query
				var rowCount string
				queryRes, err := rt.QueryServiceClient().Query(cmd.Context(), &runtimev1.QueryRequest{
					InstanceId: instanceID,
					Sql:        fmt.Sprintf("SELECT COUNT(*) FROM %s", drivers.DialectDuckDB.EscapeIdentifier(table.Name)),
				})
				if err != nil {
					rowCount = "error"
				} else if len(queryRes.Data) > 0 && len(queryRes.Data[0].Fields) > 0 {
					// Extract the count value from the first row, first column (should be only one column from COUNT(*))
					for _, countValue := range queryRes.Data[0].Fields {
						if countValue != nil {
							if countValueNumber := countValue.GetNumberValue(); countValueNumber != 0 {
								rowCount = strconv.FormatInt(int64(countValueNumber), 10)
							} else {
								rowCount = "0"
							}
						} else {
							rowCount = "unknown"
						}
						break
					}
				} else {
					rowCount = "unknown"
				}

				ch.Printf("  %-39s %-14s %-14s %s\n", table.Name, rowCount, columnCount, dbSize)
			}

			return nil
		},
	}

	tablesCmd.Flags().StringVar(&project, "project", "", "Project name")
	tablesCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	tablesCmd.Flags().BoolVar(&local, "local", false, "Target local runtime instead of Rill Cloud")

	return tablesCmd
}

// formatBytes converts bytes to human readable format
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	if bytes < KB {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < MB {
		return fmt.Sprintf("%.1f KiB", float64(bytes)/KB)
	} else if bytes < GB {
		return fmt.Sprintf("%.1f MiB", float64(bytes)/MB)
	} else if bytes < TB {
		return fmt.Sprintf("%.1f GiB", float64(bytes)/GB)
	}
	return fmt.Sprintf("%.1f TiB", float64(bytes)/TB)
}
