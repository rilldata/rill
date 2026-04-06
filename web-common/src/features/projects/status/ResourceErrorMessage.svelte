<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
  import { AlertTriangleIcon } from "lucide-svelte";

  export let message: string;
  export let status: V1ReconcileStatus;
  export let testErrors: string[] = [];

  $: hasTestErrors = testErrors.length > 0;
  $: testErrorMessage = testErrors.join("\n");
</script>

<div class="container">
  {#if status === V1ReconcileStatus.RECONCILE_STATUS_PENDING || status === V1ReconcileStatus.RECONCILE_STATUS_RUNNING}
    <Tooltip distance={8}>
      <LoadingSpinner size="18px" />
      <TooltipContent slot="tooltip-content">Reconciling</TooltipContent>
    </Tooltip>
  {:else if hasTestErrors}
    <Tooltip distance={8}>
      <button
        class="hover:bg-surface-hover rounded p-1 active:bg-surface-active group"
        onclick={() =>
          copyToClipboard(testErrorMessage, "Copied test error to clipboard")}
      >
        <CopyIcon
          size="18px"
          className="text-fg-secondary hidden group-hover:block"
        />

        <AlertTriangleIcon
          size="18px"
          class="text-yellow-500 group-hover:hidden"
        />
      </button>
      <TooltipContent slot="tooltip-content" maxWidth="300px">
        <p class="line-clamp-6">{testErrorMessage}</p>
      </TooltipContent>
    </Tooltip>
  {:else if message}
    <Tooltip distance={8}>
      <button
        class="hover:bg-surface-hover rounded p-1 active:bg-surface-active group"
        onclick={() =>
          copyToClipboard(message, "Copied error message to clipboard")}
      >
        <CopyIcon
          size="18px"
          className="text-gray-500 hidden group-hover:block"
        />

        <CancelCircle size="18px" className="text-red-500 group-hover:hidden" />
      </button>
      <TooltipContent slot="tooltip-content" maxWidth="300px">
        <p class="line-clamp-6">{message}</p>
      </TooltipContent>
    </Tooltip>
  {:else}
    <Tooltip distance={8}>
      <Check size="18px" className="text-green-500" />
      <TooltipContent slot="tooltip-content">Complete</TooltipContent>
    </Tooltip>
  {/if}
</div>

<style lang="postcss">
  .container {
    @apply flex items-center justify-center flex-none;
  }
</style>
