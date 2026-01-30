<script lang="ts">
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createRuntimeServiceGetLogs,
    RuntimeServiceGetLogsLevel,
  } from "@rilldata/web-common/runtime-client";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { onMount, tick } from "svelte";

  $: ({ instanceId } = $runtime);

  // Auto-scroll tracking
  let logsContainer: HTMLDivElement;
  let userScrolledUp = false;

  $: logsQuery = createRuntimeServiceGetLogs(
    instanceId,
    {
      limit: 200,
      ascending: true,
      level: RuntimeServiceGetLogsLevel.LOG_LEVEL_INFO,
    },
    {
      query: {
        enabled: !!instanceId,
        refetchInterval: 1000, // Always live - refresh every second
      },
    },
  );

  // Filter logs to only show model-related entries
  $: modelLogs =
    $logsQuery.data?.logs?.filter(
      (log) =>
        log.message?.toLowerCase().includes("model") ||
        log.jsonPayload?.toLowerCase().includes("model"),
    ) ?? [];

  // Auto-scroll to bottom when logs update (if user hasn't scrolled up)
  $: if (modelLogs.length > 0 && !userScrolledUp) {
    tick().then(() => scrollToBottom());
  }

  function scrollToBottom() {
    if (logsContainer) {
      logsContainer.scrollTop = logsContainer.scrollHeight;
    }
  }

  function handleScroll() {
    if (!logsContainer) return;
    const { scrollTop, scrollHeight, clientHeight } = logsContainer;
    // Consider "at bottom" if within 20px of the bottom
    userScrolledUp = scrollHeight - scrollTop - clientHeight > 20;
  }

  onMount(() => {
    // Scroll to bottom on initial load
    tick().then(() => scrollToBottom());
  });

  function formatTime(time: string | undefined): string {
    if (!time) return "";
    const date = new Date(time);
    return date.toLocaleTimeString(undefined, {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  }

  function getLevelClass(level: string | undefined): string {
    switch (level) {
      case "LOG_LEVEL_ERROR":
      case "LOG_LEVEL_FATAL":
        return "text-red-600";
      case "LOG_LEVEL_WARN":
        return "text-yellow-600";
      default:
        return "text-fg-muted";
    }
  }

  function getLevelLabel(level: string | undefined): string {
    switch (level) {
      case "LOG_LEVEL_ERROR":
        return "ERROR";
      case "LOG_LEVEL_FATAL":
        return "FATAL";
      case "LOG_LEVEL_WARN":
        return "WARN";
      case "LOG_LEVEL_INFO":
        return "INFO";
      case "LOG_LEVEL_DEBUG":
        return "DEBUG";
      default:
        return "INFO";
    }
  }
</script>

<section class="flex flex-col gap-y-2">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Model Logs</h2>
    {#if $logsQuery.isLoading}
      <DelayedSpinner isLoading={true} size="14px" />
    {/if}
  </div>

  <div
    bind:this={logsContainer}
    on:scroll={handleScroll}
    class="bg-surface border rounded-md p-3 font-mono text-xs h-64 overflow-y-auto"
  >
    {#if $logsQuery.isError}
      <div class="text-red-600">
        Error loading logs: {$logsQuery.error?.response?.data?.message ??
          $logsQuery.error?.message ??
          "Unknown error"}
      </div>
    {:else if modelLogs.length === 0}
      <div class="text-fg-secondary">No model logs found</div>
    {:else}
      {#each modelLogs as log}
        <div class="py-0.5 flex gap-x-2 hover:bg-surface-hover">
          <span class="text-fg-muted flex-none">{formatTime(log.time)}</span>
          <span class="{getLevelClass(log.level)} flex-none w-12"
            >{getLevelLabel(log.level)}</span
          >
          <span class="text-fg-secondary break-all"
            >{log.message}{log.jsonPayload ? ` ${log.jsonPayload}` : ""}</span
          >
        </div>
      {/each}
    {/if}
  </div>
</section>
