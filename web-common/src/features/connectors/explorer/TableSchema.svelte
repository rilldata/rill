<script lang="ts">
  import Tooltip from "../../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../../components/tooltip/TooltipContent.svelte";
  import { createQueryServiceTableColumns } from "../../../runtime-client";
  import { useTableMetadata } from "../selectors";
  import { runtime } from "../../../runtime-client/runtime-store";

  export let connector: string;
  export let database: string = ""; // The backend interprets an empty string as the default database
  export let databaseSchema: string = ""; // The backend interprets an empty string as the default schema
  export let table: string;
  export let useNewAPI: boolean = false;

  $: ({ instanceId } = $runtime);

  // Use appropriate API based on connector type
  $: legacyColumnsQuery = !useNewAPI
    ? createQueryServiceTableColumns(instanceId, table, {
        connector,
        database,
        databaseSchema,
      })
    : null;

  $: newTableQuery = useNewAPI
    ? useTableMetadata(instanceId, connector, database, databaseSchema, table)
    : null;

  // Normalize data from both APIs
  $: columns = useNewAPI
    ? // New API returns schema as { [columnName]: "type" }
      $newTableQuery?.data?.schema
      ? Object.entries($newTableQuery.data.schema).map(([name, type]) => ({
          name,
          type: type as string,
        }))
      : []
    : // Legacy API returns profileColumns array
      ($legacyColumnsQuery?.data?.profileColumns ?? []);

  $: error = useNewAPI ? $newTableQuery?.error : $legacyColumnsQuery?.error;
  $: isError = useNewAPI
    ? !!$newTableQuery?.error
    : !!$legacyColumnsQuery?.error;
  $: isLoading = useNewAPI
    ? $newTableQuery?.isLoading
    : $legacyColumnsQuery?.isLoading;

  function prettyPrintType(type: string) {
    // If the type starts with "CODE_", remove it
    return type.replace(/^CODE_/, "");
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
  @reference "tailwindcss";

  @reference "tailwindcss";

  .table-schema-list {
    @apply pr-4 py-1.5; /* padding-left is set dynamically above */
    @apply flex flex-col gap-y-0.5;
  }

  .table-schema-entry {
    @apply flex justify-between gap-x-2;
  }
</style>
