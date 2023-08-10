<script lang="ts">
  import { flexRender } from "@tanstack/svelte-table";
  import Pivot from "./Pivot.svelte";
  import TimeDimensionDetailsTableCell from "./TimeDimensionDetailsTableCell.svelte";
  import TimeDimensionDetailsTableHeaderCell from "./TimeDimensionDetailsTableHeaderCell.svelte";
  import { createTimeDimensionDetailsStore } from "./time-dimension-details-store";

  // Store of state to share between line chart and table
  let store = createTimeDimensionDetailsStore();

  // Mock state for now
  const FIXED_COL_CT = 6;
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

  // Mock data that is fetched from backend
  let data = {
    data: [],
    metadata: {
      rowCt: 1000,
      fixedColumnCt: FIXED_COL_CT,
      pivotColumnCt: 100,
    },
  };
  for (let r = 0; r < data.metadata.rowCt; r++) {
    const row = new Array(
      data.metadata.fixedColumnCt + data.metadata.pivotColumnCt
    )
      .fill(0)
      .map((d, i) => ({
        row: r,
        col: i,
      }));

    data.data.push(row);
  }
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
