<script lang="ts">
  import type { V1MetricsViewRowsResponseDataItem } from "@rilldata/web-common/runtime-client";
  import type { VirtualizedTableColumns } from "../virtualized-table/types";
  import VirtualTable from "../virtualized-table/VirtualTable.svelte";

  export let rows: V1MetricsViewRowsResponseDataItem[];
  export let columnNames: VirtualizedTableColumns[];
  export let name: string;
  export let rowHeight = 36;
  export let rightBorder = false;
</script>

<VirtualTable
  headerBorders
  resizableColumns
  cellBorders
  stickyBorders
  {rightBorder}
  {name}
  {rows}
  {rowHeight}
  minColWidth={120}
  maxColWidth={320}
  columns={columnNames}
/>
