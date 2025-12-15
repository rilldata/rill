<script lang="ts">
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    fileWatcher,
    resourceWatcher,
  } from "@rilldata/web-common/features/entity-management/watchers";
  import { Tooltip } from "bits-ui";

  const { retryAttempts: fileAttempts, closed: fileWatcherClosed } =
    fileWatcher;
  const { retryAttempts: resourceAttempts, closed: resourceWatcherClosed } =
    resourceWatcher;

  let state: "connected" | "disconnected" | "reconnecting" = "connected";
  let showIndicator: boolean = true;

  $: if ($fileAttempts > 0 || $resourceAttempts > 0) {
    state = "reconnecting";
  } else if ($fileWatcherClosed || $resourceWatcherClosed) {
    state = "disconnected";
  } else {
    state = "connected";
  }

  let connectedTimer: ReturnType<typeof setTimeout>;

  $: if (state !== "connected") {
    clearTimeout(connectedTimer);
    showIndicator = true;
  } else {
    connectedTimer = setTimeout(() => {
      showIndicator = false;
    }, 5000);
  }
</script>

{#if showIndicator}
  <Tooltip.Root>
    <Tooltip.Trigger>
      {$fileAttempts}
      {$resourceAttempts}
      <div class="{state}  flex-none size-[9px] rounded-full opacity-75"></div>
    </Tooltip.Trigger>

    <Tooltip.Content side="right" sideOffset={8}>
      <TooltipContent>
        {#if state === "reconnecting"}
          Connection is experiencing issues. Please try refreshing the page
        {:else if state === "disconnected"}
          Connection closed due to inactivity. Interact with the page to
          reconnect
        {:else}
          Connected to Rill engine
        {/if}
      </TooltipContent>
    </Tooltip.Content>
  </Tooltip.Root>
{/if}

<style lang="postcss">
  .reconnecting {
    @apply bg-red-600 animate-pulse;
  }

  .disconnected {
    @apply bg-yellow-600 animate-pulse;
  }

  .connected {
    @apply bg-emerald-600;
  }
</style>
