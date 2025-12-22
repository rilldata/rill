package project

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rilldata/rill/cli/cmd/env"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func CloneCmd(ch *cmdutil.Helper) *cobra.Command {
	var path string

	cloneCmd := &cobra.Command{
		Use:   "clone <project-name>",
		Args:  cobra.ExactArgs(1),
		Short: "Clone Project",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if path == "" {
				path = name
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			// get project
			res, err := client.GetProject(cmd.Context(), &adminv1.GetProjectRequest{Org: ch.Org, Project: name})
			if err != nil {
				return err
			}

			if res.Project.ArchiveAssetId != "" {
				return fmt.Errorf("project is not connected to a git repository, please redeploy the project to connect it to a git repository")
			}

			// check if dir is empty
			empty, err := isDirAbsentOrEmpty(path)
			if err != nil {
				return err
			}
			if !empty {
				ok, err := cmdutil.ConfirmPrompt(fmt.Sprintf("There are files at path %q. Do you want to overwrite?", path), "", false)
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("directory %q is not empty", path)
				}
			}

			// create directory, the go-git SDK does not create the directory
			err = recreateDir(path)
			if err != nil {
				return fmt.Errorf("failed to create directory %q: %w", path, err)
			}

			// get config
			config, err := ch.GitHelper(ch.Org, name, path).GitConfig(cmd.Context())
			if err != nil {
				return err
			}

			// clone repository
			_, err = gitutil.Clone(cmd.Context(), path, config)
			if err != nil {
				return err
			}

			var subpath string
			if res.Project.Subpath != "" {
				subpath = filepath.Join(path, res.Project.Subpath)
			} else {
				subpath = path
			}

			ch.Printf("Cloned project %q to %q\n", name, subpath)

			// download variables
			err = env.PullVars(cmd.Context(), ch, subpath, name, "prod", false)
			if err != nil {
				return fmt.Errorf("failed to download variables: %w", err)
			}

			return nil
		},
	}

	cloneCmd.Flags().SortFlags = false
	cloneCmd.Flags().StringVar(&path, "path", "", "Project path to clone to")

	return cloneCmd
}

func isDirAbsentOrEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	defer f.Close()

	// Read the directory entries
	entries, err := f.Readdirnames(1)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return len(entries) == 0, nil
		}
		return false, err
	}

	// Check if the directory is empty
	return len(entries) == 0, nil
}

func recreateDir(path string) error {
	// Remove directory and its contents if exists
	err := os.RemoveAll(path)
	if err != nil {
		// NOTE: does not return an error if the directory does not exist
		return fmt.Errorf("failed to remove dir: %w", err)
	}

	// Create the directory again
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create dir: %w", err)
	}

	return nil
}
