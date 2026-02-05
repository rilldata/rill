<script lang="ts">
  import Brain from "@rilldata/web-common/components/icons/Brain.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { formatConnectorName } from "../display-utils";

  // The admin connector is managed by Rill and should display as "Rill Managed"
  const RILL_MANAGED_CONNECTOR_NAME = "admin";

  export let aiConnector: { name?: string; type?: string } | undefined;
  export let isLoading = false;
  export let isError = false;

  $: isRillManaged = aiConnector?.name === RILL_MANAGED_CONNECTOR_NAME;
</script>

<div class="info-cell">
  <div class="cell-header">
    <Brain size="16" color="#6b7280" />
    <span class="cell-label">AI</span>
    <Tooltip distance={8}>
      <a
        href="https://docs.rilldata.com/developers/build/connectors/data-source/openai"
        target="_blank"
        rel="noreferrer noopener"
        class="info-link"
      >
        <InfoCircle size="14px" color="#9ca3af" />
      </a>
      <TooltipContent slot="tooltip-content">
        <p class="tooltip-text">Configure AI connectors for your project.</p>
      </TooltipContent>
    </Tooltip>
  </div>
  <div class="cell-content">
    {#if isLoading}
      <span class="text-sm text-gray-400">Loading...</span>
    {:else if isError}
      <span class="text-sm text-red-500">Failed to load</span>
    {:else if aiConnector}
      <span class="connector-name">
        {formatConnectorName(aiConnector.name)}
      </span>
      <div class="connector-details">
        {isRillManaged ? "Rill Managed" : aiConnector.type}
      </div>
    {:else}
      <span class="connector-name">Rill Managed</span>
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
