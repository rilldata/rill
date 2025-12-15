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

  let state: "connected" | "closed" | "reconnecting" = "connected";
  let showIndicator: boolean = true;

  let connectedTimer: ReturnType<typeof setTimeout>;

  $: reconnecting = $fileAttempts > 0 || $resourceAttempts > 0;

  $: closed = $fileWatcherClosed || $resourceWatcherClosed;

  $: if (reconnecting) {
    state = "reconnecting";
  } else if (closed) {
    state = "closed";
  } else {
    state = "connected";
  }

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
      <div class="{state}  flex-none size-[9px] rounded-full opacity-75"></div>
    </Tooltip.Trigger>

    <Tooltip.Content side="right" sideOffset={8}>
      <TooltipContent>
        {#if state === "reconnecting"}
          Attempting to reconnect
        {:else if state === "closed"}
          Connection closed due to inactivity
        {:else}
          Connected to Rill engine
        {/if}
      </TooltipContent>
    </Tooltip.Content>
  </Tooltip.Root>
{/if}

<style lang="postcss">
  .closed {
    @apply bg-red-600 animate-pulse;
  }

  .reconnecting {
    @apply bg-yellow-600 animate-pulse;
  }

  .connected {
    @apply bg-emerald-600;
  }
</style>
