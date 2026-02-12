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
  $: isConnectorReady = !connector?.errorMessage;

  $: databaseSchemasQuery = useListDatabaseSchemas(
    instanceId,
    connectorName,
    undefined,
    isConnectorReady,
  );

  $: ({ data: rawData, error, isLoading } = $databaseSchemasQuery);
  $: data = isConnectorReady ? rawData : undefined;
</script>

<div class="wrapper">
  {#if !isConnectorReady}
    <span
      class="message"
      style="padding-left: calc(24px + var(--explorer-indent-offset, 0px))"
      >Error: {connector.errorMessage}</span
    >
  {:else if isLoading}
    <span
      class="message"
      style="padding-left: calc(24px + var(--explorer-indent-offset, 0px))"
      >Loading tables...</span
    >
  {:else if error}
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
