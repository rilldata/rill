package telemetry

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
)

type Telemetry struct {
	Enabled   bool
	InstallID string
	Version   config.Version
	events    [][]byte
}

const (
	intakeURL    = "https://intake.rilldata.io/events/data-modeler-metrics"
	intakeUser   = "data-modeler"
	intakeKey    = "lkh8T90ozWJP/KxWnQ81PexRzpdghPdzuB0ly2/86TeUU8q/bKiVug=="
	developerApp = "rill-developer"
)

var ErrRillIntake = errors.New("failed to fire telemetry")

func NewTelemetry(ver config.Version) *Telemetry {
	installID, enabled, err := dotrill.AnalyticsInfo()
	if err != nil {
		// if there is any error just disable the telemetry.
		// this is simpler than null checking everywhere telemetry methods are called
		enabled = false
	}

	return &Telemetry{
		Enabled:   enabled,
		InstallID: installID,
		Version:   ver,
		events:    make([][]byte, 0),
	}
}

func (t *Telemetry) emit(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, intakeURL, bytes.NewReader(body))
	if err != nil {
		return ErrRillIntake
	}
	req.Header = http.Header{
		"Authorization": []string{fmt.Sprintf(
			"Basic %s",
			base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:%s", intakeUser, intakeKey)),
			),
		)},
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
	AppName       string `json:"app_name"`
	InstallID     string `json:"install_id"`
	BuildID       string `json:"build_id"`
	Version       string `json:"version"`
	IsDev         bool   `json:"is_dev"`
	Mode          string `json:"mode"`
	Action        string `json:"action"`
	Medium        string `json:"medium"`
	Space         string `json:"space"`
	ScreenName    string `json:"screen_name"`
	EventDatetime int64  `json:"event_datetime"`
	EventType     string `json:"event_type"`
}

func (t *Telemetry) emitBehaviourEvent(action, medium, space, screenName string) {
	if !t.Enabled {
		return
	}

	fields := BehaviourEventFields{
		AppName:       developerApp,
		InstallID:     t.InstallID,
		BuildID:       t.Version.Commit,
		Version:       t.Version.Number,
		IsDev:         t.Version.IsDev(),
		Mode:          "edit",
		Action:        action,
		Medium:        medium,
		Space:         space,
		ScreenName:    screenName,
		EventDatetime: time.Now().Unix() * 1000,
		EventType:     "behavioral",
	}
	event, err := json.Marshal(&fields)
	if err != nil {
		return
	}

	t.events = append(t.events, event)
}

func (t *Telemetry) Flush(ctx context.Context) error {
	if len(t.events) == 0 {
		return nil
	}

	body := make([]byte, 0)
	for _, event := range t.events {
		body = append(body, event...)
		body = append(body, '\n')
	}

	t.events = make([][]byte, 0)
	return t.emit(ctx, body)
}

// Error events are not needed. Will be inferred from missing events by product.
// For internal debugging we should use logs.

func (t *Telemetry) EmitDeployStart() {
	t.emitBehaviourEvent("deploy-start", "cli", "terminal", "terminal")
}

func (t *Telemetry) EmitDeploySuccess() {
	t.emitBehaviourEvent("deploy-success", "cli", "terminal", "terminal")
}

func (t *Telemetry) EmitGithubConnectedStart() {
	t.emitBehaviourEvent("ghconnected-start", "cli", "terminal", "terminal")
}

func (t *Telemetry) EmitGithubConnectedSuccess() {
	t.emitBehaviourEvent("ghconnected-success", "cli", "terminal", "terminal")
}

func (t *Telemetry) EmitDataAccessConnectedStart() {
	t.emitBehaviourEvent("dataaccess-start", "cli", "terminal", "terminal")
}

func (t *Telemetry) EmitDataAccessConnectedSuccess() {
	t.emitBehaviourEvent("dataaccess-success", "cli", "terminal", "terminal")
}

func (t *Telemetry) EmitLoginStart() {
	t.emitBehaviourEvent("login-start", "cli", "terminal", "terminal")
}

func (t *Telemetry) EmitLoginSuccess() {
	t.emitBehaviourEvent("login-success", "cli", "terminal", "terminal")
}

func (t *Telemetry) EmitLogoutStart() {
	t.emitBehaviourEvent("logout-start", "cli", "terminal", "terminal")
}

func (t *Telemetry) EmitLogoutSuccess() {
	t.emitBehaviourEvent("logout-success", "cli", "terminal", "terminal")
}
