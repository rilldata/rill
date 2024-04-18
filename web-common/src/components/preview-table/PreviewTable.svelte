<script lang="ts">
  import { V1MetricsViewRowsResponseDataItem } from "@rilldata/web-common/runtime-client";
  import type { VirtualizedTableColumns } from "../virtualized-table/types";
  import VirtualTable from "../virtualized-table/VirtualTable.svelte";

  export let rows: V1MetricsViewRowsResponseDataItem[];
  export let columnNames: VirtualizedTableColumns[];
  export let name: string;
  export let rowHeight = 36;
</script>

<VirtualTable
  headerBorders
  resizableColumns
  cellBorders
  stickyBorders
  {name}
  {rows}
  {rowHeight}
  minColWidth={120}
  maxColWidth={320}
  columns={columnNames}
/>
