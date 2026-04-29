<script lang="ts">
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../../layout/config";
  import { useRuntimeClient } from "../../../runtime-client/v2";
  import ConnectorEntry from "./ConnectorEntry.svelte";
  import type { ConnectorExplorerStore } from "./connector-explorer-store";
  import { getAnalyzedConnectors } from "@rilldata/web-common/features/connectors/selectors.ts";

  export let store: ConnectorExplorerStore;
  export let olapOnly: boolean = false;
  /** Auto-expand this connector when the list first renders */
  export let defaultExpanded: string = "";

  const client = useRuntimeClient();

  $: connectors = getAnalyzedConnectors(client, olapOnly);
  $: ({ data, error, isLoading } = $connectors);

  // When defaultExpanded is set, pre-seed the store so only that connector
  // starts expanded and others start collapsed.
  let hasAutoExpanded = false;
  $: if (defaultExpanded && data?.connectors && !hasAutoExpanded) {
    for (const c of data.connectors) {
      if (!c.name) continue;
      // Pre-seed each connector before ConnectorEntry renders.
      // This prevents getDefaultState from expanding all connectors.
      store.store.update((state) => {
        if (c.name! in state.expandedItems) return state;
        return {
          ...state,
          expandedItems: {
            ...state.expandedItems,
            [c.name!]: c.name === defaultExpanded,
          },
        };
      });
    }
    hasAutoExpanded = true;
  }
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
  {:else if isLoading}
    <div class="flex flex-col gap-y-1.5 w-full px-2 py-2">
      {#each [0.6, 0.75, 0.5] as width}
        <div
          class="h-5 bg-gray-200 animate-pulse rounded"
          style:width="{width * 100}%"
        ></div>
      {/each}
    </div>
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
