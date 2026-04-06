<script lang="ts">
  import DataTypeIcon from "@rilldata/web-common/components/data-types/DataTypeIcon.svelte";
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import Inspector from "@rilldata/web-common/layout/workspace/Inspector.svelte";
  import InspectorHeaderGrid from "@rilldata/web-common/layout/inspector/InspectorHeaderGrid.svelte";
  import { formatInteger } from "@rilldata/web-common/lib/formatters";
  import type { V1StructType } from "@rilldata/web-common/runtime-client";
  import { useGetTable } from "@rilldata/web-common/features/connectors/selectors";
  import { useRuntimeClient } from "../../runtime-client/v2";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../layout/config";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import { formatExecutionTime, prettyPrintType } from "./query-utils";

  interface ColumnEntry {
    name: string;
    type: string;
  }

  let {
    filePath,
    schema,
    rowCount,
    executionTimeMs,
    selectedTable = null,
  }: {
    filePath: string;
    schema: V1StructType | null;
    rowCount: number;
    executionTimeMs: number | null;
    /** When set, shows table schema for the selected table instead of query results */
    selectedTable?: {
      connector: string;
      database: string;
      databaseSchema: string;
      objectName: string;
    } | null;
  } = $props();

  const runtimeClient = useRuntimeClient();

  // Fetch table schema when a table is selected from the data explorer
  // Always call useGetTable; it disables itself when table is empty
  let tableQuery = $derived(
    useGetTable(
      runtimeClient,
      selectedTable?.connector ?? "",
      selectedTable?.database ?? "",
      selectedTable?.databaseSchema ?? "",
      selectedTable?.objectName ?? "",
    ),
  );

  let tableColumns = $derived(toColumnEntries($tableQuery?.data?.schema));
  let tableLoading = $derived($tableQuery?.isLoading ?? false);
  let tableError = $derived($tableQuery?.error);

  let showColumns = $state(true);
  let showTableColumns = $state(true);

  let fields = $derived(schema?.fields ?? []);
  let resultColumns = $derived(
    fields.map((f) => ({
      name: f.name ?? "",
      type: prettyPrintType(f.type?.code),
    })),
  );
  let columnCount = $derived(fields.length);

  function toColumnEntries(
    tableSchema: Record<string, unknown> | undefined | null,
  ): ColumnEntry[] {
    if (!tableSchema) return [];
    return Object.entries(tableSchema).map(([name, type]) => ({
      name,
      type: prettyPrintType(type as string),
    }));
  }
</script>

<Inspector {filePath}>
  <div class="py-2 flex flex-col gap-y-2">
    {#if selectedTable}
      <InspectorHeaderGrid>
        <svelte:fragment slot="top-left">
          <p class="truncate" title={selectedTable.objectName}>
            {selectedTable.objectName}
          </p>
        </svelte:fragment>
        <svelte:fragment slot="top-right">
          <p class="text-fg-secondary text-[11px]">
            {selectedTable.connector}
          </p>
        </svelte:fragment>
        <svelte:fragment slot="bottom-left">
          {#if selectedTable.databaseSchema}
            <p class="text-fg-secondary text-[11px] truncate">
              {selectedTable.database}.{selectedTable.databaseSchema}
            </p>
          {/if}
        </svelte:fragment>
        <svelte:fragment slot="bottom-right">
          {#if tableColumns.length > 0}
            {formatInteger(tableColumns.length)}
            {tableColumns.length === 1 ? "column" : "columns"}
          {/if}
        </svelte:fragment>
      </InspectorHeaderGrid>

      <hr />

      <div>
        <div class="px-4">
          <CollapsibleSectionTitle
            tooltipText="table columns"
            bind:active={showTableColumns}
          >
            Columns
          </CollapsibleSectionTitle>
        </div>

        {#if showTableColumns}
          <div transition:slide={{ duration: LIST_SLIDE_DURATION }}>
            {#if tableLoading}
              <p class="px-4 py-2 text-fg-secondary text-xs">Loading...</p>
            {:else if tableError}
              <p class="px-4 py-2 text-red-500 text-xs">
                {extractErrorMessage(tableError)}
              </p>
            {:else if tableColumns.length > 0}
              <ul class="flex flex-col">
                {#each tableColumns as column (column.name)}
                  <li class="column-row">
                    <DataTypeIcon type={column.type} suppressTooltip />
                    <span
                      class="truncate text-xs font-mono"
                      title={column.name}
                    >
                      {column.name}
                    </span>
                    <span
                      class="text-fg-secondary text-[10px] ml-auto flex-none uppercase"
                    >
                      {column.type}
                    </span>
                  </li>
                {/each}
              </ul>
            {:else}
              <p class="px-4 py-2 text-fg-secondary text-xs">
                No columns found
              </p>
            {/if}
          </div>
        {/if}
      </div>
    {:else if schema}
      <InspectorHeaderGrid>
        <svelte:fragment slot="top-left">
          <p>Query results</p>
        </svelte:fragment>
        <svelte:fragment slot="top-right">
          {formatInteger(rowCount)}
          {rowCount === 1 ? "row" : "rows"}
        </svelte:fragment>
        <svelte:fragment slot="bottom-left">
          {#if executionTimeMs !== null}
            <p>{formatExecutionTime(executionTimeMs)}</p>
          {/if}
        </svelte:fragment>
        <svelte:fragment slot="bottom-right">
          {formatInteger(columnCount)}
          {columnCount === 1 ? "column" : "columns"}
        </svelte:fragment>
      </InspectorHeaderGrid>

      <hr />

      <div>
        <div class="px-4">
          <CollapsibleSectionTitle
            tooltipText="result columns"
            bind:active={showColumns}
          >
            Result columns
          </CollapsibleSectionTitle>
        </div>

        {#if showColumns}
          <div transition:slide={{ duration: LIST_SLIDE_DURATION }}>
            <ul class="flex flex-col">
              {#each resultColumns as column (column.name)}
                <li class="column-row">
                  <DataTypeIcon type={column.type} suppressTooltip />
                  <span class="truncate text-xs font-mono" title={column.name}>
                    {column.name}
                  </span>
                  <span
                    class="text-fg-secondary text-[10px] ml-auto flex-none uppercase"
                  >
                    {column.type}
                  </span>
                </li>
              {/each}
            </ul>
          </div>
        {/if}
      </div>
    {:else}
      <div class="px-4 py-24 italic text-fg-disabled text-center">
        Run a query to see schema
      </div>
    {/if}
  </div>
</Inspector>

<style lang="postcss">
  .column-row {
    @apply flex items-center gap-x-2 px-4 py-1;
  }

  .column-row:hover {
    @apply bg-popover-accent;
  }
</style>
