# SSE Client

Layered Server-Sent Events client used by `FileAndResourceWatcher`, `Conversation`, and `ProjectLogsPage`. Each module owns one concern; consumers compose only the layers they need.

## Layers

```
SSESubscriber             — typed routing (decoder per event type)
       │ subscribes to
       ▼
SSEConnection             — status, retry, onBeforeReconnect hook
       │ owns
       ▼
SSEFetchClient            — fetch + AbortController transport
       │ uses
       ▼
sse-protocol              — SSE wire-format parser

SSEConnectionLifecycle    — optional; observes browser visibility/idle
                            signals and drives pause/resume on an
                            SSEConnection
```

- **`sse-protocol`**: pure parser only.
- **`sse-fetch-client`**: fetch transport only (no reconnect logic).
- **`sse-connection`**: status + retry/backoff + `onBeforeReconnect`.
- **`sse-connection-lifecycle`**: optional pause/resume policy from browser signals.
- **`sse-subscriber`**: typed decoding/routing (`undefined` type defaults to `"message"`).

## Composition

- Use **`createSSEStream`** for most consumers (current repo usage).
- Use **explicit layers** when you need direct low-level control over each object.

### `createSSEStream`

```ts
const stream = createSSEStream<{
  file: V1WatchFilesResponse;
  resource: V1WatchResourcesResponse;
}>({
  connection: {
    maxRetryAttempts: 3,
    retryOnError: true,
    retryOnClose: true,
    onBeforeReconnect: refreshJwt,
  },
  decoders: {
    file: (data) => JSON.parse(data),
    resource: (data) => JSON.parse(data),
  },
  lifecycle: {
    idleTimeouts: { short: 20_000, normal: 120_000 },
  },
});

stream.on("file", handleFile);
stream.on("resource", handleResource);
stream.onConnection("error", handleTransportError);
stream.start(url, { getJwt });
```

Facade API (thin wrapper over `SSEConnection` + `SSESubscriber`):

- `status`
- `on` / `once`
- `onConnection` / `onceConnection`
- `start`, `pause`, `resumeIfPaused`
- `close(cleanup = false)`, `cleanup()` (`close(true)`)

## Design Notes

- One-shot streams (chat) usually skip lifecycle and reconnect hooks.
- Long-lived streams (watcher/logs) use retry and, when needed, lifecycle.
- `SSEConnection` stays payload-agnostic; decoding stays in `SSESubscriber`.
