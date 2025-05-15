package project

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rilldata/rill/cli/cmd/env"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime/pkg/archive"
	"github.com/spf13/cobra"
)

func CloneCmd(ch *cmdutil.Helper) *cobra.Command {
	var path string

	cloneCmd := &cobra.Command{
		Use:   "clone [<project-name>]",
		Args:  cobra.ExactArgs(1),
		Short: "Clone Project",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}
			name := args[0]
			if path == "" {
				path = name
			}

			// check if dir is empty
			empty, err := isDirAbsentOrEmpty(path)
			if err != nil {
				return err
			}
			if !empty {
				ok, err := cmdutil.ConfirmPrompt(fmt.Sprintf("Directory %q is not empty. Do you want to overwrite it?", path), "", false)
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

			p, err := client.GetProject(cmd.Context(), &adminv1.GetProjectRequest{OrganizationName: ch.Org, Name: name})
			if err != nil {
				return err
			}

			// get creds
			creds, archiveURL, err := ch.GitCredentials(cmd.Context(), p.Project, path)
			if err != nil {
				if archiveURL == "" {
					return err
				}

				// it is based on archive URL
				dst, err := os.MkdirTemp("", "rill-archive")
				if err != nil {
					return err
				}
				defer os.RemoveAll(dst)
				if err := archive.Download(cmd.Context(), archiveURL, dst, path, false, true); err != nil {
					return err
				}
				return env.PullVars(cmd.Context(), ch, path, name, "prod", false)
				// Should it auto migrate to managed git as well ?
			}

			remote, err := creds.FullyQualifiedRemote()
			if err != nil {
				return err
			}

			// clone repository
			_, err = git.PlainCloneContext(cmd.Context(), path, false, &git.CloneOptions{
				URL:           remote,
				ReferenceName: plumbing.NewBranchReferenceName(p.Project.ProdBranch), // TODO:: may be store prod branch in .git as well ?
				SingleBranch:  true,
			})
			if err != nil {
				return err
			}

			// download variables
			return env.PullVars(cmd.Context(), ch, path, name, "prod", false)
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
		return false, err
	}

	// Check if the directory is empty
	return len(entries) == 0, nil
}

func recreateDir(path string) error {
	// Remove directory and its contents if exists
	err := os.RemoveAll(path)
	if err != nil {
		// NOTE :: does not return an error if the directory does not exist
		return fmt.Errorf("failed to remove dir: %w", err)
	}

	// Create the directory again
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create dir: %w", err)
	}

	return nil
}
