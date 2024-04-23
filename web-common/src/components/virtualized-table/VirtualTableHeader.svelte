<script lang="ts">
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
  import { V1MetricsViewColumn } from "@rilldata/web-common/runtime-client";
  import VirtualTableHeaderCell from "./VirtualTableHeaderCell.svelte";
  import type VirtualTableRowHeader from "./VirtualTableRowHeader.svelte";
  import type { ComponentType } from "svelte";
  import type VirtualTableHeaderCellContent from "./VirtualTableHeaderCellContent.svelte";

  export let rowHeaders: boolean;
  export let RowHeader: ComponentType<VirtualTableRowHeader>;
  export let pinnedColumns: Map<number, number>;
  export let columns: (VirtualizedTableColumns | V1MetricsViewColumn)[];
  export let HeaderCell: ComponentType<VirtualTableHeaderCellContent>;
  export let renderedColumns: number;
  export let resizableColumns: boolean;
  export let sortedColumn: string | null;
  export let startColumn: number;
</script>

<thead>
  <tr>
    {#if rowHeaders}
      <th class="row-number">
        <svelte:component this={RowHeader} index={"#"} />
      </th>
    {/if}

    {#each pinnedColumns as [index, position], i (index)}
      {@const { name, type } = columns[index]}
      {@const sorted = name === sortedColumn}
      <VirtualTableHeaderCell
        pinned
        {type}
        {index}
        {sorted}
        {name}
        {position}
        lastPinned={i === pinnedColumns.size - 1}
        {HeaderCell}
        on:click
        on:mousedown
        on:mouseenter
      />
    {/each}

    <th title="left-pad" />

    {#each { length: renderedColumns } as _, i (i)}
      {@const index = startColumn + i}
      {@const { name, type } = columns[index]}
      {@const sorted = name === sortedColumn}

      {#if !pinnedColumns.has(index)}
        <VirtualTableHeaderCell
          {type}
          {index}
          {sorted}
          {name}
          resizable={resizableColumns}
          {HeaderCell}
          on:mouseenter
          on:click
          on:mousedown
        />
      {/if}
    {/each}

    <th title="right-pad" />
  </tr>
</thead>

<style lang="postcss">
  thead {
    @apply sticky top-0 z-20;
  }

  thead tr {
    height: var(--header-height);
  }

  th {
    @apply truncate p-0 bg-white;
    height: var(--header-height);
  }

  .row-number {
    @apply sticky left-0 z-10 text-center;
  }
</style>
