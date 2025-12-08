package validate

import (
	"context"
	"encoding/json"
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

// ValidationResult represents the complete validation output
type ValidationResult struct {
	Success     bool              `json:"success"`
	Summary     ValidationSummary `json:"summary"`
	ParseErrors []ParseError      `json:"parse_errors,omitempty"`
	Resources   []ResourceStatus  `json:"resources,omitempty"`
	Timeout     bool              `json:"timeout,omitempty"`
}

type ValidationSummary struct {
	TotalResources  int `json:"total_resources"`
	ParseErrors     int `json:"parse_errors"`
	ReconcileErrors int `json:"reconcile_errors"`
}

// ParseError represents a parse error (serializable version of runtimev1.ParseError)
type ParseError struct {
	Message       string        `json:"message" header:"message"`
	FilePath      string        `json:"file_path,omitempty" header:"file_path"`
	StartLocation *CharLocation `json:"start_location,omitempty" header:"line"`
	External      bool          `json:"external,omitempty" header:"external"`
}

// CharLocation represents a character location in a file
type CharLocation struct {
	Line uint32 `json:"line,omitempty"`
}

type ResourceStatus struct {
	Kind     string `json:"kind" header:"kind"`
	Name     string `json:"name" header:"name"`
	Status   string `json:"status" header:"status"`
	Error    string `json:"error,omitempty" header:"error"`
	FilePath string `json:"file_path,omitempty" header:"file_path"`
}

// ValidateCmd validates and reconciles a project without starting the UI.
func ValidateCmd(ch *cmdutil.Helper) *cobra.Command {
	var verbose bool
	var debug bool
	var silent bool
	var reset bool
	var logFormat string
	var envVars []string
	var environment string
	var reconcileTimeout time.Duration
	var outputFile string
	var outputFormat string

	validateCmd := &cobra.Command{
		Use:   "validate [<path>]",
		Short: "Validate project resources",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath, err := resolveProjectPath(args)
			if err != nil {
				return err
			}

			// Validate output format
			if outputFormat != "table" && outputFormat != "json" {
				return fmt.Errorf("invalid output format %q (options: \"table\", \"json\")", outputFormat)
			}

			if !local.IsProjectInit(projectPath) {
				return fmt.Errorf("no Rill project found at %q (missing rill.yaml)", projectPath)
			}

			environment = "dev" // always use dev environment for validation
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

			parseErrors, err := parseProject(cmd.Context(), projectPath, environment)
			if err != nil {
				return err
			}

			// If there are parse errors, output them and exit
			if len(parseErrors) > 0 {
				result := &ValidationResult{
					Success: false,
					Summary: ValidationSummary{
						ParseErrors: len(parseErrors),
					},
					ParseErrors: parseErrors,
				}
				return outputResult(ch, result, outputFormat, outputFile)
			}

			ch.Interactive = false
			app, err := local.NewApp(cmd.Context(), &local.AppOptions{
				Ch:             ch,
				Verbose:        verbose,
				Debug:          debug,
				Silent:         silent,
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

			return reconcileAndReport(cmd.Context(), ch, app, reconcileTimeout, outputFormat, outputFile)
		},
	}

	validateCmd.Flags().SortFlags = false
	validateCmd.Flags().StringSliceVarP(&envVars, "env", "e", []string{}, "Set environment variables")
	validateCmd.Flags().BoolVar(&reset, "reset", false, "Clear and re-ingest source data")
	validateCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")
	validateCmd.Flags().BoolVar(&debug, "debug", false, "Collect additional debug info")
	validateCmd.Flags().BoolVar(&silent, "silent", false, "Suppress all log output")
	validateCmd.Flags().StringVar(&logFormat, "log-format", "console", "Log format (options: \"console\", \"json\")")
	validateCmd.Flags().DurationVar(&reconcileTimeout, "reconcile-timeout", 60*time.Second, "Timeout for reconciliation (e.g. 60s, 2m)")
	validateCmd.Flags().StringVar(&outputFormat, "output-format", "table", "Output format (options: \"table\", \"json\")")
	validateCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for validation results (JSON format)")

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

func parseProject(ctx context.Context, projectPath, environment string) ([]ParseError, error) {
	repo, instanceID, err := cmdutil.RepoForProjectPath(projectPath)
	if err != nil {
		return nil, err
	}

	p, err := parser.Parse(ctx, repo, instanceID, environment, defaultOLAPConnector)
	if err != nil {
		return nil, fmt.Errorf("failed to parse project: %w", err)
	}
	if p.RillYAML == nil {
		return nil, fmt.Errorf("failed to parse project: %w", parser.ErrRillYAMLNotFound)
	}
	if len(p.Errors) == 0 {
		return nil, nil
	}

	var parseErrors []ParseError
	for _, e := range p.Errors {
		parseErr := ParseError{
			Message:  e.Message,
			FilePath: e.FilePath,
			External: e.External,
		}
		// Convert start location if available
		if e.StartLocation != nil {
			parseErr.StartLocation = &CharLocation{
				Line: e.StartLocation.Line,
			}
		}
		parseErrors = append(parseErrors, parseErr)
	}
	return parseErrors, nil
}

func reconcileAndReport(ctx context.Context, ch *cmdutil.Helper, app *local.App, reconcileTimeout time.Duration, outputFormat, outputFile string) error {
	ctrl, err := app.Runtime.Controller(ctx, app.Instance.ID)
	if err != nil {
		return err
	}

	// Create a context with timeout for reconciliation
	reconcileCtx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	timedOut := false
	// Kick off reconciliation and wait for completion
	if err := ctrl.Reconcile(reconcileCtx, runtime.GlobalProjectParserName); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			timedOut = true
		} else {
			return fmt.Errorf("failed to start reconciliation: %w", err)
		}
	}

	if !timedOut {
		if err := ctrl.WaitUntilIdle(reconcileCtx, true); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				timedOut = true
			} else {
				return fmt.Errorf("failed while waiting for reconciliation to finish: %w", err)
			}
		}
	}

	// List resources (use parent ctx with a timeout to avoid blocking indefinitely)
	listCtx, listCancel := context.WithTimeout(ctx, listResourcesTimeout)
	defer listCancel()

	resources, err := ctrl.List(listCtx, "", "", false)
	if err != nil {
		return fmt.Errorf("failed to list resources: %w", err)
	}

	// Build validation result
	result := buildValidationResult(resources, timedOut)

	// Output the result
	return outputResult(ch, result, outputFormat, outputFile)
}

func buildValidationResult(resources []*runtimev1.Resource, timedOut bool) *ValidationResult {
	result := &ValidationResult{
		Success: !timedOut,
		Timeout: timedOut,
		Summary: ValidationSummary{},
	}

	for _, r := range resources {
		if r.Meta.Hidden {
			continue
		}

		result.Summary.TotalResources++

		status := runtime.PrettifyReconcileStatus(r.Meta.ReconcileStatus)
		resourceStatus := ResourceStatus{
			Kind:   runtime.PrettifyResourceKind(r.Meta.Name.Kind),
			Name:   r.Meta.Name.Name,
			Status: status,
		}

		// just use first file path if multiple for now
		if len(r.Meta.FilePaths) > 0 {
			resourceStatus.FilePath = r.Meta.FilePaths[0]
		}

		if r.Meta.ReconcileError != "" {
			result.Summary.ReconcileErrors++
			result.Success = false
			resourceStatus.Error = r.Meta.ReconcileError
		}

		result.Resources = append(result.Resources, resourceStatus)
	}

	return result
}

func outputResult(ch *cmdutil.Helper, result *ValidationResult, outputFormat, outputFile string) error {
	// Write JSON to file if requested
	var jsonData []byte
	var err error
	if outputFile != "" || outputFormat == "json" {
		jsonData, err = json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal validation result: %w", err)
		}
	}
	if outputFile != "" {
		if err = os.WriteFile(outputFile, jsonData, 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		ch.Printf("Validation results written to output file %s\n", outputFile)
	} else {
		// Output to console based on format
		if outputFormat == "json" {
			fmt.Println(string(jsonData))
		} else {
			// Table format - show parse errors and resources separately
			if len(result.ParseErrors) > 0 {
				ch.PrintfError("\nParse Errors\n")
				ch.PrintData(result.ParseErrors)
				ch.Printf("\n")
			}

			if len(result.Resources) > 0 {
				ch.PrintfSuccess("\nResources\n")
				ch.PrintData(result.Resources)
				ch.Printf("\n")
			}
		}
	}

	// Print summary in the end
	if !result.Success {
		if result.Timeout {
			return fmt.Errorf("reconciliation timed out. If a model processes full data, consider adding an explicit dev partition or rerun with --reconcile-timeout to allow more time")
		}
		return fmt.Errorf("validation failed: %d error(s) (%d parse, %d reconcile)", result.Summary.ParseErrors+result.Summary.ReconcileErrors, result.Summary.ParseErrors, result.Summary.ReconcileErrors)
	}

	ch.PrintfSuccess("Validation completed successfully\n")
	return nil
}
