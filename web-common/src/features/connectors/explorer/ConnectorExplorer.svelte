<script lang="ts">
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../../layout/config";
  import { createRuntimeServiceAnalyzeConnectors } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import ConnectorEntry from "./ConnectorEntry.svelte";
  import type { ConnectorExplorerStore } from "./connector-explorer-store";
  import type { V1ConnectorDriver } from "../../../runtime-client";

  export let store: ConnectorExplorerStore;
  export let olapOnly: boolean = false;
  export let limitToConnector: string | undefined = undefined;
  export let limitedConnectorDriver: V1ConnectorDriver | undefined = undefined;

  $: ({ instanceId } = $runtime);

  $: connectors = createRuntimeServiceAnalyzeConnectors(instanceId, {
    query: {
      enabled: !!instanceId && !limitedConnectorDriver,
      // sort alphabetically
      select: (data) => {
        if (!data?.connectors) return;

        let filtered = (
          olapOnly
            ? data.connectors.filter((c) => c?.driver?.implementsOlap)
            : data.connectors
        ).sort((a, b) => (a?.name as string).localeCompare(b?.name as string));
        if (limitToConnector) {
          filtered = filtered.filter((c) => c?.name === limitToConnector);
        }

        return { connectors: filtered };
      },
    },
  });
  $: ({ data, error } = $connectors);
  $: connectorsData = limitedConnectorDriver
    ? {
        connectors: [
          {
            name:
              (limitToConnector as string) ??
              (limitedConnectorDriver.name as string) ??
              "",
            driver: limitedConnectorDriver,
          },
        ],
      }
    : data;
</script>

<div class="wrapper">
  {#if error}
    <span class="message">
      {error.message}
    </span>
  {:else if connectorsData?.connectors}
    {#if connectorsData.connectors.length === 0}
      <span class="message"> No data found. Add data to get started! </span>
    {:else}
      <ol transition:slide={{ duration }}>
        {#each connectorsData.connectors as connector (connector.name)}
          <ConnectorEntry {connector} {store} />
        {/each}
      </ol>
    {/if}
  {/if}
</div>

<style lang="postcss">
  .wrapper {
    @apply overflow-auto px-0 pb-4;
  }

  .message {
    @apply pl-2 pr-3.5 py-2;
    @apply flex flex-none;
    @apply text-fg-secondary;
  }
</style>
