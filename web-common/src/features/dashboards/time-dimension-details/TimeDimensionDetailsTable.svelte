<script lang="ts">
  import { flexRender } from "@tanstack/svelte-table";
  import Pivot from "./Pivot.svelte";
  import TimeDimensionDetailsTableCell from "./TimeDimensionDetailsTableCell.svelte";
  import TimeDimensionDetailsTableHeaderCell from "./TimeDimensionDetailsTableHeaderCell.svelte";
  import { createTimeDimensionDetailsStore } from "./time-dimension-details-store";
  import { data, fetchData } from "./mock-data";

  // Mock data that is fetched from backend
  const FIXED_COL_CT = data.metadata.fixedColumnCt;

  // Store of state to share between line chart and table
  let store = createTimeDimensionDetailsStore();

  // Mock state for now
  let state = {
    getRowSize: () => 35,
    getColumnWidth: (idx: number) => (idx < FIXED_COL_CT ? 60 : 100),
    renderCell: (rowIdx, colIdx) =>
      flexRender(TimeDimensionDetailsTableCell, {
        rowIdx,
        colIdx,
        store,
        fixed: colIdx < FIXED_COL_CT,
        lastFixed: colIdx === FIXED_COL_CT - 1,
      }),
    renderHeaderCell: (rowIdx, colIdx) =>
      flexRender(TimeDimensionDetailsTableHeaderCell, {
        rowIdx,
        colIdx,
        store,
        fixed: colIdx < FIXED_COL_CT,
        lastFixed: colIdx === FIXED_COL_CT - 1,
      }),
  };
</script>

<h1>Store</h1>
<pre>
  {JSON.stringify($store, null, 2)}
</pre>
<h1>Table</h1>
<Pivot
  height={400}
  rowCt={data.metadata.rowCt}
  colCt={data.metadata.fixedColumnCt + data.metadata.pivotColumnCt}
  fixedColCt={data.metadata.fixedColumnCt}
  getColumnWidth={state.getColumnWidth}
  getRowSize={state.getRowSize}
  renderCell={state.renderCell}
  renderHeaderCell={state.renderHeaderCell}
/>
