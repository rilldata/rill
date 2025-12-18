<script lang="ts">
  import Tooltip from "../../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../../components/tooltip/TooltipContent.svelte";
  import { useGetTable } from "../selectors";
  import { runtime } from "../../../runtime-client/runtime-store";

  export let connector: string;
  export let database: string = ""; // The backend interprets an empty string as the default database
  export let databaseSchema: string = ""; // The backend interprets an empty string as the default schema
  export let table: string;

  $: ({ instanceId } = $runtime);
  $: newTableQuery = useGetTable(
    instanceId,
    connector,
    database,
    databaseSchema,
    table,
  );

  // New API returns schema as { [columnName]: "type" }
  $: columns = $newTableQuery?.data?.schema
    ? Object.entries($newTableQuery.data.schema).map(([name, type]) => ({
        name,
        type: type as string,
      }))
    : [];

  $: error = $newTableQuery?.error;
  $: isError = !!$newTableQuery?.error;
  $: isLoading = $newTableQuery?.isLoading;

  function prettyPrintType(type: string) {
    // Remove CODE_ prefix and normalize unsupported types to just "UNKNOWN"
    const normalized = type.replace(/^CODE_/, "");
    return normalized.startsWith("UNKNOWN(") ? "UNKNOWN" : normalized;
  }
</script>

<ul class="table-schema-list">
  {#if isError}
    <div class="{database ? 'pl-[78px]' : 'pl-[60px]'} py-1.5 text-gray-500">
      Error loading schema: {error?.response?.data?.message || error?.message}
    </div>
  {:else if isLoading}
    <div class="{database ? 'pl-[78px]' : 'pl-[60px]'} py-1.5 text-gray-500">
      Loading schema...
    </div>
  {:else if columns && columns.length > 0}
    {#each columns as column (column.name)}
      <li class="table-schema-entry {database ? 'pl-[78px]' : 'pl-[60px]'}">
        <Tooltip distance={4}>
          <span class="font-mono truncate">{column.name}</span>
          <TooltipContent slot="tooltip-content">
            {column.name}
          </TooltipContent>
        </Tooltip>
        <span class="uppercase text-gray-800">
          {prettyPrintType(column.type ?? "")}
        </span>
      </li>
    {/each}
  {:else}
    <div class="{database ? 'pl-[78px]' : 'pl-[60px]'} py-1.5 text-gray-500">
      No columns found
    </div>
  {/if}
</ul>

<style lang="postcss">
  .table-schema-list {
    @apply pr-4 py-1.5; /* padding-left is set dynamically above */
    @apply flex flex-col gap-y-0.5;
  }

  .table-schema-entry {
    @apply flex justify-between gap-x-2;
  }
</style>
