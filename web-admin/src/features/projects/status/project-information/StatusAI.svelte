<script lang="ts">
  import Brain from "@rilldata/web-common/components/icons/Brain.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { formatConnectorName } from "../display-utils";

  export let aiConnector: string | undefined;
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
        <p class="tooltip-text">
          Configure AI connectors for your project.
        </p>
      </TooltipContent>
    </Tooltip>
  </div>
  <div class="cell-content">
    {#if aiConnector}
      <span class="connector-name">{formatConnectorName(aiConnector)}</span>
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
