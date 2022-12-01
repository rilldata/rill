package initialize

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/cli/pkg/examples"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/repo/repov0"
	"github.com/spf13/cobra"
)

// InitCmd represents the init command
func InitCmd() *cobra.Command {
	var repoDSN string
	var exampleName string
	var listExamples bool

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Rill project",
		RunE: func(cmd *cobra.Command, args []string) error {
			// List examples and exit
			if listExamples {
				names, err := examples.List()
				if err != nil {
					return err
				}
				for _, name := range names {
					fmt.Println(name)
				}
				return nil
			}

			// Prepare
			isPwd := repoDSN == "."
			isExample := exampleName != "empty"
			repoDSN = filepath.Clean(repoDSN)
			name := filepath.Base(repoDSN)
			if name == "" || strings.ContainsAny(name, "./\\") {
				name = "untitled"
			}

			// Open the project as a repo
			// TODO: Init a runtime and go through its interface instead.
			conn, err := drivers.Open("file", repoDSN)
			if err != nil {
				return err
			}
			repo, ok := conn.RepoStore()
			if !ok {
				panic("file driver is not a repo") // impossible
			}
			instanceID := "" // hacky, but doesn't matter for file repos

			// Check if already initialized
			if repov0.IsInit(context.Background(), repo, instanceID) {
				if isPwd {
					return fmt.Errorf("a Rill project already exists in the current directory")
				} else {
					return fmt.Errorf("a Rill project already exists in directory '%s'", repoDSN)
				}
			}

			// Use repo parser's init for empty projects
			if !isExample {
				err := repov0.InitEmpty(context.Background(), repo, instanceID, name)
				if err != nil {
					return err
				}

				if isPwd {
					fmt.Printf("Initialized empty project in the current directory\n")
				} else {
					fmt.Printf("Initialized empty project in directory '%s'\n", repoDSN)
				}

				return nil
			}

			// It's an example project. We currently only support examples through direct file unpacking.
			// TODO: Support unpacking examples through repo parser, instead of unpacking files.

			err = examples.Init(exampleName, repoDSN)
			if err != nil {
				if err == examples.ErrExampleNotFound {
					return fmt.Errorf("example project '%s' not found", exampleName)
				}
				return fmt.Errorf("failed to initialize project (detailed error: %s)", err.Error())
			}

			if isPwd {
				fmt.Printf("Initialized example project '%s' in the current directory\n", exampleName)
			} else {
				fmt.Printf("Initialized example project '%s' in directory '%s'\n", exampleName, repoDSN)
			}

			return nil
		},
	}

	initCmd.Flags().StringVar(&repoDSN, "dir", ".", "Directory to initialize")
	initCmd.Flags().StringVar(&exampleName, "example", "empty", "Name of example project")
	initCmd.Flags().BoolVar(&listExamples, "list-examples", false, "List available example projects")

	return initCmd
}
