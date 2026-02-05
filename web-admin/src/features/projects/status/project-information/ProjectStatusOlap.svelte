<script lang="ts">
  import type { V1Connector } from "@rilldata/web-common/runtime-client";
  import TableIcon from "@rilldata/web-common/components/icons/TableIcon.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { formatConnectorName } from "../display-utils";

  export let olapConnector: V1Connector | undefined;
  export let isLoading = false;
  export let isError = false;
</script>

<div class="info-cell">
  <div class="cell-header">
    <span class="icon-muted"><TableIcon size="16px" /></span>
    <span class="cell-label">OLAP Engine</span>
    <Tooltip distance={8}>
      <a
        href="https://docs.rilldata.com/developers/build/connectors/olap"
        target="_blank"
        rel="noreferrer noopener"
        class="info-link"
      >
        <InfoCircle size="14px" color="#9ca3af" />
      </a>
      <TooltipContent slot="tooltip-content">
        <p class="tooltip-text">Learn about supported OLAP engines.</p>
      </TooltipContent>
    </Tooltip>
  </div>
  <div class="cell-content">
    {#if isLoading}
      <span class="text-sm text-gray-400">Loading...</span>
    {:else if isError}
      <span class="text-sm text-red-500">Failed to load</span>
    {:else if olapConnector}
      <span class="connector-name">
        {formatConnectorName(olapConnector.type)}
      </span>
      <div class="connector-details">
        {olapConnector.provision ? "Rill-Managed" : "Self-Managed"}
      </div>
    {:else}
      <span class="connector-name">-</span>
    {/if}
  </div>
</div>

<style lang="postcss">
  .info-cell {
    @apply flex flex-col gap-1.5;
  }

  .cell-header {
    @apply flex items-center gap-1.5;
  }

  .cell-label {
    @apply text-xs font-medium text-gray-500 uppercase tracking-wide;
  }

  .cell-content {
    @apply flex flex-col gap-1;
  }

  .connector-name {
    @apply text-sm font-medium text-gray-900;
  }

  .connector-details {
    @apply flex items-center gap-2 text-xs text-gray-600;
  }

  .icon-muted {
    @apply text-gray-500;
  }

  .icon-muted :global(svg path) {
    fill: currentColor;
  }

  .info-link {
    @apply flex items-center;
  }

  .info-link:hover :global(svg) {
    color: #6b7280;
  }

  .tooltip-text {
    @apply text-sm max-w-[200px];
  }
</style>
