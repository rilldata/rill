# Code Review: Slack Bot Implementation

## 🔴 CRITICAL ISSUES

### 1. **Fire-and-Forget Goroutines (Line 90)**
```go
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    // ...
}()
```
**This is a production disaster waiting to happen:**
- No goroutine pool or semaphore to limit concurrency
- No way to track or cancel these goroutines
- If Slack sends 1000 events, you spawn 1000 goroutines
- Memory leak: goroutines can accumulate if processing is slow
- **Fix:** Use a worker pool with bounded concurrency (e.g., `errgroup` with limit, or a channel-based worker pool)

### 2. **http.DefaultClient with No Timeout (Line 263)**
```go
resp, err := http.DefaultClient.Do(req)
```
**This will hang forever:**
- `http.DefaultClient` has NO timeout
- If the runtime API is slow/down, this blocks indefinitely
- Can exhaust file descriptors
- **Fix:** Create a proper HTTP client with timeouts:
```go
client := &http.Client{
    Timeout: 5 * time.Minute,
    Transport: &http.Transport{
        MaxIdleConns: 100,
        IdleConnTimeout: 90 * time.Second,
    },
}
```

### 3. **Silent Error Swallowing (Lines 367, 372)**
```go
_ = s.admin.DB.UpdateSlackConversation(ctx, conv.ID, ...)
_, _ = s.admin.DB.InsertSlackConversation(ctx, ...)
```
**Database errors are silently ignored:**
- If conversation state fails to save, users lose context
- No way to debug why conversations break
- **Fix:** At minimum, log errors. Better: return them or use a background job to retry.

### 4. **Race Condition in Message Updates (Line 328)**
```go
_, _, _, err = slackClient.UpdateMessage(channelID, timestamp, ...)
```
**Multiple goroutines can update the same message concurrently:**
- If SSE stream sends multiple events quickly, you get race conditions
- Last write wins, but you might lose intermediate updates
- **Fix:** Use a mutex or channel to serialize updates per message, or batch updates.

### 5. **No Rate Limiting**
**Users can spam the bot:**
- No per-user rate limiting
- No per-workspace rate limiting
- Can exhaust API quotas, database connections, etc.
- **Fix:** Add rate limiting middleware (you already have `limiter` in the server, use it!)

### 6. **Context Loss (Line 91)**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
```
**Using `context.Background()` loses all request context:**
- No trace IDs, no request IDs, no user context
- Can't correlate logs with original request
- **Fix:** Derive from request context: `context.WithTimeout(r.Context(), ...)`

## 🟠 MAJOR ISSUES

### 7. **Fragile SSE Parsing (Lines 296-357)**
**Your SSE parser is naive and will break:**
- Doesn't handle multi-line data fields properly (SSE spec allows `data:` on multiple lines)
- Doesn't handle retries or reconnection
- No handling for connection drops mid-stream
- Scanner has default buffer size (64KB) - large responses will break
- **Fix:** Use a proper SSE library or implement according to spec (handle `\n\n` boundaries, multi-line data, retries)

### 8. **No Cleanup on Context Cancellation**
**If context is cancelled mid-stream:**
- "Thinking..." message stays forever
- No way to notify user that request was cancelled
- **Fix:** Defer cleanup, send cancellation message to Slack

### 9. **Hardcoded Instance ID (Line 241)**
```go
InstanceId: "default", // Runtime proxy will route to the correct instance
```
**This is a hack:**
- What if the project doesn't have a "default" instance?
- What if instance naming changes?
- **Fix:** Actually query the deployment and use the real instance ID

### 10. **Missing Input Validation**
**Token validation is minimal:**
- Only checks prefix `rill_`
- No length validation (DoS risk)
- No format validation beyond prefix
- **Fix:** Add proper validation (length limits, format checks, maybe even verify token format with regex)

### 11. **No Deduplication**
**Slack can send duplicate events:**
- If Slack retries, you process the same event twice
- Can create duplicate conversations, send duplicate messages
- **Fix:** Check `event.EventTimeStamp` and deduplicate (Slack recommends this)

### 12. **Error Messages Might Leak Information**
```go
return s.sendSlackMessage(ctx, teamID, channelID, threadTS, 
    fmt.Sprintf("Error: Rill API returned %d: %s", resp.StatusCode, string(body)))
```
**Error body might contain sensitive data:**
- Could leak internal errors, stack traces, etc.
- **Fix:** Sanitize error messages, don't expose internal details

### 13. **No Retry Logic**
**If Slack API fails, you just give up:**
- Network hiccups cause permanent failures
- No exponential backoff
- **Fix:** Add retry logic with exponential backoff for transient errors

### 14. **Missing Observability**
**No metrics, no tracing:**
- Can't monitor bot usage
- Can't debug performance issues
- Can't track error rates
- **Fix:** Add metrics (request count, latency, error rates), add tracing spans

### 15. **Database Transaction Issues**
**No transaction handling:**
- If workspace creation fails after token insert, you're in inconsistent state
- Race condition: two requests can try to create workspace simultaneously
- **Fix:** Use database transactions for multi-step operations

## 🟡 MINOR ISSUES

### 16. **Inefficient String Concatenation (Line 319)**
```go
fullResponse += "\n"
fullResponse += block.GetText()
```
**String concatenation in a loop is inefficient:**
- Creates new strings on each iteration
- **Fix:** Use `strings.Builder`

### 17. **No Message Length Limits**
**Slack has message limits (4000 chars):**
- Long responses will be truncated or fail
- **Fix:** Split long messages into multiple messages or use blocks

### 18. **Missing Migration File Check**
**Can't verify migration was created correctly:**
- Migration file not shown in diff
- **Fix:** Ensure migration follows naming convention and is tested

### 19. **No Test Coverage**
**Zero tests:**
- SSE parsing logic is complex and untested
- Event handling logic is untested
- **Fix:** Add unit tests, especially for SSE parsing and edge cases

### 20. **Hardcoded Agent (Line 244)**
```go
Agent: "analyst_agent",
```
**Hardcoded agent type:**
- What if users want different agents?
- **Fix:** Make it configurable or derive from context

## 🔵 EDGE CASES MISSING

1. **What if user deletes their token from Rill?** - Token becomes invalid, bot should handle gracefully
2. **What if workspace is removed from Slack?** - No cleanup mechanism
3. **What if conversation ID changes mid-stream?** - Current code overwrites, might lose context
4. **What if Slack message update fails but stream continues?** - No recovery mechanism
5. **What if multiple users in same thread?** - Conversation state is per-thread, not per-user
6. **What if project is deleted?** - No handling for missing project
7. **What if deployment is hibernated?** - No handling for unavailable deployments
8. **What if SSE stream sends malformed JSON?** - Parser silently ignores (line 309: `err == nil`)
9. **What if Slack API rate limits us?** - No handling for 429 responses
10. **What if token encryption key is rotated?** - Old tokens become unreadable, no migration path

## 📋 RECOMMENDATIONS

1. **Add a worker pool** for processing Slack events
2. **Add proper HTTP client** with timeouts and connection pooling
3. **Add rate limiting** per user/workspace
4. **Add retry logic** with exponential backoff
5. **Add observability** (metrics, tracing, structured logging)
6. **Add tests** especially for SSE parsing
7. **Add input validation** and sanitization
8. **Handle edge cases** (deleted projects, invalid tokens, etc.)
9. **Add cleanup mechanisms** for abandoned conversations
10. **Add monitoring/alerts** for bot health

## 🎯 VERDICT

**This code is not production-ready.** While it might work in happy-path scenarios, it will fail catastrophically under load or when edge cases occur. The fire-and-forget goroutines alone are a red flag that suggests this wasn't designed with production constraints in mind.

**Recommendation:** Refactor before merging. At minimum, fix the critical issues (#1-6) before this sees production traffic.
