<script lang="ts">
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import { connectorIconMapping } from "@rilldata/web-common/features/connectors/connector-icon-mapping.ts";

  export let connectorDriver: V1ConnectorDriver;

  $: connectorName = connectorDriver.name!;
  $: displayIcon = connectorIconMapping[connectorName];
  $: displayName = connectorDriver.displayName ?? connectorName;
</script>

{#if displayName}
  <div class="flex flex-row items-center px-6 py-4 gap-1 border-b">
    {#if displayIcon}
      <svelte:component this={displayIcon} size="18px" />
    {/if}
    <span class="text-lg leading-none font-semibold">{displayName}</span>
  </div>
{/if}
