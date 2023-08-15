<script lang="ts">
  import Pivot from "./Pivot.svelte";
  import TimeDimensionDetailsTableCell from "./TimeDimensionDetailsTableCell.svelte";
  import TimeDimensionDetailsTableHeaderCell from "./TimeDimensionDetailsTableHeaderCell.svelte";
  import { createTimeDimensionDetailsStore } from "./time-dimension-details-store";
  import { data } from "./mock-data";
  import { setContext } from "svelte";

  // Store of state to share between line chart and table
  let store = createTimeDimensionDetailsStore();

  setContext("tdt-store", {
    store,
    headers: data.headers,
  });

  // TODO: make this dynamic based on the dimension and measure being rendered
  const columnWidths = [120, 126, 64, 64, 64];

  // Mock state for now
  let state = {
    getRowSize: () => 24,
    getColumnWidth: (idx: number) => columnWidths[idx] ?? 60,
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
  cellComponent={TimeDimensionDetailsTableCell}
  headerComponent={TimeDimensionDetailsTableHeaderCell}
/>
