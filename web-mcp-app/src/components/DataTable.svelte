<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { TableModel } from "../lib/adapt";

  let { table }: { table: TableModel } = $props();
</script>

<!-- A scrollable region has to be reachable from the keyboard (WCAG 2.1.1); the a11y rule
     does not know that scroll containers are the documented exception. -->
<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div
  class="table-wrap"
  role="region"
  aria-label={m.mcp_chart_data_table()}
  tabindex="0"
>
  <table class="ui-copy-number">
    <thead>
      <tr>
        {#each table.columns as column (column.key)}
          <th scope="col" class:num={column.kind === "measure"}
            >{column.label}</th
          >
        {/each}
      </tr>
    </thead>
    <tbody>
      {#each table.rows as row, index (index)}
        <tr>
          {#each table.columns as column (column.key)}
            <td class:num={column.kind === "measure"}
              >{column.format(row[column.key])}</td
            >
          {/each}
        </tr>
      {/each}
    </tbody>
  </table>
  {#if table.totalRows > table.rows.length}
    <p class="note">
      {m.mcp_chart_rows_shown({
        shown: table.rows.length,
        total: table.totalRows,
      })}
    </p>
  {/if}
</div>

<style lang="postcss">
  .table-wrap {
    @apply size-full overflow-auto rounded-[var(--radius)] border border-border;
  }
  table {
    @apply w-full border-collapse text-xs;
  }
  th,
  td {
    @apply px-3 py-2 text-left whitespace-nowrap border-b border-border;
  }
  th {
    @apply sticky top-0 z-10 bg-surface-subtle font-semibold text-fg-secondary;
  }
  td {
    @apply text-fg-primary;
  }
  .num {
    @apply text-right;
  }
  tbody tr:last-child td {
    @apply border-b-0;
  }
  .note {
    @apply m-0 px-3 py-2 text-xs text-fg-tertiary;
  }
</style>
