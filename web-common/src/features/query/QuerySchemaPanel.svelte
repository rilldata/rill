<script lang="ts">
  import DataTypeIcon from "@rilldata/web-common/components/data-types/DataTypeIcon.svelte";
  import TableSchema from "@rilldata/web-common/features/connectors/explorer/TableSchema.svelte";
  import ResizableSidebar from "@rilldata/web-common/layout/ResizableSidebar.svelte";
  import { formatInteger } from "@rilldata/web-common/lib/formatters";
  import type { V1StructType } from "@rilldata/web-common/runtime-client";
  import { formatExecutionTime, prettyPrintType } from "./query-utils";

  let {
    schema,
    rowCount,
    executionTimeMs,
    selectedTable = null,
  }: {
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

  let fields = $derived(schema?.fields ?? []);
  let resultColumns = $derived(
    fields.map((f) => ({
      name: f.name ?? "",
      type: prettyPrintType(f.type?.code),
    })),
  );
</script>

<ResizableSidebar
  id="query-schema-sidebar"
  minWidth={180}
  maxWidth={400}
  defaultWidth={260}
  additionalClass="overflow-auto bg-surface-subtle border-l"
>
  {#if selectedTable}
    <div class="schema-header">
      <span class="truncate font-medium" title={selectedTable.objectName}>
        {selectedTable.objectName}
      </span>
      <span class="text-fg-secondary text-[11px]">
        {selectedTable.connector}
      </span>
    </div>
    {#if selectedTable.databaseSchema}
      <div class="px-4 pb-1 text-fg-secondary text-[11px] truncate">
        {selectedTable.database}.{selectedTable.databaseSchema}
      </div>
    {/if}
    <div class="section-label">SCHEMA</div>
    <TableSchema
      connector={selectedTable.connector}
      database={selectedTable.database}
      databaseSchema={selectedTable.databaseSchema}
      table={selectedTable.objectName}
      forcedLeftPadding="pl-4"
    />
  {:else if schema}
    <div class="schema-header">
      <span class="font-medium">Query results</span>
      <span class="text-fg-secondary text-[11px]">
        {formatInteger(resultColumns.length)}
        {resultColumns.length === 1 ? "column" : "columns"}
      </span>
    </div>
    <div class="px-4 pb-2 text-fg-secondary text-[11px] flex gap-x-3">
      <span>
        {formatInteger(rowCount)}
        {rowCount === 1 ? "row" : "rows"}
      </span>
      {#if executionTimeMs !== null}
        <span>{formatExecutionTime(executionTimeMs)}</span>
      {/if}
    </div>
    <div class="section-label">SCHEMA</div>
    <ul class="schema-list">
      {#each resultColumns as column (column.name)}
        <li class="schema-entry">
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
  {:else}
    <div class="px-4 py-24 italic text-fg-disabled text-center text-sm">
      Run a query to see schema
    </div>
  {/if}
</ResizableSidebar>

<style lang="postcss">
  .schema-header {
    @apply flex items-center justify-between gap-x-2 px-4 py-2 text-sm;
  }

  .section-label {
    @apply text-[11px] text-fg-secondary font-semibold tracking-wide px-4 py-1 border-t;
  }

  .schema-list {
    @apply flex flex-col py-1;
  }

  .schema-entry {
    @apply flex items-center gap-x-2 px-4 py-1;
  }

  .schema-entry:hover {
    @apply bg-popover-accent;
  }
</style>
