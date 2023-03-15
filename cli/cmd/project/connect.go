package project

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/variable"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

const (
	pollTimeout  = 10 * time.Minute
	pollInterval = 5 * time.Second
)

func ConnectCmd(cfg *config.Config) *cobra.Command {
	var name, description, prodBranch, projectPath string
	var slots int
	var public bool
	var variables []string

	connectCmd := &cobra.Command{
		Use:   "connect",
		Short: "Setup continuous deployment to Rill Cloud",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Allow setting project path as arg (instead of flag)
			if len(args) > 0 {
				projectPath = args[0]
			}

			// Extract Git remote
			remotes, err := gitutil.ExtractRemotes(projectPath)
			if err != nil {
				if !errors.Is(err, git.ErrRepositoryNotExists) {
					return fmt.Errorf("failed to parse .git remotes: %w", err)
				}
				// Fall through to len(remotes) check
			}

			// Print setup instructions if no remote was found
			if len(remotes) == 0 {
				fmt.Print(githubSetupMsg)
				os.Exit(1)
			}

			// Parse into a https://github.com/account/repo (no .git) format
			githubURL, err := remotesToGithubURL(remotes)
			if err != nil {
				return err
			}

			// Create admin client
			client, err := client.New(cfg.AdminURL, cfg.AdminToken())
			if err != nil {
				return err
			}
			defer client.Close()

			// Check for access to the Github URL
			ghRes, err := client.GetGithubRepoStatus(cmd.Context(), &adminv1.GetGithubRepoStatusRequest{
				GithubUrl: githubURL,
			})
			if err != nil {
				return err
			}

			// If the user has not already granted access, open browser and poll for access
			if !ghRes.HasAccess {
				// Print instructions to grant access
				fmt.Printf("Rill projects deploy continuously when you push changes to Github.\n\n")
				fmt.Printf("Open this URL in your browser to grant Rill access to your Github repository:\n\n")
				fmt.Printf("\t%s\n\n", ghRes.GrantAccessUrl)

				// Open browser if possible
				_ = browser.Open(ghRes.GrantAccessUrl)

				// Poll for permission granted
				pollCtx, cancel := context.WithTimeout(cmd.Context(), pollTimeout)
				defer cancel()
				for {
					select {
					case <-pollCtx.Done():
						return pollCtx.Err()
					case <-time.After(pollInterval):
						// Ready to check again.
					}

					// Poll for access to the Github URL
					pollRes, err := client.GetGithubRepoStatus(cmd.Context(), &adminv1.GetGithubRepoStatusRequest{
						GithubUrl: githubURL,
					})
					if err != nil {
						return err
					}

					if pollRes.HasAccess {
						// Success
						ghRes = pollRes
						break
					}

					// Sleep and poll again
				}
			}

			// We now have access to the Github repo

			// Infer project name from Github remote (if not explicitly set)
			if name == "" {
				name = path.Base(githubURL)
			}

			// Use Github project's default branch (if not explicitly set)
			if prodBranch == "" {
				prodBranch = ghRes.DefaultBranch
			}

			parsedVariables, err := variable.Parse(variables)
			if err != nil {
				return err
			}

			// Create the project (automatically deploys prod branch)
			projRes, err := client.CreateProject(cmd.Context(), &adminv1.CreateProjectRequest{
				OrganizationName: cfg.Org(),
				Name:             name,
				Description:      description,
				ProductionSlots:  int64(slots),
				ProductionBranch: prodBranch,
				Public:           public,
				GithubUrl:        githubURL,
				Variables:        parsedVariables,
			})
			if err != nil {
				return err
			}

			// Success!
			fmt.Printf("Created project %s/%s\n", cfg.Org(), projRes.Project.Name)
			return nil
		},
	}

	connectCmd.Flags().SortFlags = false
	connectCmd.Flags().StringVar(&projectPath, "project", ".", "Project directory")
	connectCmd.Flags().StringVar(&prodBranch, "prod-branch", "", "Git branch to deploy from (default: the default Git branch)")
	connectCmd.Flags().IntVar(&slots, "prod-slots", 2, "Slots to allocate for production deployments")
	connectCmd.Flags().StringVar(&name, "name", "", "Project name (default: the Github repo name)")
	connectCmd.Flags().StringVar(&description, "description", "", "Project description")
	connectCmd.Flags().BoolVar(&public, "public", false, "Make dashboards publicly accessible")
	connectCmd.Flags().StringSliceVarP(&variables, "env", "e", []string{}, "Set project variables")

	return connectCmd
}

func remotesToGithubURL(remotes []gitutil.Remote) (string, error) {
	// Return the first Github URL found.
	// If no Github remotes were found, return the first error.
	var firstErr error
	for _, remote := range remotes {
		ghurl, err := remoteToGithubURL(remote.URL)
		if err == nil {
			// Found a Github remote. Success!
			return ghurl, nil
		}
		if firstErr == nil {
			firstErr = fmt.Errorf("invalid remote %q: %w", remote.URL, err)
		}
	}

	// This condition is handled upstream to print better instructions. Adding here only for safety.
	if firstErr == nil {
		return "", fmt.Errorf("no git remotes found")
	}

	return "", firstErr
}

func remoteToGithubURL(remote string) (string, error) {
	ep, err := transport.NewEndpoint(remote)
	if err != nil {
		return "", err
	}

	if ep.Host != "github.com" {
		return "", fmt.Errorf("must be a git remote on github.com")
	}

	account, repo := path.Split(ep.Path)
	account = strings.Trim(account, "/")
	repo = strings.TrimSuffix(repo, ".git")
	if account == "" || repo == "" || strings.Contains(account, "/") {
		return "", fmt.Errorf("not a valid github.com remote")
	}

	githubURL := &url.URL{
		Scheme: "https",
		Host:   ep.Host,
		Path:   strings.TrimSuffix(ep.Path, ".git"),
	}

	return githubURL.String(), nil
}

const githubSetupMsg = `No git remote was found.

Rill projects deploy continuously when you push changes to Github.
Therefore, your project must be on Github before you connect it to Rill.

Follow these steps to push your project to Github.

1. Initialize git

	git init

2. Add and commit files

	git add .
	git commit -m 'initial commit'

3. Create a new GitHub repository on https://github.com/new

4. Link git to the remote repository

	git remote add origin https://github.com/your-account/your-repo.git

5. Push your repository

	git push -u origin main

6. Connect Rill to your repository

	rill project connect

`
