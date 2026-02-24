<script lang="ts">
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../../layout/config";
  import { useRuntimeClient } from "../../../runtime-client/v2";
  import { createRuntimeServiceAnalyzeConnectors } from "../../../runtime-client/v2/gen";
  import ConnectorEntry from "./ConnectorEntry.svelte";
  import type { ConnectorExplorerStore } from "./connector-explorer-store";

  export let store: ConnectorExplorerStore;
  export let olapOnly: boolean = false;

  const client = useRuntimeClient();

  $: connectors = createRuntimeServiceAnalyzeConnectors(
    client,
    {},
    {
      query: {
        // Retry transient errors during runtime resets (e.g. project initialization)
        retry: (failureCount) => failureCount < 3,
        retryDelay: 1000,
        // sort alphabetically
        select: (data) => {
          if (!data?.connectors) return;

          const filtered = (
            olapOnly
              ? data.connectors.filter((c) => c?.driver?.implementsOlap)
              : data.connectors
          ).sort((a, b) =>
            (a?.name as string).localeCompare(b?.name as string),
          );
          return { connectors: filtered };
        },
      },
    },
  );
  $: ({ data, error } = $connectors);
</script>

<div class="wrapper">
  {#if error}
    <span class="message">
      {error.message}
    </span>
  {:else if data?.connectors}
    {#if data.connectors.length === 0}
      <span class="message"> No data found. Add data to get started! </span>
    {:else}
      <ol transition:slide={{ duration }}>
        {#each data.connectors as connector (connector.name)}
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
