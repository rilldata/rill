<script lang="ts">
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../../layout/config";
  import type { V1AnalyzedConnector } from "../../../runtime-client";
  import DatabaseEntry from "./DatabaseEntry.svelte";
  import { useListDatabaseSchemas } from "../selectors";
  import type { ConnectorExplorerStore } from "./connector-explorer-store";

  export let instanceId: string;
  export let connector: V1AnalyzedConnector;
  export let store: ConnectorExplorerStore;
  export let searchPattern: string = "";

  $: connectorName = connector?.name as string;
  $: hasError = !!connector?.errorMessage;

  $: queryEnabled = !hasError;

  $: databaseSchemasQuery = useListDatabaseSchemas(
    instanceId,
    connectorName,
    undefined,
    queryEnabled,
  );

  $: ({ data: rawData, error, isLoading } = $databaseSchemasQuery);

  // TanStack Query returns cached data even when disabled
  $: data = queryEnabled ? rawData : undefined;
</script>

<div class="wrapper">
  {#if hasError}
    <span
      class="message"
      style="padding-left: calc(24px + var(--explorer-indent-offset, 0px))"
      >Error: {connector.errorMessage}</span
    >
  {:else if isLoading && queryEnabled}
    <span
      class="message"
      style="padding-left: calc(24px + var(--explorer-indent-offset, 0px))"
      >Loading tables...</span
    >
  {:else if error && queryEnabled}
    <span
      class="message"
      style="padding-left: calc(24px + var(--explorer-indent-offset, 0px))"
      >Error: {error.message || error.response?.data?.message}</span
    >
  {:else if data}
    {#if data.length === 0}
      <span
        class="message"
        style="padding-left: calc(24px + var(--explorer-indent-offset, 0px))"
        >No tables found</span
      >
    {:else}
      <ol transition:slide={{ duration }}>
        {#each data as database (database)}
          <DatabaseEntry
            {instanceId}
            {connector}
            {database}
            {store}
            {searchPattern}
          />
        {/each}
      </ol>
    {/if}
  {/if}
</div>

<style lang="postcss">
  .wrapper {
    @apply flex flex-col overflow-y-auto;
  }

  .message {
    @apply pr-3.5 py-2; /* left-padding is set inline above */
    @apply text-fg-secondary;
  }
</style>
