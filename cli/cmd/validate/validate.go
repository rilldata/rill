package validate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rilldata/rill/cli/cmd/env"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/local"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/spf13/cobra"
)

const (
	defaultOLAPConnector = "duckdb"
	listResourcesTimeout = 10 * time.Second
)

// ValidateCmd validates and reconciles a project without starting the UI.
func ValidateCmd(ch *cmdutil.Helper) *cobra.Command {
	var verbose bool
	var debug bool
	var reset bool
	var logFormat string
	var envVars []string
	var environment string
	var reconcileTimeout time.Duration

	validateCmd := &cobra.Command{
		Use:   "validate [<path>]",
		Short: "Validate project resources",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath, err := resolveProjectPath(args)
			if err != nil {
				return err
			}

			if !local.IsProjectInit(projectPath) {
				return fmt.Errorf("no Rill project found at %q (missing rill.yaml)", projectPath)
			}

			if ch.IsAuthenticated() {
				err := env.PullVars(cmd.Context(), ch, projectPath, "", environment, false)
				if err != nil && !errors.Is(err, cmdutil.ErrNoMatchingProject) {
					ch.PrintfWarn("Warning: failed to pull environment credentials: %v\n", err)
				}
			}

			if err := enforceRepoLimits(cmd.Context(), projectPath, ch); err != nil {
				return err
			}

			parsedLogFormat, ok := local.ParseLogFormat(logFormat)
			if !ok {
				return fmt.Errorf("invalid log format %q", logFormat)
			}

			envVarsMap, err := parseVariables(envVars)
			if err != nil {
				return err
			}

			if err := parseProject(cmd.Context(), ch, projectPath, environment); err != nil {
				return err
			}

			ch.Interactive = false
			app, err := local.NewApp(cmd.Context(), &local.AppOptions{
				Ch:             ch,
				Verbose:        verbose,
				Debug:          debug,
				Reset:          reset,
				Environment:    environment,
				ProjectPath:    projectPath,
				LogFormat:      parsedLogFormat,
				Variables:      envVarsMap,
				LocalURL:       "",           // No UI, so no local URL
				AllowedOrigins: []string{""}, // No UI, so no allowed origins
				ServeUI:        false,
			})
			if err != nil {
				return err
			}
			defer app.Close()

			return reconcileAndReport(cmd.Context(), ch, app, reconcileTimeout)
		},
	}

	validateCmd.Flags().SortFlags = false
	validateCmd.Flags().StringSliceVarP(&envVars, "env", "e", []string{}, "Set environment variables")
	validateCmd.Flags().StringVar(&environment, "environment", "dev", `Environment name`)
	validateCmd.Flags().BoolVar(&reset, "reset", false, "Clear and re-ingest source data")
	validateCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")
	validateCmd.Flags().BoolVar(&debug, "debug", false, "Collect additional debug info")
	validateCmd.Flags().StringVar(&logFormat, "log-format", "console", "Log format (options: \"console\", \"json\")")
	validateCmd.Flags().DurationVar(&reconcileTimeout, "reconcile-timeout", 60*time.Second, "Timeout for reconciliation (e.g. 60s, 2m)")

	return validateCmd
}

func resolveProjectPath(args []string) (string, error) {
	if len(args) == 0 {
		return ".", nil
	}

	projectPath := args[0]
	if strings.HasSuffix(projectPath, ".git") {
		repoName, err := gitutil.CloneRepo(projectPath)
		if err != nil {
			return "", fmt.Errorf("clone repo error: %w", err)
		}
		projectPath = repoName
	}

	return projectPath, nil
}

func enforceRepoLimits(ctx context.Context, projectPath string, ch *cmdutil.Helper) error {
	if _, err := os.Stat(projectPath); err != nil {
		return err
	}

	repo, _, err := cmdutil.RepoForProjectPath(projectPath)
	if err != nil {
		return err
	}
	_, err = repo.ListGlob(ctx, "**", false)
	if err != nil {
		if errors.Is(err, drivers.ErrRepoListLimitExceeded) {
			ch.PrintfError("The project directory exceeds the limit of %d files. Please open Rill against a directory with fewer files or set \"ignore_paths\" in rill.yaml.\n", drivers.RepoListLimit)
		}
		return fmt.Errorf("failed to list project files: %w", err)
	}
	return nil
}

func parseVariables(vals []string) (map[string]string, error) {
	res := make(map[string]string)
	for _, val := range vals {
		parsed, err := godotenv.Unmarshal(val)
		if err != nil {
			return nil, fmt.Errorf("failed to parse variable %q: %w", val, err)
		}
		for k, v := range parsed {
			res[k] = v
		}
	}
	return res, nil
}

func parseProject(ctx context.Context, ch *cmdutil.Helper, projectPath, environment string) error {
	repo, instanceID, err := cmdutil.RepoForProjectPath(projectPath)
	if err != nil {
		return err
	}

	p, err := parser.Parse(ctx, repo, instanceID, environment, defaultOLAPConnector)
	if err != nil {
		return fmt.Errorf("failed to parse project: %w", err)
	}
	if p.RillYAML == nil {
		return fmt.Errorf("failed to parse project: %w", parser.ErrRillYAMLNotFound)
	}
	if len(p.Errors) == 0 {
		return nil
	}

	ch.PrintfError("Parsing failed. Skipping reconciliation.\n\n")
	var table []*parseErrorTableRow
	for _, e := range p.Errors {
		table = append(table, &parseErrorTableRow{
			Path:  e.FilePath,
			Error: e.Message,
		})
	}
	ch.PrintData(table)
	return fmt.Errorf("project parsing failed")
}

func reconcileAndReport(ctx context.Context, ch *cmdutil.Helper, app *local.App, reconcileTimeout time.Duration) error {
	ctrl, err := app.Runtime.Controller(ctx, app.Instance.ID)
	if err != nil {
		return err
	}

	// Create a context with timeout for reconciliation
	reconcileCtx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	// Kick off reconciliation and wait for completion
	if err := ctrl.Reconcile(reconcileCtx, runtime.GlobalProjectParserName); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return reportTimeout(ctx, ch, ctrl, reconcileTimeout)
		}
		return fmt.Errorf("failed to start reconciliation: %w", err)
	}

	if err := ctrl.WaitUntilIdle(reconcileCtx, true); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return reportTimeout(ctx, ch, ctrl, reconcileTimeout)
		}
		return fmt.Errorf("failed while waiting for reconciliation to finish: %w", err)
	}

	resources, err := ctrl.List(ctx, "", "", true)
	if err != nil {
		return fmt.Errorf("failed to list resources: %w", err)
	}

	reconcileErrors := renderResourceStatus(ch, resources)

	if len(reconcileErrors) > 0 {
		return fmt.Errorf("reconciliation completed with %d error(s)", len(reconcileErrors))
	}

	ch.PrintfSuccess("\nValidation completed without errors.\n")
	return nil
}

func renderResourceStatus(ch *cmdutil.Helper, resources []*runtimev1.Resource) []string {
	var table []*resourceTableRow
	var reconcileErrors []string

	for _, r := range resources {
		if r.Meta.Hidden {
			continue
		}

		table = append(table, newResourceTableRow(r))
		if r.Meta.ReconcileError != "" {
			reconcileErrors = append(reconcileErrors, fmt.Sprintf("%s/%s: %s", r.Meta.Name.Kind, r.Meta.Name.Name, r.Meta.ReconcileError))
		}
	}

	ch.PrintfSuccess("\nReconcile status\n")
	ch.PrintData(table)
	ch.Printf("\n")

	return reconcileErrors
}

func reportTimeout(ctx context.Context, ch *cmdutil.Helper, ctrl *runtime.Controller, reconcileTimeout time.Duration) error {
	listCtx, cancel := context.WithTimeout(ctx, listResourcesTimeout)
	defer cancel()

	resources, err := ctrl.List(listCtx, "", "", true)
	if err != nil {
		return fmt.Errorf("reconciliation timed out after %s and listing resources failed: %w", reconcileTimeout, err)
	}

	ch.PrintfError("\nReconciliation timed out after %s. If a model processes full data, consider adding an explicit dev partition or rerun with --reconcile-timeout to allow more time. Check Pending or Running reconcile below\n", reconcileTimeout)
	// print status of all resources
	_ = renderResourceStatus(ch, resources)

	return fmt.Errorf("reconciliation timed out after %s", reconcileTimeout)
}

type resourceTableRow struct {
	Type   string `header:"type"`
	Name   string `header:"name"`
	Status string `header:"status"`
	Error  string `header:"error"`
}

func newResourceTableRow(r *runtimev1.Resource) *resourceTableRow {
	truncErr := r.Meta.ReconcileError
	if len(truncErr) > 120 {
		truncErr = truncErr[:120] + "..."
	}

	return &resourceTableRow{
		Type:   runtime.PrettifyResourceKind(r.Meta.Name.Kind),
		Name:   r.Meta.Name.Name,
		Status: runtime.PrettifyReconcileStatus(r.Meta.ReconcileStatus),
		Error:  truncErr,
	}
}

type parseErrorTableRow struct {
	Path  string `header:"path"`
	Error string `header:"error"`
}
