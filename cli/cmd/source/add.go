package source

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/rilldata/rill/cli/pkg/local"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/artifacts/artifactsv0"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

// addCmd represents the add command, it requires min 1 args as source name
func AddCmd(ver string) *cobra.Command {
	var olapDriver string
	var olapDSN string
	var repoDSN string
	var sourceName string
	var delimiter string
	var verbose bool

	var addCmd = &cobra.Command{
		Use:   "add <file>",
		Short: "Add a local file source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dataPath := args[0]
			if !filepath.IsAbs(dataPath) {
				relPath, err := filepath.Rel(repoDSN, dataPath)
				if err != nil {
					return err
				}
				dataPath = relPath
			}

			app, err := local.NewApp(ver, verbose, olapDriver, olapDSN, repoDSN)
			if err != nil {
				return err
			}

			if !app.IsProjectInit() {
				return fmt.Errorf("not a valid Rill project")
			}

			if sourceName == "" {
				sourceName = fileutil.Stem(dataPath)
			}

			props := map[string]any{"path": dataPath}
			if delimiter != "" {
				props["csv.delimiter"] = delimiter
			}

			propsPB, err := structpb.NewStruct(props)
			if err != nil {
				return fmt.Errorf("can't serialize artifact: %w", err)
			}

			src := &runtimev1.Source{
				Name:       sourceName,
				Connector:  "file",
				Properties: propsPB,
			}

			repo, err := app.Runtime.Repo(context.Background(), app.Instance.ID)
			if err != nil {
				panic(err) // Should never happen
			}

			c := artifactsv0.New(repo, app.Instance.ID)
			sourcePath, err := c.PutSource(context.Background(), repo, app.Instance.ID, src)
			if err != nil {
				return err
			}

			err = app.ReconcileSource(sourcePath)
			if err != nil {
				return err
			}

			return nil
		},
	}

	addCmd.Flags().StringVar(&olapDriver, "db-driver", local.DefaultOLAPDriver, "OLAP database driver")
	addCmd.Flags().StringVar(&olapDSN, "db", local.DefaultOLAPDSN, "OLAP database DSN")
	addCmd.Flags().StringVar(&repoDSN, "dir", ".", "Project directory")
	addCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")
	addCmd.Flags().StringVar(&sourceName, "name", "", "Source name")
	addCmd.Flags().StringVar(&delimiter, "delimiter", "", "CSV delimiter override (it will autodetect if not set)")

	return addCmd
}
