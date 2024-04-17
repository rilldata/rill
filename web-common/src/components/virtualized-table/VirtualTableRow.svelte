<script lang="ts">
  import type VirtualTableRowHeader from "./VirtualTableRowHeader.svelte";
  import VirtualTableCell from "./VirtualTableCell.svelte";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
  import {
    V1MetricsViewColumn,
    V1MetricsViewRowsResponseDataItem,
  } from "@rilldata/web-common/runtime-client";
  import { ComponentType } from "svelte";

  export let selected: boolean;
  export let rowHeaders: boolean;
  export let rowIndex: number;
  export let pinnedColumns: Map<number, number>;
  export let RowHeader: ComponentType<VirtualTableRowHeader>;
  export let PinnedCell: ComponentType<VirtualTableCell>;
  export let Cell: ComponentType<VirtualTableCell>;
  export let renderedColumns: number;
  export let columns: (VirtualizedTableColumns | V1MetricsViewColumn)[];
  export let columnAccessor: string;
  export let sortedColumn: string | null;
  export let startColumn: number;
  export let cells: V1MetricsViewRowsResponseDataItem;
  export let valueAccessor: (columnLabel: string) => string;
</script>

<tr class:selected>
  {#if rowHeaders}
    <td class="row-number">
      <svelte:component this={RowHeader} index={rowIndex + 1} />
    </td>
  {/if}

  {#each pinnedColumns as [columnIndex, position], i (i)}
    {@const column = columns[columnIndex]}
    {@const columnLabel = String(column[columnAccessor])}
    <td
      class="pinned"
      class:last-pinned={i === pinnedColumns.size - 1}
      data-index={columnIndex}
      data-column={columnLabel}
      style:left="{position}px"
      on:mouseenter
    >
      <svelte:component
        this={PinnedCell}
        sorted={column.name === sortedColumn}
        {selected}
        value={cells[valueAccessor(columnLabel)] ?? cells[columnLabel]}
        type={column.type}
        formattedValue={cells[valueAccessor(columnLabel)]}
      />
    </td>
  {/each}

  <td title="left-pad" />

  {#each { length: renderedColumns } as _, i (i)}
    {@const columnIndex = startColumn + i}
    {@const column = columns[columnIndex]}
    {@const columnLabel = String(column[columnAccessor])}
    {@const sorted = columns[columnIndex].name === sortedColumn}
    {@const pinned = pinnedColumns.has(columnIndex)}
    {#if !pinned}
      <td data-index={rowIndex} data-column={columnLabel} on:mouseenter>
        <svelte:component
          this={Cell}
          {sorted}
          {selected}
          value={cells[columnLabel]}
          formattedValue={cells[valueAccessor(columnLabel)]}
          type={column.type}
        />
      </td>
    {/if}
  {/each}

  <td title="right-pad" />
</tr>

<style lang="postcss">
  td {
    @apply truncate p-0 bg-white;
  }

  :global(.cell-borders td) {
    @apply border-r border-b;
  }

  tr:nth-last-of-type(2) td {
    @apply border-b-0;
  }

  :global(.sticky-borders td:first-of-type) {
    @apply border-r;
  }

  :global(.header-borders td:first-of-type) {
    @apply border-b;
  }

  td:nth-last-of-type(2) {
    @apply border-r-0;
  }

  tr {
    height: var(--row-height);
  }

  .pinned {
    @apply sticky;
  }

  td.pinned {
    @apply z-10;
  }

  .row-number {
    @apply sticky left-0 z-10 text-center;
  }

  tr:hover > td {
    @apply bg-gray-100;
  }

  td:not(:first-of-type):hover {
    filter: brightness(0.95) !important;
  }

  .last-pinned {
    box-shadow: 2px 0 0 0px gray;
  }

  .selected {
    @apply text-black font-bold;
  }
</style>
