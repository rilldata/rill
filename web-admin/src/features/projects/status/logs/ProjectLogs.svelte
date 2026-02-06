<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import {
    SSEConnectionManager,
    ConnectionStatus,
  } from "@rilldata/web-common/runtime-client/sse-connection-manager";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { V1LogLevel, type V1Log } from "@rilldata/web-common/runtime-client";

  // Maximum number of logs to keep in memory to prevent excessive memory usage
  const MAX_LOGS = 500;
  // Number of recent logs to fetch when initially connecting to get some history
  const REPLAY_LIMIT = 100;

  let logs: V1Log[] = [];
  let logsContainer: HTMLDivElement;
  let connectionError: string | null = null;

  const logsConnection = new SSEConnectionManager({
    maxRetryAttempts: 5,
    retryOnError: true,
    retryOnClose: true,
  });

  $: connectionStatus = logsConnection.status;
  $: isConnected = $connectionStatus === ConnectionStatus.OPEN;
  $: isConnecting = $connectionStatus === ConnectionStatus.CONNECTING;
  $: isClosed = $connectionStatus === ConnectionStatus.CLOSED;
  $: hasConnectionError = isClosed && connectionError !== null;

  onMount(() => {
    const { host, instanceId, jwt } = $runtime;
    if (!host || !instanceId) return;

    // Use the unified SSE endpoint with events=log
    const url = `${host}/v1/instances/${instanceId}/sse?events=log&logs_replay=true&logs_replay_limit=${REPLAY_LIMIT}`;

    logsConnection.on("message", handleMessage);
    logsConnection.on("error", handleError);
    logsConnection.on("open", handleOpen);

    logsConnection.start(url, {
      headers: jwt?.token ? { Authorization: `Bearer ${jwt.token}` } : {},
    });
  });

  onDestroy(() => {
    logsConnection.close(true);
  });

  function handleMessage(message: { data: string; type?: string }) {
    try {
      // Only process log events
      if (message.type && message.type !== "log") return;

      const response = JSON.parse(message.data);
      // WatchLogsResponse has a `log` field directly
      const log = response.log as V1Log;
      if (log) {
        logs = [...logs, log].slice(-MAX_LOGS);

        // Always auto-scroll to latest log
        if (logsContainer) {
          requestAnimationFrame(() => {
            logsContainer.scrollTop = logsContainer.scrollHeight;
          });
        }
      }
    } catch (e) {
      // Log parse errors in development for debugging
      if (import.meta.env.DEV) {
        console.warn("Failed to parse log message:", e);
      }
    }
  }

  function handleError(error: Error) {
    console.error("Logs SSE error:", error);
    connectionError = error.message || "Connection failed";
  }

  function handleOpen() {
    // Clear error when connection succeeds
    connectionError = null;
  }

  function retryConnection() {
    const { host, instanceId, jwt } = $runtime;
    if (!host || !instanceId) return;

    connectionError = null;
    const url = `${host}/v1/instances/${instanceId}/sse?events=log&logs_replay=true&logs_replay_limit=${REPLAY_LIMIT}`;
    logsConnection.start(url, {
      headers: jwt?.token ? { Authorization: `Bearer ${jwt.token}` } : {},
    });
  }

  function getLevelClass(level: V1LogLevel | undefined): string {
    switch (level) {
      case V1LogLevel.LOG_LEVEL_DEBUG:
        return "text-fg-secondary";
      case V1LogLevel.LOG_LEVEL_INFO:
        return "text-fg-primary";
      case V1LogLevel.LOG_LEVEL_WARN:
        return "text-yellow-600";
      case V1LogLevel.LOG_LEVEL_ERROR:
        return "text-red-600";
      default:
        return "text-fg-secondary";
    }
  }

  function getLevelLabel(level: V1LogLevel | undefined): string {
    switch (level) {
      case V1LogLevel.LOG_LEVEL_DEBUG:
        return "DEBUG";
      case V1LogLevel.LOG_LEVEL_INFO:
        return "INFO";
      case V1LogLevel.LOG_LEVEL_WARN:
        return "WARN";
      case V1LogLevel.LOG_LEVEL_ERROR:
        return "ERROR";
      default:
        return "UNKNOWN";
    }
  }

  function formatTime(time: string | undefined): string {
    if (!time) return "";
    try {
      const date = new Date(time);
      // Format as ISO timestamp with milliseconds (like 2026-01-27T17:23:06.351)
      return date.toISOString().slice(0, 23);
    } catch {
      return "";
    }
  }
</script>

<section class="logs-section">
  <div class="logs-header">
    <div class="flex items-center gap-x-2">
      <h2 class="text-lg font-medium">Logs</h2>
      <span
        class="status-indicator"
        class:connected={isConnected}
        class:connecting={isConnecting}
        class:error={hasConnectionError}
      />
      {#if isConnecting}
        <span class="text-sm text-fg-secondary">Connecting...</span>
      {:else if hasConnectionError}
        <span class="text-sm text-red-600">Disconnected</span>
      {/if}
    </div>
  </div>

  <div class="logs-container" bind:this={logsContainer}>
    {#if hasConnectionError}
      <div class="error-state">
        <span class="text-red-600">Connection failed: {connectionError}</span>
        <button class="retry-button" on:click={retryConnection}> Retry </button>
      </div>
    {:else if logs.length === 0}
      <div class="empty-state">Waiting for logs...</div>
    {:else}
      {#each logs as log, i (log.time ?? i)}
        <div class="log-entry {getLevelClass(log.level)}">
          {formatTime(log.time)}
          {getLevelLabel(log.level)}
          {log.message}
          {log.jsonPayload ?? ""}
        </div>
      {/each}
    {/if}
  </div>
</section>

<style lang="postcss">
  .logs-section {
    @apply flex flex-col gap-y-2 min-w-0 overflow-hidden;
  }

  .logs-header {
    @apply flex items-center justify-between;
  }

  .status-indicator {
    @apply w-2 h-2 rounded-full bg-gray-400;
  }

  .status-indicator.connected {
    @apply bg-green-500;
  }

  .status-indicator.connecting {
    @apply bg-yellow-500 animate-pulse;
  }

  .status-indicator.error {
    @apply bg-red-500;
  }

  .logs-container {
    @apply h-96 overflow-y-auto overflow-x-hidden font-mono text-xs;
    @apply bg-surface-background border border-gray-200 rounded-md p-2;
  }

  .empty-state {
    @apply flex items-center justify-center h-full text-fg-secondary;
  }

  .error-state {
    @apply flex flex-col items-center justify-center h-full gap-2;
  }

  .retry-button {
    @apply px-3 py-1 text-sm font-medium text-primary-600 bg-primary-50;
    @apply border border-primary-200 rounded;
  }

  .retry-button:hover {
    @apply bg-primary-100;
  }

  .log-entry {
    @apply py-0.5 break-words;
    word-break: break-word;
  }

  .log-entry:hover {
    @apply bg-surface-hover;
  }
</style>
