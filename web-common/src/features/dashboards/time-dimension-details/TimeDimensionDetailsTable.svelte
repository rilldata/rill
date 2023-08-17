<script lang="ts">
  import Pivot from "./Pivot.svelte";
  import TimeDimensionDetailsTableCell from "./TimeDimensionDetailsTableCell.svelte";
  import TimeDimensionDetailsTableHeaderCell from "./TimeDimensionDetailsTableHeaderCell.svelte";
  import { createTimeDimensionDetailsStore } from "./time-dimension-details-store";
  import { data } from "./mock-data";
  import { createTDTContext } from "./context";
  import {
    FILTER_OVERFLOW_WIDTH,
    HEADER_HEIGHT,
    ROW_HEIGHT,
  } from "./constants";

  // Store of state to share between line chart and table
  let store = createTimeDimensionDetailsStore();

  createTDTContext({
    store,
    headers: data.headers,
  });

  // TODO: make this dynamic based on the dimension and measure being rendered
  const columnWidths = [120, 126, 64, 64, 64];

  // Mock state for now
  let state = {
    getRowSize: () => ROW_HEIGHT,
    getColumnWidth: (idx: number) => columnWidths[idx] ?? 60,
  };

  // Leave padding on table header and body to make space for the filter check/X overflow
  const headerStyle = `padding-left: ${FILTER_OVERFLOW_WIDTH}px;`;
  const bodyStyle = `padding-left: ${FILTER_OVERFLOW_WIDTH}px;`;
</script>

<h1>Store</h1>
<pre style="max-height: 200px; overflow: auto; width: fit-content;">
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
  headerHeight={HEADER_HEIGHT}
  {headerStyle}
  {bodyStyle}
/>
