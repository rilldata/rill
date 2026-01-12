package server

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/go-redis/redis_rate/v10"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/observability"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	// slackEventWorkerPoolSize limits concurrent Slack event processing
	slackEventWorkerPoolSize = 50
	// slackHTTPClientTimeout is the timeout for HTTP requests to Rill API
	slackHTTPClientTimeout = 5 * time.Minute
	// slackEventProcessingTimeout is the timeout for processing a single Slack event
	slackEventProcessingTimeout = 4 * time.Minute
	// slackAPICallTimeout is the timeout for Slack API calls
	slackAPICallTimeout = 30 * time.Second
	// maxTokenLength prevents DoS via extremely long tokens
	maxTokenLength = 1000
	// slackMessageUpdateInterval prevents too frequent message updates
	slackMessageUpdateInterval = 500 * time.Millisecond
	// slackAPIMaxRetries is the maximum number of retries for Slack API calls
	slackAPIMaxRetries = 3
	// slackAPIRetryBaseDelay is the base delay for exponential backoff
	slackAPIRetryBaseDelay = 1 * time.Second
	// slackMessageMaxLength is Slack's message length limit
	slackMessageMaxLength = 4000
)

// slackEventWorkerPool manages a bounded pool of workers for processing Slack events
type slackEventWorkerPool struct {
	events chan slackEventJob
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	server *Server
}

type slackEventJob struct {
	ctx   context.Context
	event slackevents.EventsAPIEvent
}

// slackHTTPClient is a properly configured HTTP client for Rill API requests
var slackHTTPClient = &http.Client{
	Timeout: slackHTTPClientTimeout,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	},
}

// slackEventDeduplicator prevents processing duplicate Slack events
type slackEventDeduplicator struct {
	mu      sync.Mutex
	events  map[string]time.Time
	cleanup *time.Ticker
}

func newSlackEventDeduplicator() *slackEventDeduplicator {
	dedup := &slackEventDeduplicator{
		events:  make(map[string]time.Time),
		cleanup: time.NewTicker(1 * time.Hour), // Cleanup every hour
	}
	// Start cleanup goroutine
	go dedup.cleanupLoop()
	return dedup
}

func (d *slackEventDeduplicator) isProcessed(eventKey string) bool {
	if eventKey == "" {
		return false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	_, exists := d.events[eventKey]
	return exists
}

func (d *slackEventDeduplicator) markProcessed(eventKey string) {
	if eventKey == "" {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.events[eventKey] = time.Now()
}

func (d *slackEventDeduplicator) cleanupLoop() {
	for range d.cleanup.C {
		d.mu.Lock()
		cutoff := time.Now().Add(-24 * time.Hour) // Keep events for 24 hours
		for key, timestamp := range d.events {
			if timestamp.Before(cutoff) {
				delete(d.events, key)
			}
		}
		d.mu.Unlock()
	}
}

func (d *slackEventDeduplicator) shutdown() {
	if d.cleanup != nil {
		d.cleanup.Stop()
	}
}

// messageUpdateMutex tracks mutexes per message to prevent race conditions
type messageUpdateMutex struct {
	mu      sync.Mutex
	mutexes map[string]*sync.Mutex
	cleanup *time.Ticker
}

func newMessageUpdateMutex() *messageUpdateMutex {
	m := &messageUpdateMutex{
		mutexes: make(map[string]*sync.Mutex),
		cleanup: time.NewTicker(1 * time.Hour),
	}
	go m.cleanupLoop()
	return m
}

func (m *messageUpdateMutex) getMutex(messageKey string) *sync.Mutex {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mutexes[messageKey] == nil {
		m.mutexes[messageKey] = &sync.Mutex{}
	}
	return m.mutexes[messageKey]
}

func (m *messageUpdateMutex) cleanupLoop() {
	for range m.cleanup.C {
		// Cleanup is handled by messageUpdateTracker
	}
}

func (m *messageUpdateMutex) shutdown() {
	if m.cleanup != nil {
		m.cleanup.Stop()
	}
}

// messageUpdateTracker tracks and serializes updates to Slack messages
type messageUpdateTracker struct {
	mu        sync.Mutex
	lastUpdate map[string]time.Time
	cleanup   *time.Ticker
}

func newMessageUpdateTracker() *messageUpdateTracker {
	tracker := &messageUpdateTracker{
		lastUpdate: make(map[string]time.Time),
		cleanup:    time.NewTicker(1 * time.Hour), // Cleanup every hour
	}
	// Start cleanup goroutine
	go tracker.cleanupLoop()
	return tracker
}

func (m *messageUpdateTracker) cleanupLoop() {
	for range m.cleanup.C {
		m.cleanupEntries()
	}
}

func (m *messageUpdateTracker) shutdown() {
	if m.cleanup != nil {
		m.cleanup.Stop()
	}
}

func (m *messageUpdateTracker) shouldUpdate(messageKey string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	last, ok := m.lastUpdate[messageKey]
	if !ok || time.Since(last) >= slackMessageUpdateInterval {
		m.lastUpdate[messageKey] = time.Now()
		return true
	}
	return false
}

// cleanupEntries removes entries older than 1 hour to prevent memory leak
func (m *messageUpdateTracker) cleanupEntries() {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Hour)
	for key, last := range m.lastUpdate {
		if last.Before(cutoff) {
			delete(m.lastUpdate, key)
		}
	}
}

// newSlackEventWorkerPool creates a new worker pool for processing Slack events
func newSlackEventWorkerPool(s *Server) *slackEventWorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &slackEventWorkerPool{
		events: make(chan slackEventJob, 100), // Buffer up to 100 events
		ctx:    ctx,
		cancel: cancel,
		server: s,
	}

	// Start worker goroutines
	for i := 0; i < slackEventWorkerPoolSize; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool
}

func (p *slackEventWorkerPool) worker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			return
		case job := <-p.events:
			ctx, cancel := context.WithTimeout(job.ctx, slackEventProcessingTimeout)
			if err := p.server.handleSlackEvent(ctx, job.event); err != nil {
				p.server.logger.Error("failed to handle Slack event",
					zap.Error(err),
					zap.String("event_type", job.event.Type),
					observability.ZapCtx(ctx))
			}
			cancel()
		}
	}
}

func (p *slackEventWorkerPool) submit(ctx context.Context, event slackevents.EventsAPIEvent) error {
	select {
	case p.events <- slackEventJob{ctx: ctx, event: event}:
		return nil
	case <-p.ctx.Done():
		return p.ctx.Err()
	default:
		// Pool is full, log and drop event (better than blocking)
		p.server.logger.Warn("Slack event pool full, dropping event",
			zap.String("event_type", event.Type),
			observability.ZapCtx(ctx))
		return fmt.Errorf("event pool full")
	}
}

func (p *slackEventWorkerPool) shutdown() {
	p.cancel()
	close(p.events)
	p.wg.Wait()
}

// registerSlackEndpoints registers all Slack-related HTTP endpoints.
func (s *Server) registerSlackEndpoints(mux *http.ServeMux) {
	inner := http.NewServeMux()

	// Slack Events API webhook
	observability.MuxHandle(inner, "/slack/events", http.HandlerFunc(s.slackWebhook))

	// Slack slash command for setting token
	observability.MuxHandle(inner, "/slack/commands/set-token", http.HandlerFunc(s.slackSlashCommand))

	mux.Handle("/slack/", observability.Middleware("admin", s.logger, inner))
}

// slackWebhook handles Slack Events API webhook requests.
func (s *Server) slackWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "expected a POST request", http.StatusBadRequest)
		return
	}

	// Verify Slack signature
	verifier, err := slack.NewSecretsVerifier(r.Header, s.opts.SlackSigningSecret)
	if err != nil {
		s.logger.Error("failed to create Slack verifier", zap.Error(err))
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	bodyReader := io.TeeReader(r.Body, &verifier)
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		s.logger.Error("failed to read request body", zap.Error(err))
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	if err := verifier.Ensure(); err != nil {
		s.logger.Error("failed to verify Slack signature", zap.Error(err))
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// Handle URL verification challenge
	var challenge struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
		Token     string `json:"token"`
		TeamID    string `json:"team_id"`
	}
	if err := json.Unmarshal(body, &challenge); err == nil {
		if challenge.Type == "url_verification" {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(challenge.Challenge))
			return
		}
	}

	// Parse event
	event, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		s.logger.Error("failed to parse Slack event", zap.Error(err))
		http.Error(w, "failed to parse event", http.StatusBadRequest)
		return
	}

	// Deduplicate events using event timestamp (Slack can retry)
	// Slack Events API wraps callback events with EventID and EventTime
	// We need to extract these from the raw JSON since EventsAPIEvent doesn't expose them
	var eventKey string
	var rawEvent struct {
		EventID    string `json:"event_id"`
		EventTime  int    `json:"event_time"`
		TeamID     string `json:"team_id"`
	}
	if err := json.Unmarshal(body, &rawEvent); err == nil {
		if rawEvent.EventID != "" {
			eventKey = fmt.Sprintf("%s:%s", rawEvent.TeamID, rawEvent.EventID)
		} else if rawEvent.EventTime > 0 {
			// Fallback to timestamp if EventID not available
			eventKey = fmt.Sprintf("%s:%d", rawEvent.TeamID, rawEvent.EventTime)
		}
	}
	
	// If no EventID or EventTime, use content hash as fallback
	if eventKey == "" {
		hash := sha256.Sum256(body)
		eventKey = fmt.Sprintf("%s:%s", event.TeamID, hex.EncodeToString(hash[:16])) // Use first 16 bytes
	}
	
	if s.slackDedup.isProcessed(eventKey) {
		s.logger.Debug("duplicate Slack event ignored",
			zap.String("event_key", eventKey),
			zap.String("event_type", event.Type),
			observability.ZapCtx(r.Context()))
		w.WriteHeader(http.StatusOK)
		return
	}
	s.slackDedup.markProcessed(eventKey)

	// Submit event to worker pool (non-blocking)
	// Use request context to preserve trace IDs and request metadata
	reqCtx := r.Context()
	if err := s.slackPool.submit(reqCtx, event); err != nil {
		s.logger.Error("failed to submit Slack event to worker pool",
			zap.Error(err),
			zap.String("event_type", event.Type),
			observability.ZapCtx(reqCtx))
		// Still return 200 to Slack to avoid retries
	}

	w.WriteHeader(http.StatusOK)
}

// handleSlackEvent processes a Slack event.
func (s *Server) handleSlackEvent(ctx context.Context, event slackevents.EventsAPIEvent) error {
	switch event.Type {
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			return s.handleAppMention(ctx, ev, event.TeamID)
		case *slackevents.MessageEvent:
			// Only handle DMs and thread replies (not regular channel messages)
			if ev.ChannelType == "im" || ev.ThreadTimeStamp != "" {
				return s.handleMessage(ctx, ev, event.TeamID)
			}
		}
	}
	return nil
}

// handleAppMention handles when the bot is mentioned in a channel.
func (s *Server) handleAppMention(ctx context.Context, event *slackevents.AppMentionEvent, teamID string) error {
	// Rate limit per user
	rateLimitKey := fmt.Sprintf("slack:user:%s:%s", teamID, event.User)
	if err := s.limiter.Limit(ctx, rateLimitKey, redis_rate.PerMinute(10)); err != nil {
		s.logger.Warn("rate limit exceeded for Slack user",
			zap.String("user_id", event.User),
			zap.String("team_id", teamID),
			observability.ZapCtx(ctx))
		// Try to send rate limit message, but don't fail if it errors (Slack already got 200)
		if sendErr := s.sendSlackMessageWithRetry(ctx, teamID, event.Channel, nil, "Rate limit exceeded. Please wait a moment before sending another message."); sendErr != nil {
			s.logger.Warn("failed to send rate limit message",
				zap.Error(sendErr),
				observability.ZapCtx(ctx))
		}
		return nil // Don't return error - rate limit is expected behavior
	}

	// Register workspace if needed (with retry for race conditions)
	workspace, err := s.admin.DB.FindSlackWorkspace(ctx, teamID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			workspace, err = s.admin.DB.InsertSlackWorkspace(ctx, &database.InsertSlackWorkspaceOptions{
				TeamID: teamID,
			})
			if err != nil {
				// Might be a race condition, try to fetch again
				if errors.Is(err, database.ErrNotUnique) {
					workspace, err = s.admin.DB.FindSlackWorkspace(ctx, teamID)
					if err != nil {
						return fmt.Errorf("failed to register workspace: %w", err)
					}
				} else {
					return fmt.Errorf("failed to register workspace: %w", err)
				}
			}
		} else {
			return fmt.Errorf("failed to find workspace: %w", err)
		}
	}

	// Extract text (remove bot mention)
	text := strings.TrimSpace(strings.ReplaceAll(event.Text, fmt.Sprintf("<@%s>", event.BotID), ""))

	// Use thread timestamp if this is a reply, otherwise use event timestamp to create new thread
	threadTS := event.ThreadTimeStamp
	if threadTS == "" {
		threadTS = event.TimeStamp
	}

	return s.processSlackMessage(ctx, workspace.ID, event.User, event.Channel, text, threadTS, teamID)
}

// handleMessage handles direct messages and thread replies.
func (s *Server) handleMessage(ctx context.Context, event *slackevents.MessageEvent, teamID string) error {
	// Skip bot messages
	if event.BotID != "" {
		return nil
	}

	// Skip messages with subtypes (like channel_join, etc.)
	if event.SubType != "" {
		return nil
	}

	// Rate limit per user
	rateLimitKey := fmt.Sprintf("slack:user:%s:%s", teamID, event.User)
	if err := s.limiter.Limit(ctx, rateLimitKey, redis_rate.PerMinute(10)); err != nil {
		s.logger.Warn("rate limit exceeded for Slack user",
			zap.String("user_id", event.User),
			zap.String("team_id", teamID),
			observability.ZapCtx(ctx))
		// Try to send rate limit message, but don't fail if it errors (Slack already got 200)
		if sendErr := s.sendSlackMessageWithRetry(ctx, teamID, event.Channel, nil, "Rate limit exceeded. Please wait a moment before sending another message."); sendErr != nil {
			s.logger.Warn("failed to send rate limit message",
				zap.Error(sendErr),
				observability.ZapCtx(ctx))
		}
		return nil // Don't return error - rate limit is expected behavior
	}

	// Register workspace if needed (with retry for race conditions)
	workspace, err := s.admin.DB.FindSlackWorkspace(ctx, teamID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			workspace, err = s.admin.DB.InsertSlackWorkspace(ctx, &database.InsertSlackWorkspaceOptions{
				TeamID: teamID,
			})
			if err != nil {
				// Might be a race condition, try to fetch again
				if errors.Is(err, database.ErrNotUnique) {
					workspace, err = s.admin.DB.FindSlackWorkspace(ctx, teamID)
					if err != nil {
						return fmt.Errorf("failed to register workspace: %w", err)
					}
				} else {
					return fmt.Errorf("failed to register workspace: %w", err)
				}
			}
		} else {
			return fmt.Errorf("failed to find workspace: %w", err)
		}
	}

	text := strings.TrimSpace(event.Text)
	if text == "" {
		return nil
	}

	// Check if it's a token (starts with rill_ and no spaces)
	if strings.HasPrefix(text, "rill_") && !strings.Contains(text, " ") {
		// Validate token length to prevent DoS
		if len(text) > maxTokenLength {
			return s.sendSlackMessageWithRetry(ctx, teamID, event.Channel, nil, fmt.Sprintf("Error: Token is too long (max %d characters)", maxTokenLength))
		}
		return s.handleSetToken(ctx, workspace.ID, event.User, text, event.Channel, nil, teamID)
	}

	// Use thread timestamp if available, otherwise use event timestamp
	threadTS := event.ThreadTimeStamp
	if threadTS == "" && event.ChannelType == "im" {
		threadTS = event.TimeStamp
	}

	return s.processSlackMessage(ctx, workspace.ID, event.User, event.Channel, text, threadTS, teamID)
}

// processSlackMessage processes a message and streams the Rill response.
func (s *Server) processSlackMessage(ctx context.Context, workspaceID, userID, channelID, text, threadTS, teamID string) error {
	// Get user token with decrypted secret
	token, err := s.admin.DB.FindSlackUserTokenWithSecret(ctx, workspaceID, userID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return s.promptForToken(ctx, teamID, userID, channelID, threadTS)
		}
		return fmt.Errorf("failed to get token: %w", err)
	}

	// Get or create conversation
	var conversationID string
	var threadTSPtr *string
	if threadTS != "" {
		threadTSPtr = &threadTS
	}

	conv, err := s.admin.DB.FindSlackConversation(ctx, workspaceID, channelID, threadTSPtr)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return fmt.Errorf("failed to find conversation: %w", err)
	}

	if conv != nil {
		conversationID = conv.RillConversationID
	}

	// Get org/project from config, or fall back to first available
	org := s.opts.SlackDefaultOrg
	project := s.opts.SlackDefaultProject

	// If not configured, find the first org and project
	if org == "" || project == "" {
		orgs, err := s.admin.DB.FindOrganizations(ctx, "", 1)
		if err != nil || len(orgs) == 0 {
			return s.sendSlackMessageWithRetry(ctx, teamID, channelID, threadTSPtr, "Error: No organizations found. Please configure RILL_ADMIN_SLACK_DEFAULT_ORG and RILL_ADMIN_SLACK_DEFAULT_PROJECT, or create an organization.")
		}
		org = orgs[0].Name

		projects, err := s.admin.DB.FindProjectsForOrganization(ctx, orgs[0].ID, "", 1)
		if err != nil || len(projects) == 0 {
			return s.sendSlackMessageWithRetry(ctx, teamID, channelID, threadTSPtr, fmt.Sprintf("Error: No projects found in organization %s. Please configure RILL_ADMIN_SLACK_DEFAULT_PROJECT, or create a project.", org))
		}
		project = projects[0].Name
	}

	// Stream Rill response using runtime proxy
	return s.streamRillResponseToSlack(ctx, teamID, workspaceID, userID, channelID, threadTSPtr, org, project, token.Token, conversationID, text)
}

// streamRillResponseToSlack streams a Rill chat response to Slack.
func (s *Server) streamRillResponseToSlack(ctx context.Context, teamID, workspaceID, userID, channelID string, threadTS *string, org, project, rillToken, conversationID, prompt string) error {
	// Find project and deployment to get actual instance ID
	proj, err := s.admin.DB.FindProjectByName(ctx, org, project)
	if err != nil {
		s.logger.Error("failed to find project",
			zap.Error(err),
			zap.String("org", org),
			zap.String("project", project),
			observability.ZapCtx(ctx))
		return s.sendSlackMessageWithRetry(ctx, teamID, channelID, threadTS, fmt.Sprintf("Error: Project %s/%s not found", org, project))
	}

	if proj.PrimaryDeploymentID == nil {
		return s.sendSlackMessageWithRetry(ctx, teamID, channelID, threadTS, "Error: No deployment found for project")
	}

	depl, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
	if err != nil {
		s.logger.Error("failed to find deployment",
			zap.Error(err),
			zap.String("deployment_id", *proj.PrimaryDeploymentID),
			observability.ZapCtx(ctx))
		return s.sendSlackMessageWithRetry(ctx, teamID, channelID, threadTS, "Error: Deployment not found")
	}

	// Note: Deployment status checking would go here if we had a status field
	// For now, we'll let the API call fail and handle the error

	// Use the runtime proxy URL with actual instance ID
	apiURL := fmt.Sprintf("%s/v1/organizations/%s/projects/%s/runtime/v1/instances/%s/ai/complete/stream",
		s.admin.URLs.External(), org, project, depl.RuntimeInstanceID)

	// Build request body
	reqBody := &runtimev1.CompleteStreamingRequest{
		InstanceId:     depl.RuntimeInstanceID,
		ConversationId: conversationID,
		Prompt:         prompt,
		Agent:          "analyst_agent",
	}

	bodyBytes, err := protojson.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+rillToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	// Make request using proper HTTP client with timeouts
	resp, err := slackHTTPClient.Do(req)
	if err != nil {
		// Sanitize error message - don't expose internal details
		s.logger.Error("failed to connect to Rill API",
			zap.Error(err),
			zap.String("org", org),
			zap.String("project", project),
			observability.ZapCtx(ctx))
		return s.sendSlackMessageWithRetry(ctx, teamID, channelID, threadTS, "Error: Failed to connect to Rill API. Please try again later.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read and discard body to allow connection reuse
		_, _ = io.Copy(io.Discard, resp.Body)
		// Sanitize error message - don't expose response body which might contain sensitive info
		s.logger.Error("Rill API returned error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("org", org),
			zap.String("project", project),
			observability.ZapCtx(ctx))

		// Provide more helpful error messages for common status codes
		var errorMsg string
		switch resp.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			errorMsg = "Error: Authentication failed. Your token may be invalid or expired. Please update your token with `/set-token`."
		case http.StatusNotFound:
			errorMsg = "Error: Resource not found. Please check that your project and deployment are configured correctly."
		case http.StatusServiceUnavailable:
			errorMsg = "Error: Service temporarily unavailable. Please try again in a moment."
		default:
			errorMsg = fmt.Sprintf("Error: Rill API returned status %d. Please check your project configuration.", resp.StatusCode)
		}
		return s.sendSlackMessageWithRetry(ctx, teamID, channelID, threadTS, errorMsg)
	}

	// Send initial "thinking" message with timeout and retry
	slackClient := slack.New(s.opts.SlackBotToken)
	opts := []slack.MsgOption{
		slack.MsgOptionText("Thinking...", false),
	}
	if threadTS != nil && *threadTS != "" {
		opts = append(opts, slack.MsgOptionTS(*threadTS))
	}

	postCtx, postCancel := context.WithTimeout(ctx, slackAPICallTimeout)
	defer postCancel()

	var timestamp string
	var postErr error
	delay := slackAPIRetryBaseDelay
	for attempt := 0; attempt < slackAPIMaxRetries; attempt++ {
		var ts string
		_, ts, postErr = slackClient.PostMessageContext(postCtx, channelID, opts...)
		if postErr == nil {
			timestamp = ts
			break
		}

		// Handle rate limits - check for *RateLimitedError
		var retryAfter time.Duration
		var rateLimitErr *slack.RateLimitedError
		if errors.As(postErr, &rateLimitErr) {
			retryAfter = rateLimitErr.RetryAfter
		}
		if retryAfter > 0 {
			delay = retryAfter
			s.logger.Warn("Slack API rate limited, retrying",
				zap.Int("attempt", attempt+1),
				zap.Duration("retry_after", delay),
				zap.String("channel", channelID),
				observability.ZapCtx(ctx))
			select {
			case <-postCtx.Done():
				return postCtx.Err()
			case <-time.After(delay):
				continue
			}
		}

		// For other errors, exponential backoff
		if attempt < slackAPIMaxRetries-1 {
			delay = slackAPIRetryBaseDelay * time.Duration(1<<attempt)
			select {
			case <-postCtx.Done():
				return postCtx.Err()
			case <-time.After(delay):
				continue
			}
		}
	}

	if postErr != nil {
		s.logger.Error("failed to post initial Slack message after retries",
			zap.Error(postErr),
			zap.String("channel", channelID),
			observability.ZapCtx(ctx))
		return fmt.Errorf("failed to post initial message: %w", postErr)
	}

	if timestamp == "" {
		return fmt.Errorf("failed to get message timestamp")
	}

	// Parse SSE stream with proper handling for multi-line data
	var fullResponseBuilder strings.Builder
	var finalConversationID string
	scanner := bufio.NewScanner(resp.Body)
	// Increase buffer size for large responses (default is 64KB)
	buf := make([]byte, 0, 256*1024) // 256KB buffer
	scanner.Buffer(buf, 1024*1024)    // Allow up to 1MB

	var currentEvent struct {
		Type string
		Data strings.Builder
	}
	messageKey := fmt.Sprintf("%s:%s", channelID, timestamp)

	// Track if we need to clean up the "Thinking..." message on cancellation or error
	cleanupMessage := true
	cancelled := false
	defer func() {
		if cleanupMessage {
			// Try to update or delete the message
			if cancelled || ctx.Err() != nil {
				// Context was cancelled, try to delete the "Thinking..." message
				deleteCtx, deleteCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer deleteCancel()
				_, _, _ = slackClient.DeleteMessageContext(deleteCtx, channelID, timestamp)
			} else if fullResponseBuilder.Len() == 0 {
				// Empty response, delete the "Thinking..." message
				deleteCtx, deleteCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer deleteCancel()
				_, _, _ = slackClient.DeleteMessageContext(deleteCtx, channelID, timestamp)
			}
		}
	}()

	for scanner.Scan() {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			cleanupMessage = true
			cancelled = true
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
		}

		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines (event boundaries)
		if trimmed == "" {
			// Process complete event
			if currentEvent.Data.Len() > 0 {
				dataStr := currentEvent.Data.String()
				if dataStr == "[DONE]" {
					break
				}

				// Handle empty data (SSE spec allows empty data: lines)
				// Empty data means the event has no payload, just reset
				if dataStr == "" {
					currentEvent = struct {
						Type string
						Data strings.Builder
					}{}
					continue
				}

				var response runtimev1.CompleteStreamingResponse
				if err := protojson.Unmarshal([]byte(dataStr), &response); err != nil {
					s.logger.Warn("failed to unmarshal SSE response",
						zap.Error(err),
						zap.String("data_preview", truncateUTF8(dataStr, 100)),
						zap.String("event_type", currentEvent.Type),
						observability.ZapCtx(ctx))
					// Continue processing other events - don't fail on malformed JSON
					currentEvent = struct {
						Type string
						Data strings.Builder
					}{}
					continue
				} else {
					if response.ConversationId != "" {
						finalConversationID = response.ConversationId
					}

					// Extract text from message
					if response.Message != nil && len(response.Message.Content) > 0 {
						for _, block := range response.Message.Content {
							if text := block.GetText(); text != "" {
								if fullResponseBuilder.Len() > 0 {
									fullResponseBuilder.WriteString("\n")
								}
								fullResponseBuilder.WriteString(text)
							}
						}
					}

					// Update message as we receive chunks (with rate limiting and mutex to prevent races)
					fullResponse := fullResponseBuilder.String()
					if fullResponse != "" && s.slackTracker.shouldUpdate(messageKey) {
						// Use mutex per message to prevent concurrent updates
						msgMutex := s.slackMutex.getMutex(messageKey)
						msgMutex.Lock()
						defer msgMutex.Unlock()

						// Truncate long messages using UTF-8 aware truncation
						responseText := truncateUTF8(fullResponse, slackMessageMaxLength)

						if err := s.updateSlackMessageWithRetry(ctx, slackClient, channelID, timestamp, responseText); err != nil {
							s.logger.Warn("failed to update Slack message",
								zap.Error(err),
								zap.String("channel", channelID),
								observability.ZapCtx(ctx))
							// Continue processing - don't fail the whole stream
						}
					}
				}
			}
			// Reset for next event
			currentEvent = struct {
				Type string
				Data strings.Builder
			}{}
			continue
		}

		// Skip comments
		if strings.HasPrefix(trimmed, ":") {
			continue
		}

		if strings.HasPrefix(trimmed, "event:") {
			currentEvent.Type = strings.TrimSpace(trimmed[6:])
		} else if strings.HasPrefix(trimmed, "data:") {
			// Handle multi-line data fields (SSE spec allows multiple data: lines)
			// Note: SSE spec says empty data: line should be treated as data with empty string
			data := strings.TrimSpace(trimmed[5:])
			if currentEvent.Data.Len() > 0 {
				currentEvent.Data.WriteString("\n")
			}
			currentEvent.Data.WriteString(data)
		} else if strings.HasPrefix(trimmed, "id:") {
			// SSE can have id: field for event IDs (we ignore but parse correctly)
			// This ensures we don't break on valid SSE extensions
		} else if strings.HasPrefix(trimmed, "retry:") {
			// SSE can have retry: field for reconnection timeout (we ignore)
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return fmt.Errorf("failed to read SSE stream: %w", err)
	}

	// Check if we got an empty response
	fullResponse := fullResponseBuilder.String()
	if fullResponse == "" {
		cleanupMessage = true // Delete "Thinking..." if no response
		return s.sendSlackMessageWithRetry(ctx, teamID, channelID, threadTS, "Sorry, I didn't receive a response. Please try again.")
	}

	cleanupMessage = false // Don't cleanup if we completed successfully

	// Save/update conversation with proper error handling
	if finalConversationID != "" {
		// Check if conversation belongs to this user (prevent cross-user conversation hijacking)
		if conv, err := s.admin.DB.FindSlackConversation(ctx, workspaceID, channelID, threadTS); err == nil {
			// Verify user matches (for thread safety)
			if conv.UserID != userID {
				s.logger.Warn("conversation user mismatch",
					zap.String("conversation_user", conv.UserID),
					zap.String("request_user", userID),
					zap.String("workspace_id", workspaceID),
					zap.String("channel_id", channelID),
					observability.ZapCtx(ctx))
				// Create a new conversation for this user instead
				if _, err := s.admin.DB.InsertSlackConversation(ctx, &database.InsertSlackConversationOptions{
					WorkspaceID:        workspaceID,
					ChannelID:          channelID,
					ThreadTS:           threadTS,
					RillConversationID: finalConversationID,
					UserID:             userID,
				}); err != nil {
					s.logger.Error("failed to insert Slack conversation for user",
						zap.Error(err),
						zap.String("workspace_id", workspaceID),
						zap.String("channel_id", channelID),
						observability.ZapCtx(ctx))
				}
			} else {
				// Update existing
				if err := s.admin.DB.UpdateSlackConversation(ctx, conv.ID, &database.UpdateSlackConversationOptions{
					RillConversationID: finalConversationID,
				}); err != nil {
					s.logger.Error("failed to update Slack conversation",
						zap.Error(err),
						zap.String("conversation_id", conv.ID),
						observability.ZapCtx(ctx))
					// Don't return error - conversation state is best-effort
				}
			}
		} else if errors.Is(err, database.ErrNotFound) {
			// Create new
			if _, err := s.admin.DB.InsertSlackConversation(ctx, &database.InsertSlackConversationOptions{
				WorkspaceID:        workspaceID,
				ChannelID:          channelID,
				ThreadTS:           threadTS,
				RillConversationID: finalConversationID,
				UserID:             userID,
			}); err != nil {
				s.logger.Error("failed to insert Slack conversation",
					zap.Error(err),
					zap.String("workspace_id", workspaceID),
					zap.String("channel_id", channelID),
					observability.ZapCtx(ctx))
				// Don't return error - conversation state is best-effort
			}
		} else {
			s.logger.Error("failed to find Slack conversation",
				zap.Error(err),
				zap.String("workspace_id", workspaceID),
				zap.String("channel_id", channelID),
				observability.ZapCtx(ctx))
		}
	}

	// Final message update with mutex protection
	msgMutex := s.slackMutex.getMutex(messageKey)
	msgMutex.Lock()
	defer msgMutex.Unlock()

	// Truncate using UTF-8 aware truncation
	responseText := truncateUTF8(fullResponse, slackMessageMaxLength)

	if err := s.updateSlackMessageWithRetry(ctx, slackClient, channelID, timestamp, responseText); err != nil {
		s.logger.Error("failed to update final Slack message",
			zap.Error(err),
			zap.String("channel", channelID),
			observability.ZapCtx(ctx))
		return fmt.Errorf("failed to update final message: %w", err)
	}

	return nil
}

// handleSetToken handles setting a user's Rill token.
func (s *Server) handleSetToken(ctx context.Context, workspaceID, userID, token, channelID string, threadTS *string, teamID string) error {
	// Validate token format and length
	if !strings.HasPrefix(token, "rill_") {
		return s.sendSlackMessageWithRetry(ctx, teamID, channelID, threadTS, "Error: Invalid token format. Rill tokens should start with 'rill_'")
	}
	if len(token) > maxTokenLength {
		return s.sendSlackMessageWithRetry(ctx, teamID, channelID, threadTS, fmt.Sprintf("Error: Token is too long (max %d characters)", maxTokenLength))
	}
	if len(token) < 10 {
		return s.sendSlackMessageWithRetry(ctx, teamID, channelID, threadTS, "Error: Token is too short")
	}

	// Validate token by making a test API call
	// Use the configured or first available org/project to test the token
	org := s.opts.SlackDefaultOrg
	project := s.opts.SlackDefaultProject

	// If not configured, find the first org and project
	if org == "" || project == "" {
		orgs, err := s.admin.DB.FindOrganizations(ctx, "", 1)
		if err == nil && len(orgs) > 0 {
			org = orgs[0].Name
			projects, err := s.admin.DB.FindProjectsForOrganization(ctx, orgs[0].ID, "", 1)
			if err == nil && len(projects) > 0 {
				project = projects[0].Name
			}
		}
	}

	if org != "" && project != "" {
		proj, err := s.admin.DB.FindProjectByName(ctx, org, project)
		if err == nil && proj.PrimaryDeploymentID != nil {
			depl, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
			if err == nil {
				// Make a lightweight test call to validate the token
				testURL := fmt.Sprintf("%s/v1/instances/%s/ping", depl.RuntimeHost, depl.RuntimeInstanceID)
				testReq, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
				if err == nil {
					testReq.Header.Set("Authorization", "Bearer "+token)
					testResp, err := slackHTTPClient.Do(testReq)
					if err == nil {
						testResp.Body.Close()
						if testResp.StatusCode == http.StatusUnauthorized || testResp.StatusCode == http.StatusForbidden {
							return s.sendSlackMessageWithRetry(ctx, teamID, channelID, threadTS, "Error: Token is invalid or expired. Please generate a new token with `rill token issue`.")
						}
					}
				}
			}
		}
	}

	// Save token (database will encrypt it)
	_, err := s.admin.DB.InsertSlackUserToken(ctx, &database.InsertSlackUserTokenOptions{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Token:       token, // Plain token - database encrypts it
	})
	if err != nil {
		s.logger.Error("failed to save Slack user token",
			zap.Error(err),
			zap.String("workspace_id", workspaceID),
			zap.String("user_id", userID),
			observability.ZapCtx(ctx))
		return fmt.Errorf("failed to save token: %w", err)
	}

	return s.sendSlackMessageWithRetry(ctx, teamID, channelID, threadTS, "✅ Token saved successfully! You can now ask me questions about your Rill data.")
}

// promptForToken sends a message prompting the user to set their token.
func (s *Server) promptForToken(ctx context.Context, teamID, userID, channelID, threadTS string) error {
	message := "Hi! I need your Rill personal access token to interact with Rill Cloud.\n\n" +
		"To get your token:\n" +
		"1. Run `rill token issue --display-name \"Slack Bot\"` in your terminal\n" +
		"2. Send me your token in a DM with: `/set-token <your-token>`\n\n" +
		"Or you can set it directly by sending me a DM with just your token.\n\n" +
		"Your token will be stored securely and only used to authenticate your requests to Rill Cloud."

	return s.sendSlackMessageWithRetry(ctx, teamID, channelID, &threadTS, message)
}

// sendSlackMessageWithRetry sends a message to Slack with retry logic and error handling.
func (s *Server) sendSlackMessageWithRetry(ctx context.Context, teamID, channelID string, threadTS *string, text string) error {
	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, slackAPICallTimeout)
	defer cancel()

	client := slack.New(s.opts.SlackBotToken)
	opts := []slack.MsgOption{
		slack.MsgOptionText(text, false),
	}
	if threadTS != nil && *threadTS != "" {
		opts = append(opts, slack.MsgOptionTS(*threadTS))
	}

	var lastErr error
	delay := slackAPIRetryBaseDelay

	for attempt := 0; attempt < slackAPIMaxRetries; attempt++ {
		_, _, err := client.PostMessageContext(ctx, channelID, opts...)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if it's a rate limit error - slack-go returns *RateLimitedError
		var retryAfter time.Duration
		var rateLimitErr *slack.RateLimitedError
		if errors.As(err, &rateLimitErr) {
			retryAfter = rateLimitErr.RetryAfter
		}
		if retryAfter > 0 {
			delay = retryAfter
			s.logger.Warn("Slack API rate limited, retrying",
				zap.Int("attempt", attempt+1),
				zap.Duration("retry_after", delay),
				zap.String("channel", channelID),
				observability.ZapCtx(ctx))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				continue
			}
		}

		// Handle specific Slack errors
		errStr := err.Error()
		if strings.Contains(errStr, "channel_not_found") || strings.Contains(errStr, "not_in_channel") {
			s.logger.Warn("Slack channel error",
				zap.String("error", errStr),
				zap.String("channel", channelID),
				observability.ZapCtx(ctx))
			// Don't retry - this is a permanent error
			return fmt.Errorf("channel error: %s", errStr)
		}

		if strings.Contains(errStr, "invalid_auth") || strings.Contains(errStr, "account_inactive") {
			s.logger.Error("Slack authentication error",
				zap.String("error", errStr),
				observability.ZapCtx(ctx))
			// Don't retry - this is a configuration error
			return fmt.Errorf("authentication error: %s", errStr)
		}

		// For other errors, use exponential backoff
		if attempt < slackAPIMaxRetries-1 {
			delay = slackAPIRetryBaseDelay * time.Duration(1<<attempt)
			s.logger.Warn("Slack API call failed, retrying",
				zap.Int("attempt", attempt+1),
				zap.Duration("retry_after", delay),
				zap.Error(err),
				observability.ZapCtx(ctx))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				continue
			}
		}
	}

	s.logger.Error("Slack API call failed after retries",
		zap.Int("attempts", slackAPIMaxRetries),
		zap.Error(lastErr),
		observability.ZapCtx(ctx))
	return fmt.Errorf("failed after %d attempts: %w", slackAPIMaxRetries, lastErr)
}

// sendSlackMessage is a convenience wrapper that uses retry logic.
func (s *Server) sendSlackMessage(ctx context.Context, teamID, channelID string, threadTS *string, text string) error {
	return s.sendSlackMessageWithRetry(ctx, teamID, channelID, threadTS, text)
}

// updateSlackMessageWithRetry updates a Slack message with retry logic.
func (s *Server) updateSlackMessageWithRetry(ctx context.Context, client *slack.Client, channelID, timestamp, text string) error {
	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, slackAPICallTimeout)
	defer cancel()

	var lastErr error
	delay := slackAPIRetryBaseDelay

	for attempt := 0; attempt < slackAPIMaxRetries; attempt++ {
		_, _, _, err := client.UpdateMessage(channelID, timestamp, slack.MsgOptionText(text, false))
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if it's a rate limit error - use errors.As for pointer type
		var retryAfter time.Duration
		var rateLimitErr *slack.RateLimitedError
		if errors.As(err, &rateLimitErr) {
			retryAfter = rateLimitErr.RetryAfter
		}
		if retryAfter > 0 {
			delay = retryAfter
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				continue
			}
		}

		// For channel/message errors, don't retry
		errStr := err.Error()
		if strings.Contains(errStr, "channel_not_found") || strings.Contains(errStr, "message_not_found") {
			return fmt.Errorf("message error: %s", errStr)
		}

		// Exponential backoff for other errors
		if attempt < slackAPIMaxRetries-1 {
			delay = slackAPIRetryBaseDelay * time.Duration(1<<attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				continue
			}
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", slackAPIMaxRetries, lastErr)
}

// truncateUTF8 safely truncates a string to maxLen runes (not bytes) to avoid cutting UTF-8 characters.
func truncateUTF8(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	// Truncate to maxLen-3 to leave room for "..."
	truncated := string(runes[:maxLen-3])
	// Ensure the result is valid UTF-8
	if !utf8.ValidString(truncated) {
		// If invalid, find the last valid UTF-8 boundary
		for len(truncated) > 0 && !utf8.ValidString(truncated) {
			truncated = truncated[:len(truncated)-1]
		}
	}
	return truncated + "..."
}

// slackSlashCommand handles the /set-token slash command.
func (s *Server) slackSlashCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "expected a POST request", http.StatusBadRequest)
		return
	}

	// Verify request came from Slack
	verifier, err := slack.NewSecretsVerifier(r.Header, s.opts.SlackSigningSecret)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	bodyReader := io.TeeReader(r.Body, &verifier)
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	if err := verifier.Ensure(); err != nil {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse command from form data
	// Slack sends slash commands as application/x-www-form-urlencoded
	r.Body = io.NopCloser(strings.NewReader(string(body)))
	if err := r.ParseForm(); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	command := slack.SlashCommand{
		Token:       r.FormValue("token"),
		TeamID:      r.FormValue("team_id"),
		TeamDomain:  r.FormValue("team_domain"),
		ChannelID:   r.FormValue("channel_id"),
		ChannelName: r.FormValue("channel_name"),
		UserID:      r.FormValue("user_id"),
		UserName:    r.FormValue("user_name"),
		Command:     r.FormValue("command"),
		Text:        r.FormValue("text"),
		ResponseURL: r.FormValue("response_url"),
	}

	if command.Command != "/set-token" {
		http.Error(w, "unknown command", http.StatusBadRequest)
		return
	}

	token := strings.TrimSpace(command.Text)
	if token == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"response_type": "ephemeral",
			"text":          "Usage: /set-token <your-rill-token>",
		})
		return
	}

	// Validate token format before processing
	if !strings.HasPrefix(token, "rill_") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"response_type": "ephemeral",
			"text":          "Error: Invalid token format. Rill tokens should start with 'rill_'",
		})
		return
	}
	if len(token) > maxTokenLength {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"response_type": "ephemeral",
			"text":          fmt.Sprintf("Error: Token is too long (max %d characters)", maxTokenLength),
		})
		return
	}

	// Get or create workspace
	workspace, err := s.admin.DB.FindSlackWorkspace(r.Context(), command.TeamID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			workspace, err = s.admin.DB.InsertSlackWorkspace(r.Context(), &database.InsertSlackWorkspaceOptions{
				TeamID: command.TeamID,
			})
			if err != nil {
				if errors.Is(err, database.ErrNotUnique) {
					// Race condition, try to fetch again
					workspace, err = s.admin.DB.FindSlackWorkspace(r.Context(), command.TeamID)
					if err != nil {
						http.Error(w, "failed to register workspace", http.StatusInternalServerError)
						return
					}
				} else {
					http.Error(w, "failed to register workspace", http.StatusInternalServerError)
					return
				}
			}
		} else {
			http.Error(w, "failed to find workspace", http.StatusInternalServerError)
			return
		}
	}

	// Handle setting token (this will validate and save)
	err = s.handleSetToken(r.Context(), workspace.ID, command.UserID, token, command.ChannelID, nil, command.TeamID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		// Extract user-friendly error message
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "invalid") || strings.Contains(errorMsg, "expired") {
			errorMsg = "Token is invalid or expired. Please generate a new token with `rill token issue`."
		}
		json.NewEncoder(w).Encode(map[string]string{
			"response_type": "ephemeral",
			"text":          fmt.Sprintf("Error: %s", errorMsg),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"response_type": "ephemeral",
		"text":          "✅ Token saved successfully!",
	})
}

