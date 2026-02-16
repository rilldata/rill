<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";

  export let message: string;
  export let status: V1ReconcileStatus;

  // Check if the error message indicates waiting for an upstream dependency
  $: isWaitingForUpstreamDependency = message && (
    message.includes("is not idle") || 
    message.startsWith("dependency error:")
  );
</script>

<div class="container">
  {#if status === V1ReconcileStatus.RECONCILE_STATUS_PENDING || status === V1ReconcileStatus.RECONCILE_STATUS_RUNNING || isWaitingForUpstreamDependency}
    {#if isWaitingForUpstreamDependency}
      <Tooltip distance={8}>
        <LoadingSpinner size="18px" />
        <TooltipContent slot="tooltip-content" maxWidth="300px">
          <p>Waiting for upstream dependencies to finish processing</p>
        </TooltipContent>
      </Tooltip>
    {:else}
      <LoadingSpinner size="18px" />
    {/if}
  {:else if message}
    <Tooltip distance={8}>
      <button
        class="hover:bg-surface-hover rounded p-1 active:bg-gray-200 group"
        on:click={() =>
          copyToClipboard(message, "Copied error message to clipboard")}
      >
        <CopyIcon
          size="18px"
          className="text-fg-secondary hidden group-hover:block"
        />

        <CancelCircle size="18px" className="text-red-500 group-hover:hidden" />
      </button>
      <TooltipContent slot="tooltip-content" maxWidth="300px">
        <p class="line-clamp-6">{message}</p>
      </TooltipContent>
    </Tooltip>
  {:else}
    <Check size="18px" className="text-green-500" />
  {/if}
</div>

<style lang="postcss">
  .container {
    @apply flex items-center justify-center flex-none;
  }
</style>
