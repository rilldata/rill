<script lang="ts">
  import { page } from "$app/stores";
  import { onMount, onDestroy } from "svelte";
  import {
    ConnectionStatus,
    createSSEStream,
  } from "@rilldata/web-common/runtime-client/sse";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    V1LogLevel,
    type V1Log,
    type V1WatchLogsResponse,
  } from "@rilldata/web-common/runtime-client";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import {
    createUrlFilterSync,
    parseArrayParam,
    parseStringParam,
  } from "@rilldata/web-common/lib/url-filter-sync";
  import { ProjectLogsStore, type LogEntry } from "./project-logs-store";

  const runtimeClient = useRuntimeClient();

  const MAX_LOGS = 500;
  const REPLAY_LIMIT = 100;

  const filterableLevels = [
    { value: V1LogLevel.LOG_LEVEL_DEBUG, label: "Debug" },
    { value: V1LogLevel.LOG_LEVEL_INFO, label: "Info" },
    { value: V1LogLevel.LOG_LEVEL_WARN, label: "Warn" },
    { value: V1LogLevel.LOG_LEVEL_ERROR, label: "Error" },
  ];

  const logStore = new ProjectLogsStore(MAX_LOGS);
  // Bumped on each addLog so Svelte reactivity re-runs filtering.
  let logsVersion = 0;
  let logsContainer: HTMLDivElement;
  let connectionError: string | null = null;
  let filterDropdownOpen = false;
  let searchText = parseStringParam($page.url.searchParams.get("q"));
  let selectedLevels = parseArrayParam($page.url.searchParams.get("level"));
  let mounted = false;

  const filterSync = createUrlFilterSync([
    { key: "q", type: "string" },
    { key: "level", type: "array" },
  ]);
  filterSync.init($page.url);

  // Sync URL → local state on external navigation (back/forward)
  $: if (mounted && filterSync.hasExternalNavigation($page.url)) {
    filterSync.markSynced($page.url);
    searchText = parseStringParam($page.url.searchParams.get("q"));
    selectedLevels = parseArrayParam($page.url.searchParams.get("level"));
  }

  // Sync filter state → URL
  $: if (mounted) {
    filterSync.syncToUrl({ q: searchText, level: selectedLevels });
  }

  // Typed decoder layer: `createSSEStream` composes connection + subscriber so
  // component code stays in UI concerns.
  const logsStream = createSSEStream<{
    log: V1WatchLogsResponse;
    error: { code?: string; message?: string };
  }>({
    connection: {
      maxRetryAttempts: 5,
      retryOnError: true,
      retryOnClose: true,
    },
    decoders: {
      log: (data) => JSON.parse(data) as V1WatchLogsResponse,
      error: (data) => {
        try {
          return JSON.parse(data) as { code?: string; message?: string };
        } catch {
          return { message: data };
        }
      },
    },
  });

  $: connectionStatus = logsStream.status;
  $: isConnected = $connectionStatus === ConnectionStatus.OPEN;
  $: isConnecting = $connectionStatus === ConnectionStatus.CONNECTING;
  $: isClosed = $connectionStatus === ConnectionStatus.CLOSED;
  $: hasConnectionError = isClosed && connectionError !== null;

  let filteredLogs: LogEntry[] = [];
  let totalLogs = 0;
  $: {
    // Depend on logsVersion so Svelte re-runs when new logs arrive; the
    // reference below keeps the compiler from treating it as unused.
    void logsVersion;
    filteredLogs = logStore.getFiltered({
      levels: selectedLevels,
      search: searchText,
    });
    totalLogs = logStore.getAll().length;
  }

  $: selectedLevelLabel = (() => {
    if (selectedLevels.length === 0) return "All levels";
    if (selectedLevels.length === 1) {
      return (
        filterableLevels.find((l) => l.value === selectedLevels[0])?.label ??
        "1 level"
      );
    }
    const first = filterableLevels.find(
      (l) => l.value === selectedLevels[0],
    )?.label;
    return `${first}, +${selectedLevels.length - 1} other${selectedLevels.length > 2 ? "s" : ""}`;
  })();

  let unsubs: (() => void)[] = [];

  onMount(() => {
    mounted = true;
    const { host, instanceId } = runtimeClient;
    if (!host || !instanceId) return;

    const url = `${host}/v1/instances/${instanceId}/sse?events=log&logs_replay=true&logs_replay_limit=${REPLAY_LIMIT}`;

    unsubs = [
      logsStream.on("log", handleLogMessage),
      logsStream.on("error", handleServerError),
      logsStream.onConnection("error", handleTransportError),
      logsStream.onConnection("open", handleOpen),
    ];

    logsStream.start(url, {
      getJwt: () => runtimeClient.getJwt(),
    });
  });

  onDestroy(() => {
    unsubs.forEach((fn) => fn());
    logsStream.cleanup();
  });

  function isNearBottom(el: HTMLElement, threshold = 50): boolean {
    return el.scrollHeight - el.scrollTop - el.clientHeight <= threshold;
  }

  function handleLogMessage(response: V1WatchLogsResponse) {
    const log = response.log as V1Log | undefined;
    if (!log) return;

    logStore.addLog(log);
    logsVersion++;

    if (logsContainer && isNearBottom(logsContainer)) {
      requestAnimationFrame(() => {
        logsContainer.scrollTop = logsContainer.scrollHeight;
      });
    }
  }

  function handleServerError(payload: { code?: string; message?: string }) {
    connectionError = payload.message || payload.code || "Server error";
  }

  function handleTransportError(error: Error) {
    console.error("Logs SSE error:", error);
    connectionError = error.message || "Connection failed";
  }

  function handleOpen() {
    connectionError = null;
  }

  function retryConnection() {
    const { host, instanceId } = runtimeClient;
    if (!host || !instanceId) return;

    connectionError = null;
    const url = `${host}/v1/instances/${instanceId}/sse?events=log&logs_replay=true&logs_replay_limit=${REPLAY_LIMIT}`;
    logsStream.start(url, {
      getJwt: () => runtimeClient.getJwt(),
    });
  }

  function getLevelClass(level: V1LogLevel | undefined): string {
    switch (level) {
      case V1LogLevel.LOG_LEVEL_DEBUG:
        return "text-fg-secondary";
      case V1LogLevel.LOG_LEVEL_INFO:
        return "text-fg-primary";
      case V1LogLevel.LOG_LEVEL_WARN:
        return "text-yellow-700";
      case V1LogLevel.LOG_LEVEL_ERROR:
        return "text-red-700";
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
      return date.toISOString().slice(0, 23);
    } catch {
      return "";
    }
  }

  function toggleLevel(level: string) {
    if (selectedLevels.includes(level)) {
      selectedLevels = selectedLevels.filter((l) => l !== level);
    } else {
      selectedLevels = [...selectedLevels, level];
    }
  }

  function clearFilters() {
    selectedLevels = [];
    searchText = "";
  }
</script>

<section class="flex flex-col gap-y-4 size-full">
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-x-2">
      <h2 class="text-lg font-medium">Logs</h2>
      <span
        class="status-badge"
        class:status-live={isConnected}
        class:status-connecting={isConnecting}
        class:status-error={hasConnectionError}
      >
        <span class="status-dot"></span>
        {#if isConnected}
          Live
        {:else if isConnecting}
          Connecting
        {:else if hasConnectionError}
          Disconnected
        {:else}
          Idle
        {/if}
      </span>
    </div>
  </div>

  <div class="flex flex-row items-center gap-x-4 min-h-9">
    <div class="flex-1 min-w-0 min-h-9">
      <Search
        bind:value={searchText}
        placeholder="Search"
        large
        autofocus={false}
        showBorderOnFocus={false}
        retainValueOnMount
      />
    </div>

    <DropdownMenu.Root bind:open={filterDropdownOpen}>
      <DropdownMenu.Trigger
        class="min-w-fit min-h-9 flex flex-row gap-1 items-center rounded-sm border bg-input {filterDropdownOpen
          ? 'bg-gray-200'
          : 'hover:bg-surface-hover'} px-2 py-1"
      >
        <span class="text-fg-secondary font-medium">
          {selectedLevelLabel}
        </span>
        {#if filterDropdownOpen}
          <CaretUpIcon size="12px" />
        {:else}
          <CaretDownIcon size="12px" />
        {/if}
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="start" class="w-48">
        {#each filterableLevels as level}
          <DropdownMenu.CheckboxItem
            closeOnSelect={false}
            checked={selectedLevels.includes(level.value)}
            onCheckedChange={() => toggleLevel(level.value)}
          >
            {level.label}
          </DropdownMenu.CheckboxItem>
        {/each}
      </DropdownMenu.Content>
    </DropdownMenu.Root>

    {#if selectedLevels.length > 0 || searchText}
      <button
        class="shrink-0 text-sm text-primary-500 hover:text-primary-600 whitespace-nowrap"
        onclick={clearFilters}
      >
        Clear
      </button>
    {/if}
  </div>

  <div class="logs-container" bind:this={logsContainer}>
    {#if hasConnectionError}
      <div class="error-state">
        <span class="text-red-600">Connection failed: {connectionError}</span>
        <button class="retry-button" onclick={retryConnection}> Retry </button>
      </div>
    {:else if totalLogs === 0}
      <div class="empty-state">Waiting for logs...</div>
    {:else if filteredLogs.length === 0}
      <div class="empty-state">No logs match the current filters</div>
    {:else}
      {#each filteredLogs as log (log._id)}
        <div class="log-entry {getLevelClass(log.level)}">
          <p>
            {formatTime(log.time)}
            {getLevelLabel(log.level)}
            {log.message}
            {log.jsonPayload ?? ""}
          </p>
        </div>
      {/each}
    {/if}
  </div>
</section>

<style lang="postcss">
  .status-badge {
    @apply inline-flex items-center gap-1.5 text-xs font-medium px-2 py-0.5 rounded-full;
    @apply bg-gray-100 text-fg-secondary;
  }

  .status-dot {
    @apply w-1.5 h-1.5 rounded-full bg-gray-400;
  }

  .status-live {
    @apply bg-green-100 text-green-700;
  }

  .status-live .status-dot {
    @apply bg-green-500;
  }

  .status-connecting {
    @apply bg-yellow-100 text-yellow-700;
  }

  .status-connecting .status-dot {
    @apply bg-yellow-500 animate-pulse;
  }

  .status-error {
    @apply bg-red-100 text-red-700;
  }

  .status-error .status-dot {
    @apply bg-red-500;
  }

  .logs-container {
    @apply flex-1 overflow-y-auto overflow-x-hidden font-mono text-xs;
    @apply bg-surface-background border border-gray-200 rounded-md p-2;
    min-height: 300px;
    max-height: 70vh;
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
