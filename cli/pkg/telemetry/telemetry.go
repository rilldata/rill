package telemetry

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
)

type Telemetry struct {
	Enabled     bool
	InstallID   string
	Version     string
	BuildCommit string
	BuildTime   string
	IsDev       bool
	authHeader  string
}

const (
	RillIntakeURL  = "https://intake.rilldata.io/events/data-modeler-metrics"
	RillIntakeUser = "data-modeler"
	// TODO: get this from vault and embedd into the binary
	RillIntakePassword = "lkh8T90ozWJP/KxWnQ81PexRzpdghPdzuB0ly2/86TeUU8q/bKiVug=="
	RillDeveloperApp   = "rill-developer"
)

var ErrRillIntake = errors.New("failed to fire telemetry")

func NewTelemetry(ver config.Version) (*Telemetry, error) {
	installID, enabled, err := dotrill.AnalyticsInfo()
	if err != nil {
		return nil, err
	}

	encodedAuth := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", RillIntakeUser, RillIntakePassword)),
	)

	return &Telemetry{
		Enabled:     enabled,
		InstallID:   installID,
		Version:     ver.Number,
		BuildCommit: ver.Commit,
		BuildTime:   ver.Timestamp,
		IsDev:       ver.IsDev(),
		authHeader:  fmt.Sprintf("Basic %s", encodedAuth),
	}, nil
}

func (t *Telemetry) emit(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, RillIntakeURL, bytes.NewReader(body))
	if err != nil {
		return ErrRillIntake
	}
	req.Header = http.Header{
		"Authorization": []string{t.authHeader},
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ErrRillIntake
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ErrRillIntake
	}

	return nil
}

type BehaviourEventFields struct {
	AppName    string
	InstallID  string
	BuildID    string
	Version    string
	IsDev      bool
	Mode       string
	Action     string
	Medium     string
	Space      string
	ScreenName string
}

func (t *Telemetry) emitBehaviourEvent(ctx context.Context, action, medium, space, screenName string) error {
	fields := BehaviourEventFields{
		AppName:    RillDeveloperApp,
		InstallID:  t.InstallID,
		BuildID:    t.BuildCommit,
		Version:    t.Version,
		IsDev:      t.IsDev,
		Mode:       "edit",
		Action:     action,
		Medium:     medium,
		Space:      space,
		ScreenName: screenName,
	}
	body, err := json.Marshal(&fields)
	if err != nil {
		return err
	}

	return t.emit(ctx, body)
}

// Error events are not needed. Will be inferred from missing events by product.
// For internal debugging we should use logs.

func (t *Telemetry) EmitDeployStart(ctx context.Context) error {
	return t.emitBehaviourEvent(ctx, "deploy-start", "cli", "terminal", "terminal")
}

func (t *Telemetry) EmitDeploySuccess(ctx context.Context) error {
	return t.emitBehaviourEvent(ctx, "deploy-success", "cli", "terminal", "terminal")
}

func (t *Telemetry) EmitGithubConnectedStart(ctx context.Context) error {
	return t.emitBehaviourEvent(ctx, "ghconnected-start", "cli", "terminal", "terminal")
}

func (t *Telemetry) EmitGithubConnectedSuccess(ctx context.Context) error {
	return t.emitBehaviourEvent(ctx, "ghconnected-success", "cli", "terminal", "terminal")
}

func (t *Telemetry) EmitDataAccessConnectedStart(ctx context.Context) error {
	return t.emitBehaviourEvent(ctx, "dataaccess-start", "cli", "terminal", "terminal")
}

func (t *Telemetry) EmitDataAccessConnectedSuccess(ctx context.Context) error {
	return t.emitBehaviourEvent(ctx, "dataaccess-success", "cli", "terminal", "terminal")
}
