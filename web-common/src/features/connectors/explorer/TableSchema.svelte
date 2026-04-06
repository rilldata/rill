<script lang="ts">
  import Tooltip from "../../../components/tooltip/Tooltip.svelte";
  import TooltipContent from "../../../components/tooltip/TooltipContent.svelte";
  import { extractErrorMessage } from "../../../lib/errors";
  import { prettyPrintType } from "../../query/query-utils";
  import { useGetTable } from "../selectors";
  import { useRuntimeClient } from "../../../runtime-client/v2";

  let {
    connector,
    database = "",
    databaseSchema = "",
    table,
    forcedLeftPadding = undefined,
  }: {
    connector: string;
    database?: string; // The backend interprets an empty string as the default database
    databaseSchema?: string; // The backend interprets an empty string as the default schema
    table: string;
    forcedLeftPadding?: string | undefined;
  } = $props();

  const client = useRuntimeClient();

  let newTableQuery = $derived(
    useGetTable(client, connector, database, databaseSchema, table),
  );

  // New API returns schema as { [columnName]: "type" }
  let columns = $derived(
    $newTableQuery?.data?.schema
      ? Object.entries($newTableQuery.data.schema).map(([name, type]) => ({
          name,
          type: type,
        }))
      : [],
  );

  let error = $derived($newTableQuery?.error);
  let isError = $derived(!!$newTableQuery?.error);
  let isLoading = $derived($newTableQuery?.isLoading);

  let leftPadding = $derived(
    forcedLeftPadding ?? (database ? "pl-[78px]" : "pl-[60px]"),
  );
</script>

<ul class="table-schema-list">
  {#if isError}
    <div class="{leftPadding} py-1.5 text-fg-secondary">
      Error loading schema: {extractErrorMessage(error)}
    </div>
  {:else if isLoading}
    <div class="{leftPadding} py-1.5 text-fg-secondary">Loading schema...</div>
  {:else if columns && columns.length > 0}
    {#each columns as column (column.name)}
      <li class="table-schema-entry {leftPadding}">
        <Tooltip distance={4}>
          <span class="font-mono truncate">{column.name}</span>
          <TooltipContent slot="tooltip-content">
            {column.name}
          </TooltipContent>
        </Tooltip>
        <span class="uppercase text-fg-primary">
          {prettyPrintType(column.type ?? "")}
        </span>
      </li>
    {/each}
  {:else}
    <div class="{leftPadding} py-1.5 text-fg-secondary">No columns found</div>
  {/if}
</ul>

<style lang="postcss">
  .table-schema-list {
    @apply pr-4 py-1.5; /* leftPadding-left is set dynamically above */
    @apply flex flex-col gap-y-0.5;
  }

  .table-schema-entry {
    @apply flex justify-between gap-x-2;
  }
</style>
