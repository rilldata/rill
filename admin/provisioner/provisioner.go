package provisioner

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/database"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/client"
	"github.com/rilldata/rill/runtime/drivers/github"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.uber.org/zap"
)

type Instance struct {
	Host       string
	Audience   string
	InstanceID string
}

type ProvisionOptions struct {
	OLAPDriver           string
	OLAPDSN              string
	Slots                int
	GithubURL            string
	GitBranch            string
	GithubInstallationID int64
	Region               string
	Variables            map[string]string
}

type Provisioner interface {
	Provision(ctx context.Context, opts *ProvisionOptions) (*Instance, error)
	Teardown(ctx context.Context, host, instanceID, olapDriver string) error
	Close() error
}

type staticSpec struct {
	Runtimes []*staticRuntime `json:"runtimes"`
}

type staticRuntime struct {
	Host     string `json:"host"`
	Region   string `json:"region"`
	Slots    int    `json:"slots"`
	DataDir  string `json:"data_dir"`
	Audience string `json:"audience_url"`
}

type staticProvisioner struct {
	spec   *staticSpec
	logger *zap.Logger
	db     database.DB
	issuer *auth.Issuer
}

func NewStatic(spec string, logger *zap.Logger, db database.DB, issuer *auth.Issuer) (Provisioner, error) {
	sps := &staticSpec{}
	err := json.Unmarshal([]byte(spec), sps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provisioner spec: %w", err)
	}

	return &staticProvisioner{
		spec:   sps,
		logger: logger,
		db:     db,
		issuer: issuer,
	}, nil
}

func (p *staticProvisioner) Provision(ctx context.Context, opts *ProvisionOptions) (*Instance, error) {
	// Get slots currently used
	stats, err := p.db.ResolveRuntimeSlotsUsed(ctx)
	if err != nil {
		return nil, err
	}

	// Find runtime with available capacity
	var target *staticRuntime
	for _, candidate := range p.spec.Runtimes {
		if opts.Region != "" && opts.Region != candidate.Region {
			continue
		}

		available := true
		for _, stat := range stats {
			if stat.RuntimeHost == candidate.Host && stat.SlotsUsed+opts.Slots > candidate.Slots {
				available = false
				break
			}
		}

		if available {
			target = candidate
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("no runtimes found with sufficient available slots")
	}

	// Create JWT for runtime client
	jwt, err := p.issuer.NewToken(auth.TokenOptions{
		AudienceURL:       target.Audience,
		TTL:               time.Hour,
		SystemPermissions: []auth.Permission{auth.ManageInstances},
	})
	if err != nil {
		return nil, err
	}

	// Make runtime client
	rt, err := client.New(target.Host, jwt)
	if err != nil {
		return nil, err
	}
	defer rt.Close()

	// Build repo info
	repoDSN, err := json.Marshal(github.DSN{
		GithubURL:      opts.GithubURL,
		Branch:         opts.GitBranch,
		InstallationID: opts.GithubInstallationID,
	})
	if err != nil {
		return nil, err
	}

	// Generate new instanceID
	instanceID := strings.ReplaceAll(uuid.New().String(), "-", "")

	// Build instance config DSN
	var ingestLimit int64
	var embedCatalog bool
	if opts.OLAPDriver == "duckdb" {
		if opts.OLAPDSN != "" {
			return nil, fmt.Errorf("passing a DSN is not allowed for driver 'duckdb'")
		}

		embedCatalog = true

		ingestLimit = int64(datasize.GB * datasize.ByteSize(5*opts.Slots)) // 5GB * slots
		cpus := 1 * opts.Slots
		memory := 2 * opts.Slots

		opts.OLAPDSN = fmt.Sprintf("%s.db?rill_pool_size=%d&threads=%d&max_memory=%dGB", path.Join(target.DataDir, instanceID), cpus, cpus, memory)
	}

	// Create the instance
	_, err = rt.CreateInstance(ctx, &runtimev1.CreateInstanceRequest{
		InstanceId:          instanceID,
		OlapDriver:          opts.OLAPDriver,
		OlapDsn:             opts.OLAPDSN,
		RepoDriver:          "github",
		RepoDsn:             string(repoDSN),
		EmbedCatalog:        embedCatalog,
		Variables:           opts.Variables,
		IngestionLimitBytes: ingestLimit,
	})
	if err != nil {
		return nil, err
	}

	inst := &Instance{
		Host:       target.Host,
		Audience:   target.Audience,
		InstanceID: instanceID,
	}
	return inst, nil
}

func (p *staticProvisioner) Teardown(ctx context.Context, host, instanceID, olapDriver string) error {
	// Find audience
	var audience string
	for _, candidate := range p.spec.Runtimes {
		if candidate.Host == host {
			audience = candidate.Audience
			break
		}
	}
	if audience == "" {
		return fmt.Errorf("could not find a runtime matching host %q", host)
	}

	// Create JWT for runtime client
	jwt, err := p.issuer.NewToken(auth.TokenOptions{
		AudienceURL:       audience,
		TTL:               time.Hour,
		SystemPermissions: []auth.Permission{auth.ManageInstances},
	})
	if err != nil {
		return err
	}

	// Make runtime client
	rt, err := client.New(host, jwt)
	if err != nil {
		return err
	}
	defer rt.Close()

	// Only drop DB if it's DuckDB
	dropDB := false
	if olapDriver == "duckdb" {
		dropDB = true
	}

	// Delete the instance
	_, err = rt.DeleteInstance(ctx, &runtimev1.DeleteInstanceRequest{
		InstanceId: instanceID,
		DropDb:     dropDB,
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *staticProvisioner) Close() error {
	return nil
}
