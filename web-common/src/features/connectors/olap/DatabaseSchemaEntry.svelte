<script lang="ts">
  import { Database, Folder } from "lucide-svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import { V1AnalyzedConnector } from "../../../runtime-client";
  import TableEntry from "./TableEntry.svelte";
  import { useTables } from "./selectors";

  export let instanceId: string;
  export let connector: V1AnalyzedConnector;
  export let database: string;
  export let databaseSchema: string;

  let showTables = true;

  $: connectorName = connector?.name as string;
  $: tablesQuery = useTables(
    instanceId,
    connectorName,
    database,
    databaseSchema,
  );
  $: ({ data } = $tablesQuery);

  $: typedData = data as
    | {
        name: string;
        database: string;
        databaseSchema: string;
        hasUnsupportedDataTypes: boolean;
      }[]
    | undefined;
</script>

<li aria-label={`${database}.${databaseSchema}`} class="database-schema-entry">
  <button
    class="database-schema-entry-header {database ? 'pl-[40px]' : 'pl-[22px]'}"
    class:open={showTables}
    on:click={() => (showTables = !showTables)}
  >
    <CaretDownIcon
      className="transform transition-transform text-gray-400 {showTables
        ? 'rotate-0'
        : '-rotate-90'}"
    />
    <!-- Some databases do not have a full "database -> databaseSchema -> table" hierarchy. 
      When there are only two organizational levels,the API returns "databaseSchema -> table". 
      However, in these cases, we should use a Database icon (not a Folder icon) to represent the organizational structure. -->
    {#if !database}
      <Database size="14px" class="shrink-0 text-gray-400" />
    {:else}
      <Folder size="14px" class="shrink-0 text-gray-400" />
    {/if}
    <span class="truncate">
      {databaseSchema}
    </span>
  </button>

  {#if showTables}
    {#if typedData && typedData.length > 0}
      <ol>
        {#each typedData as tableInfo (tableInfo)}
          <TableEntry
            connectorInstanceId={instanceId}
            connector={connectorName}
            {database}
            {databaseSchema}
            table={tableInfo.name}
            hasUnsupportedDataTypes={tableInfo.hasUnsupportedDataTypes}
          />
        {/each}
      </ol>
    {/if}
  {/if}
</li>

<style lang="postcss">
  .database-schema-entry {
    @apply w-full;
    @apply flex flex-col;
  }

  .database-schema-entry-header {
    @apply h-6 pr-2; /* left-padding is set dynamically above */
    @apply flex items-center gap-x-1;
    @apply sticky top-0 z-10;
  }

  button:hover {
    @apply bg-slate-100;
  }

  .database-schema-entry:not(.open) .database-schema-entry-header {
    @apply bg-white;
  }
</style>
