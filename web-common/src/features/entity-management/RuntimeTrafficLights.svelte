<script lang="ts">
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { fileAndResourceWatcher } from "./file-and-resource-watcher";
  import { Tooltip } from "bits-ui";
  import { ConnectionStatus } from "@rilldata/web-common/runtime-client/sse-connection-manager";

  const { status: statusStore } = fileAndResourceWatcher;

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
