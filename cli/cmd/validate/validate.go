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
	"github.com/rilldata/rill/cli/pkg/printer"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/spf13/cobra"
)

// ValidationResult represents the complete validation output
type ValidationResult struct {
	Summary     ValidationSummary `json:"summary"`
	ParseErrors []ParseError      `json:"parse_errors,omitempty"`
	Resources   []ResourceStatus  `json:"resources,omitempty"`
}

type ValidationSummary struct {
	TotalResources  int `json:"total_resources"`
	ParseErrors     int `json:"parse_errors"`
	ReconcileErrors int `json:"reconcile_errors"`
}

// ParseError represents a parse error (serializable version of runtimev1.ParseError)
type ParseError struct {
	Message   string `json:"message" header:"message"`
	FilePath  string `json:"file_path,omitempty" header:"file_path"`
	StartLine uint32 `json:"start_line,omitempty" header:"start_line"`
}

type ResourceStatus struct {
	Kind     string `json:"kind" header:"kind"`
	Name     string `json:"name" header:"name"`
	Status   string `json:"status" header:"status"`
	Error    string `json:"error,omitempty" header:"error"`
	FilePath string `json:"file_path,omitempty" header:"file_path"`
	Timeout  bool   `json:"timeout,omitempty" header:"timeout"`
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
	var modelTimeoutSeconds uint32
	var outputFile string

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
			outputFormat := ch.Printer.Format
			switch outputFormat {
			case printer.FormatHuman, printer.FormatJSON:
				// supported formats
			default:
				return fmt.Errorf("only human and json output format is supported for validate command")
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
			envVarsMap["rill.model.timeout_override"] = fmt.Sprintf("%d", modelTimeoutSeconds)
			// Prevent resource updates when parse errors are present and surface the actual parser output instead of re-parsing here.
			envVarsMap["rill.parser.skip_updates_if_parse_errors"] = "true"

			ch.Interactive = false
			app, err := local.NewApp(cmd.Context(), &local.AppOptions{
				Ch:             ch,
				Verbose:        verbose,
				Silent:         silent,
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

			return reconcileAndReport(cmd.Context(), ch, app, outputFormat, outputFile)
		},
	}

	validateCmd.Flags().SortFlags = false
	validateCmd.Flags().StringSliceVarP(&envVars, "env", "e", []string{}, "Set environment variables")
	validateCmd.Flags().BoolVar(&reset, "reset", false, "Clear and re-ingest source data")
	validateCmd.Flags().StringVar(&environment, "environment", "dev", `Environment name`)
	validateCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")
	validateCmd.Flags().BoolVar(&silent, "silent", false, "Suppress all log output by setting log level to panic, overrides verbose flag")
	validateCmd.Flags().BoolVar(&debug, "debug", false, "Collect additional debug info")
	validateCmd.Flags().StringVar(&logFormat, "log-format", "console", "Log format (options: \"console\", \"json\")")
	validateCmd.Flags().Uint32Var(&modelTimeoutSeconds, "model-timeout-seconds", 60, "Timeout for reconciliation of models, set 0 for no timeout")
	validateCmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "Output file for validation results (JSON format)")

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

func reconcileAndReport(ctx context.Context, ch *cmdutil.Helper, app *local.App, outputFormat printer.Format, outputFile string) error {
	ctrl, err := app.Runtime.Controller(ctx, app.Instance.ID)
	if err != nil {
		return err
	}

	// Kick off reconciliation and wait for completion
	if err := ctrl.Reconcile(ctx, runtime.GlobalProjectParserName); err != nil {
		return fmt.Errorf("failed to start reconciliation: %w", err)
	}

	time.Sleep(3 * time.Second) // brief sleep to allow reconciliation to start

	if err := ctrl.WaitUntilIdle(ctx, true); err != nil {
		return fmt.Errorf("failed while waiting for reconciliation to finish: %w", err)
	}

	resources, err := ctrl.List(ctx, "", "", false)
	if err != nil {
		return fmt.Errorf("failed to list resources: %w", err)
	}

	// Build validation result
	result := buildValidationResult(resources)

	// Output the result
	return outputResult(ch, result, outputFormat, outputFile)
}

func buildValidationResult(resources []*runtimev1.Resource) *ValidationResult {
	result := &ValidationResult{
		Summary: ValidationSummary{},
	}

	for _, r := range resources {
		if r.Meta.Name.Kind == runtime.GlobalProjectParserName.Kind && r.Meta.Name.Name == runtime.GlobalProjectParserName.Name {
			parseErrors := parseErrorsFromParser(r)
			if len(parseErrors) > 0 {
				result.Summary.ParseErrors = len(parseErrors)
			}
			continue
		}
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
			resourceStatus.Error = r.Meta.ReconcileError
			// Check if the error is due to context deadline exceeded (timeout)
			if strings.Contains(r.Meta.ReconcileError, "context deadline exceeded") {
				resourceStatus.Timeout = true
			}
		}

		result.Resources = append(result.Resources, resourceStatus)
	}

	return result
}

func parseErrorsFromParser(parserRes *runtimev1.Resource) []ParseError {
	if parserRes == nil || parserRes.GetProjectParser() == nil {
		return nil
	}

	var parseErrors []ParseError
	for _, e := range parserRes.GetProjectParser().State.ParseErrors {
		if e == nil {
			continue
		}
		pe := ParseError{
			Message:  e.Message,
			FilePath: e.FilePath,
		}
		if e.StartLocation != nil {
			pe.StartLine = e.StartLocation.Line
		}
		parseErrors = append(parseErrors, pe)
	}
	return parseErrors
}

func outputResult(ch *cmdutil.Helper, result *ValidationResult, outputFormat printer.Format, outputFile string) error {
	// Write JSON to file if requested
	var jsonData []byte
	var err error
	if outputFile != "" || outputFormat == printer.FormatJSON {
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
		if outputFormat == printer.FormatJSON {
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
	if result.Summary.ParseErrors+result.Summary.ReconcileErrors > 0 {
		return fmt.Errorf("validation failed: %d error(s) (%d parse, %d reconcile)", result.Summary.ParseErrors+result.Summary.ReconcileErrors, result.Summary.ParseErrors, result.Summary.ReconcileErrors)
	}

	ch.PrintfSuccess("Validation completed successfully\n")
	return nil
}
