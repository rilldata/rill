<script lang="ts">
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { ConnectionStatus } from "@rilldata/web-common/runtime-client/sse";
  import { getContext } from "svelte";
  import { writable, type Writable } from "svelte/store";
  import { WATCHER_CONTEXT_KEY, type WatcherContext } from "./watcher-context";

  // Context is supplied by <FileAndResourceWatcher>. If this component is
  // rendered outside a watcher provider (e.g. a route that doesn't mount
  // the watcher), fall back to a static CLOSED store so it still renders
  // rather than throws.
  const context = getContext<WatcherContext | undefined>(WATCHER_CONTEXT_KEY);
  const statusStore: Writable<ConnectionStatus> =
    context?.status ?? writable(ConnectionStatus.CLOSED);

  let showIndicator: boolean = true;
  let connectedTimer: ReturnType<typeof setTimeout>;

  $: status = $statusStore;

  $: if (status !== ConnectionStatus.OPEN) {
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
      <div class="{status}  flex-none size-[9px] rounded-full opacity-75"></div>
    </Tooltip.Trigger>

    <Tooltip.Content side="right" sideOffset={8} class="z-50">
      <TooltipContent>
        {#if status === ConnectionStatus.CONNECTING}
          Attempting to reconnect
        {:else if status === ConnectionStatus.CLOSED}
          Disconnected from Rill engine
        {:else if status === ConnectionStatus.PAUSED}
          Connection paused due to inactivity
        {:else if status === ConnectionStatus.OPEN}
          Connected to Rill engine
        {/if}
      </TooltipContent>
    </Tooltip.Content>
  </Tooltip.Root>
{/if}

<style lang="postcss">
  .closed,
  .paused {
    @apply bg-red-600 animate-pulse;
  }

  .connecting {
    @apply bg-yellow-600 animate-pulse;
  }

  .open {
    @apply bg-emerald-600;
  }
</style>
