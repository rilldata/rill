<script lang="ts">
  import DataTypeIcon from "@rilldata/web-common/components/data-types/DataTypeIcon.svelte";
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import Inspector from "@rilldata/web-common/layout/workspace/Inspector.svelte";
  import InspectorHeaderGrid from "@rilldata/web-common/layout/inspector/InspectorHeaderGrid.svelte";
  import { formatInteger } from "@rilldata/web-common/lib/formatters";
  import type { V1StructType } from "@rilldata/web-common/runtime-client";
  import { createQueryServiceTableColumns } from "@rilldata/web-common/runtime-client";
  import { useGetTable } from "@rilldata/web-common/features/connectors/selectors";
  import ColumnProfile from "@rilldata/web-common/features/column-profile/ColumnProfile.svelte";
  import { runtime } from "../../runtime-client/runtime-store";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../layout/config";

  export let filePath: string;
  export let schema: V1StructType | null;
  export let rowCount: number;
  export let executionTimeMs: number | null;

  /** When set, shows table profiling for the selected table instead of query results */
  export let selectedTable: {
    connector: string;
    database: string;
    databaseSchema: string;
    objectName: string;
  } | null = null;

  $: ({ instanceId } = $runtime);

  // Try profiling API first (works in dev deployments)
  $: profilingQuery = selectedTable
    ? createQueryServiceTableColumns(
        instanceId,
        selectedTable.objectName,
        {
          connector: selectedTable.connector,
          database: selectedTable.database,
          databaseSchema: selectedTable.databaseSchema,
        },
      )
    : null;

  $: profilingAvailable =
    $profilingQuery?.isSuccess &&
    ($profilingQuery?.data?.profileColumns?.length ?? 0) > 0;
  $: profilingError = $profilingQuery?.isError;

  // Fallback: use ConnectorService GetTable (always works)
  $: fallbackQuery =
    selectedTable && profilingError
      ? useGetTable(
          instanceId,
          selectedTable.connector,
          selectedTable.database,
          selectedTable.databaseSchema,
          selectedTable.objectName,
        )
      : null;

  $: fallbackColumns = $fallbackQuery?.data?.schema
    ? Object.entries($fallbackQuery.data.schema).map(([name, type]) => ({
        name,
        type: type as string,
      }))
    : [];
  $: fallbackLoading = $fallbackQuery?.isLoading ?? false;

  // Column count: from profiling or fallback
  $: tableColumnCount = profilingAvailable
    ? ($profilingQuery?.data?.profileColumns?.length ?? 0)
    : fallbackColumns.length;

  let showColumns = true;
  let showTableColumns = true;

  $: fields = schema?.fields ?? [];
  $: columnCount = fields.length;

  function prettyPrintType(code: string | undefined): string {
    if (!code) return "UNKNOWN";
    const normalized = code.replace(/^CODE_/, "");
    return normalized.startsWith("UNKNOWN(") ? "UNKNOWN" : normalized;
  }

  function formatTime(ms: number): string {
    return ms < 1000 ? `${ms}ms` : `${(ms / 1000).toFixed(1)}s`;
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
          {#if tableColumnCount > 0}
            {formatInteger(tableColumnCount)} {tableColumnCount === 1
              ? "column"
              : "columns"}
          {/if}
        </svelte:fragment>
      </InspectorHeaderGrid>

      <hr />

      {#if profilingAvailable}
        <!-- Full RD-style profiling with graphs, stats, sample data -->
        <ColumnProfile
          connector={selectedTable.connector}
          database={selectedTable.database}
          databaseSchema={selectedTable.databaseSchema}
          objectName={selectedTable.objectName}
        />
      {:else if $profilingQuery?.isLoading}
        <p class="px-4 py-2 text-fg-secondary text-xs">Loading...</p>
      {:else}
        <!-- Fallback: simple column list -->
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
              {#if fallbackLoading}
                <p class="px-4 py-2 text-fg-secondary text-xs">Loading...</p>
              {:else if fallbackColumns.length > 0}
                <ul class="flex flex-col">
                  {#each fallbackColumns as column (column.name)}
                    <li class="column-row">
                      <DataTypeIcon
                        type={prettyPrintType(column.type)}
                        suppressTooltip
                      />
                      <span class="truncate text-xs font-mono" title={column.name}>
                        {column.name}
                      </span>
                      <span
                        class="text-fg-secondary text-[10px] ml-auto flex-none uppercase"
                      >
                        {prettyPrintType(column.type)}
                      </span>
                    </li>
                  {/each}
                </ul>
              {:else}
                <p class="px-4 py-2 text-fg-secondary text-xs">No columns found</p>
              {/if}
            </div>
          {/if}
        </div>
      {/if}
    {:else if schema}
      <InspectorHeaderGrid>
        <svelte:fragment slot="top-left">
          <p>Query results</p>
        </svelte:fragment>
        <svelte:fragment slot="top-right">
          {formatInteger(rowCount)} {rowCount === 1 ? "row" : "rows"}
        </svelte:fragment>
        <svelte:fragment slot="bottom-left">
          {#if executionTimeMs !== null}
            <p>{formatTime(executionTimeMs)}</p>
          {/if}
        </svelte:fragment>
        <svelte:fragment slot="bottom-right">
          {formatInteger(columnCount)} {columnCount === 1
            ? "column"
            : "columns"}
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
              {#each fields as field (field.name)}
                <li class="column-row">
                  <DataTypeIcon
                    type={prettyPrintType(field.type?.code)}
                    suppressTooltip
                  />
                  <span class="truncate text-xs font-mono" title={field.name}>
                    {field.name}
                  </span>
                  <span
                    class="text-fg-secondary text-[10px] ml-auto flex-none uppercase"
                  >
                    {prettyPrintType(field.type?.code)}
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
