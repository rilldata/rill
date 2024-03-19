<script lang="ts">
  import VirtualTable from "../table/VirtualTable.svelte";
  import HeaderCell from "./HeaderCell.svelte";
  import Cell from "./Cell.svelte";
  import RowHeader from "./RowHeader.svelte";
  import type { VirtualizedTableColumns } from "../virtualized-table/types";
  import type { V1MetricsViewColumn } from "@rilldata/web-common/runtime-client";

  export let rows;
  export let columns: (VirtualizedTableColumns | V1MetricsViewColumn)[];
</script>

<VirtualTable
  columnAccessor="name"
  {rows}
  {columns}
  cellBorders
  stickyBorders
  resizableColumns
  headerBorders
  rowHeight={32}
  {HeaderCell}
  {Cell}
  {RowHeader}
/>
