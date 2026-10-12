package typesafe

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/mitchellh/mapstructure"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const defaultBaseURL = "https://api.typesafe.ai"

func init() {
	drivers.Register("typesafe", driver{})
	drivers.RegisterAsConnector("typesafe", driver{})
}

var spec = drivers.Spec{
	DisplayName: "Typesafe",
	Description: "Connect to an Typesafe API server for language models.",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "api_key",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "API Key",
			Description: "API key for connecting to a Typesafe server.",
			Secret:      true,
		},
		{
			Key:         "model",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Model",
			Description: "The model to use.",
			Placeholder: "",
		},
		{
			Key:         "base_url",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Base URL",
			Description: "Custom base URL for the Typesafe API.",
			Placeholder: "",
		},
	},
	ImplementsAI: true,
}

type driver struct{}

var _ drivers.Driver = driver{}

// Spec implements drivers.Driver.
func (d driver) Spec() drivers.Spec {
	return spec
}

// Open implements drivers.Driver.
func (d driver) Open(_, instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	conf := &configProperties{}
	err := mapstructure.WeakDecode(config, conf)
	if err != nil {
		return nil, err
	}

	return &handle{
		config: conf,
	}, nil
}

// HasAnonymousSourceAccess implements drivers.Driver.
func (d driver) HasAnonymousSourceAccess(ctx context.Context, srcProps map[string]any, logger *zap.Logger) (bool, error) {
	return false, drivers.ErrNotImplemented
}

// TertiarySourceConnectors implements drivers.Driver.
func (d driver) TertiarySourceConnectors(ctx context.Context, srcProps map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, drivers.ErrNotImplemented
}

type configProperties struct {
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	BaseURL string `mapstructure:"base_url"`
}

type handle struct {
	config *configProperties
}

var _ drivers.AIService = (*handle)(nil)

// AsAI implements drivers.Handle.
func (h *handle) AsAI(instanceID string) (drivers.AIService, bool) {
	return h, true
}

// AsAdmin implements drivers.Handle.
func (h *handle) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Handle.
func (h *handle) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsFileStore implements drivers.Handle.
func (h *handle) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsInformationSchema implements drivers.Handle.
func (h *handle) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (h *handle) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (h *handle) AsModelManager(instanceID string) (drivers.ModelManager, error) {
	return nil, drivers.ErrNotImplemented
}

// AsNotifier implements drivers.Handle.
func (h *handle) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

// AsOLAP implements drivers.Handle.
func (h *handle) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Handle.
func (h *handle) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsRegistry implements drivers.Handle.
func (h *handle) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Handle.
func (h *handle) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (h *handle) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// Close implements drivers.Handle.
func (h *handle) Close() error {
	return nil
}

// Config implements drivers.Handle.
func (h *handle) Config() map[string]any {
	var configMap map[string]any
	_ = mapstructure.Decode(h.config, &configMap)
	return configMap
}

// Driver implements drivers.Handle.
func (h *handle) Driver() string {
	return "typesafe"
}

// Migrate implements drivers.Handle.
func (h *handle) Migrate(ctx context.Context) error {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (h *handle) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// Ping implements drivers.Handle.
func (h *handle) Ping(ctx context.Context) error {
	return nil
}

// Complete implements drivers.AIService.
func (h *handle) Complete(ctx context.Context, opts *drivers.CompleteOptions) (*drivers.CompleteResult, error) {
	return nil, errors.New("not implemented")
}

// Evaluate implements drivers.AIService.
// It calls the TypeSafe evaluation endpoint; see https://docs.typesafe.ai/api.
// TODO: Retry 429 and 529 responses with exponential backoff; see https://docs.typesafe.ai/api#handling-rate-limits.
func (h *handle) Evaluate(ctx context.Context, req *aiv1.EvaluateRequest) (*aiv1.EvaluateResponse, error) {
	model := req.Model
	if model == "" {
		model = h.config.Model
	}

	body := &evaluateRequest{
		Model:     model,
		State:     req.State,
		Questions: make(map[string]*evaluateQuestion, len(req.Questions)),
	}
	for id, q := range req.Questions {
		wq, err := convertQuestion(q)
		if err != nil {
			return nil, fmt.Errorf("invalid question %q: %w", id, err)
		}
		body.Questions[id] = wq
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	baseURL := h.config.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/systemone", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+h.config.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpRes.Body.Close()

	resBody, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if httpRes.StatusCode < 200 || httpRes.StatusCode >= 300 {
		return nil, fmt.Errorf("evaluate request failed with status %d: %s", httpRes.StatusCode, resBody)
	}

	res := &evaluateResponse{}
	if err := json.Unmarshal(resBody, res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	answers := make(map[string]*aiv1.EvaluateAnswer, len(res.Answers))
	for id, a := range res.Answers {
		answer, err := convertAnswer(a)
		if err != nil {
			return nil, fmt.Errorf("invalid answer %q: %w", id, err)
		}
		answers[id] = answer
	}

	return &aiv1.EvaluateResponse{
		Model:   res.Model,
		Answers: answers,
	}, nil
}

// evaluateRequest is the JSON request body of the TypeSafe evaluation endpoint.
type evaluateRequest struct {
	Model     string                       `json:"model"`
	State     *structpb.Value              `json:"state"`
	Questions map[string]*evaluateQuestion `json:"questions"`
}

// evaluateQuestion is a question in the TypeSafe API.
// The question type's oneof case is flattened into a "type" field next to the question's fields.
// Criteria has a different shape per type, so it is untyped.
type evaluateQuestion struct {
	Type         string          `json:"type"`
	Instructions *structpb.Value `json:"instructions"`
	Criteria     any             `json:"criteria,omitempty"`
}

// evaluateNoulCriteria is the criteria of a noul question in the TypeSafe API.
type evaluateNoulCriteria struct {
	True  *structpb.Value `json:"true,omitempty"`
	False *structpb.Value `json:"false,omitempty"`
}

// evaluateResponse is the JSON response body of the TypeSafe evaluation endpoint.
type evaluateResponse struct {
	Model   string                     `json:"model"`
	Answers map[string]*evaluateAnswer `json:"answers"`
	// TODO: Propagate usage once aiv1.EvaluateResponse has a field for it.
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// evaluateAnswer is an answer in the TypeSafe API.
// It holds the union of fields across all answer types; the "type" field determines which ones are set.
type evaluateAnswer struct {
	Type          string             `json:"type"`
	Noul          float64            `json:"noul"`
	Choice        string             `json:"choice"`
	Score         float64            `json:"score"`
	Legend        map[string]string  `json:"legend"`
	Probabilities map[string]float64 `json:"probabilities"`
	Confidence    float64            `json:"confidence"`
}

// convertQuestion converts a Rill question to a TypeSafe question.
func convertQuestion(q *aiv1.EvaluateQuestion) (*evaluateQuestion, error) {
	switch q := q.Question.(type) {
	case *aiv1.EvaluateQuestion_Noul:
		res := &evaluateQuestion{
			Type:         "noul",
			Instructions: q.Noul.Instructions,
		}
		// Guard against assigning a typed nil to the interface, which would serialize as null.
		if c := q.Noul.Criteria; c != nil {
			res.Criteria = &evaluateNoulCriteria{True: c.True, False: c.False}
		}
		return res, nil
	case *aiv1.EvaluateQuestion_Choice:
		// The API accepts null for options that need no description.
		// A null value doesn't survive a protobuf round trip; it arrives as a Value with no kind, which fails to serialize.
		criteria := make(map[string]*structpb.Value, len(q.Choice.Criteria))
		for k, v := range q.Choice.Criteria {
			if v.GetKind() == nil {
				v = nil
			}
			criteria[k] = v
		}
		return &evaluateQuestion{
			Type:         "choice",
			Instructions: q.Choice.Instructions,
			Criteria:     criteria,
		}, nil
	case *aiv1.EvaluateQuestion_Score:
		return &evaluateQuestion{
			Type:         "score",
			Instructions: q.Score.Instructions,
			Criteria:     q.Score.Criteria,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported question type: %T", q)
	}
}

// convertAnswer converts a TypeSafe answer to a Rill answer.
func convertAnswer(a *evaluateAnswer) (*aiv1.EvaluateAnswer, error) {
	switch a.Type {
	case "noul":
		return &aiv1.EvaluateAnswer{
			Answer: &aiv1.EvaluateAnswer_Noul{Noul: &aiv1.EvaluateNoulAnswer{
				Noul: a.Noul,
			}},
		}, nil
	case "choice":
		return &aiv1.EvaluateAnswer{
			Answer: &aiv1.EvaluateAnswer_Choice{Choice: &aiv1.EvaluateChoiceAnswer{
				Choice:        a.Choice,
				Probabilities: a.Probabilities,
				Confidence:    a.Confidence,
			}},
		}, nil
	case "score":
		return &aiv1.EvaluateAnswer{
			Answer: &aiv1.EvaluateAnswer_Score{Score: &aiv1.EvaluateScoreAnswer{
				Score:         a.Score,
				Legend:        a.Legend,
				Probabilities: a.Probabilities,
				Confidence:    a.Confidence,
			}},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported answer type %q", a.Type)
	}
}
