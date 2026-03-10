package python

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
)

func init() {
	drivers.Register("python", driver{})
	drivers.RegisterAsConnector("python", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Python",
	Description: "Execute Python scripts that produce data.",
	DocsURL:     "https://docs.rilldata.com/developers/build/connect/python",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "python_path",
			Type:        drivers.StringPropertyType,
			DisplayName: "Python Path",
			Description: "Path to Python executable. If empty, auto-detects python3 on the system.",
			Placeholder: ".rill/.venv/bin/python",
		},
		{
			Key:         "requirements",
			Type:        drivers.StringPropertyType,
			DisplayName: "Requirements",
			Description: "Comma-separated list of pip packages or a path to requirements.txt.",
		},
		{
			Key:         "packages",
			Type:        drivers.StringPropertyType,
			DisplayName: "Packages",
			Description: "List of pip packages to install in the virtual environment.",
		},
		{
			Key:         "venv_path",
			Type:        drivers.StringPropertyType,
			DisplayName: "Virtual Environment Path",
			Description: "Path to the virtual environment directory. Defaults to .rill/.venv.",
			Placeholder: ".rill/.venv",
		},
	},
	SourceProperties: []*drivers.PropertySpec{
		{
			Key:         "code_path",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Script Path",
			Description: "Path to the Python script relative to the project root.",
			Placeholder: "scripts/extract.py",
		},
		{
			Key:         "name",
			Type:        drivers.StringPropertyType,
			DisplayName: "Source name",
			Description: "The name of the source.",
			Placeholder: "my_python_source",
			Required:    true,
		},
	},
}

type driver struct{}

// ConfigProperties holds the connector-level configuration.
type ConfigProperties struct {
	PythonPath   string   `mapstructure:"python_path"`
	Requirements string   `mapstructure:"requirements"`
	Packages     []string `mapstructure:"packages"`
	VenvPath     string   `mapstructure:"venv_path"`
}

// ModelInputProperties holds the per-model properties from YAML.
type ModelInputProperties struct {
	CodePath          string            `mapstructure:"code_path"`
	Args              []string          `mapstructure:"args"`
	Env               map[string]string `mapstructure:"script_env"`
	CreateSecretsFromConnectors []string          `mapstructure:"create_secrets_from_connectors"`
}

// Decode parses raw properties into ModelInputProperties.
func (p *ModelInputProperties) Decode(props map[string]any) error {
	err := mapstructure.WeakDecode(props, p)
	if err != nil {
		return fmt.Errorf("failed to parse input properties: %w", err)
	}
	if p.CodePath == "" {
		return fmt.Errorf("missing property `code_path`")
	}
	return nil
}

func (d driver) Open(_, instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("python driver can't be shared")
	}

	cfg := &ConfigProperties{}
	err := mapstructure.WeakDecode(config, cfg)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		config: cfg,
		logger: logger,
	}
	return conn, nil
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return true, nil
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

// Connection implements drivers.Handle for the Python connector.
type Connection struct {
	config *ConfigProperties
	logger *zap.Logger
}

var _ drivers.Handle = &Connection{}

// Ping verifies that a Python binary is available and installs packages if configured.
func (c *Connection) Ping(ctx context.Context) error {
	pythonPath := c.resolvePythonPath()
	_, err := exec.LookPath(pythonPath)
	if err != nil {
		return fmt.Errorf("python not found at %q: %w (run 'rill python setup' to configure)", pythonPath, err)
	}
	// Install packages if configured
	if err := c.EnsurePackages(ctx); err != nil {
		return err
	}
	return nil
}

func (c *Connection) Driver() string {
	return "python"
}

func (c *Connection) Config() map[string]any {
	m := make(map[string]any)
	err := mapstructure.Decode(c.config, &m)
	if err != nil {
		c.logger.Warn("error generating python config", zap.Error(err))
	}
	return m
}

func (c *Connection) Migrate(ctx context.Context) error {
	return nil
}

func (c *Connection) MigrationStatus(ctx context.Context) (current int, desired int, err error) {
	return 0, 0, nil
}

func (c *Connection) Close() error {
	return nil
}

func (c *Connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

func (c *Connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

func (c *Connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

func (c *Connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

func (c *Connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

func (c *Connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

func (c *Connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

func (c *Connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

func (c *Connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

func (c *Connection) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

func (c *Connection) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

func (c *Connection) AsModelManager(instanceID string) (drivers.ModelManager, error) {
	return nil, drivers.ErrNotImplemented
}

func (c *Connection) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// resolvePythonPath returns the configured Python path or a default.
func (c *Connection) resolvePythonPath() string {
	if c.config.PythonPath != "" {
		return c.config.PythonPath
	}
	return "python3"
}

// EnsurePackages installs packages listed in the connector config if a venv exists.
// It merges packages from both the `packages` list and `requirements` string.
func (c *Connection) EnsurePackages(ctx context.Context) error {
	pkgs := c.allPackages()
	if len(pkgs) == 0 {
		return nil
	}

	// Find pip in the venv
	pythonPath := c.resolvePythonPath()
	venvDir := filepath.Dir(filepath.Dir(pythonPath)) // e.g. .rill/.venv/bin/python → .rill/.venv
	pipPath := filepath.Join(venvDir, "bin", "pip")
	if _, err := os.Stat(pipPath); err != nil {
		// Windows fallback
		pipPath = filepath.Join(venvDir, "Scripts", "pip.exe")
		if _, err := os.Stat(pipPath); err != nil {
			// No venv pip found; try using python -m pip
			pipPath = ""
		}
	}

	var cmd *exec.Cmd
	if pipPath != "" {
		args := append([]string{"install", "-q"}, pkgs...)
		cmd = exec.CommandContext(ctx, pipPath, args...)
	} else {
		args := append([]string{"-m", "pip", "install", "-q"}, pkgs...)
		cmd = exec.CommandContext(ctx, pythonPath, args...)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("python: failed to install packages %v: %w\n%s", pkgs, err, string(out))
	}

	c.logger.Info("python: installed packages", zap.Strings("packages", pkgs))
	return nil
}

// allPackages merges packages from the `packages` list and `requirements` string.
func (c *Connection) allPackages() []string {
	var pkgs []string
	pkgs = append(pkgs, c.config.Packages...)
	if c.config.Requirements != "" {
		for _, p := range strings.Split(c.config.Requirements, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				pkgs = append(pkgs, p)
			}
		}
	}
	return pkgs
}
