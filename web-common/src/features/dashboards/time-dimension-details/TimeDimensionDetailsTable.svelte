<script lang="ts">
  import { flexRender } from "@tanstack/svelte-table";
  import Pivot from "./Pivot.svelte";
  import TimeDimensionDetailsTableCell from "./TimeDimensionDetailsTableCell.svelte";
  import TimeDimensionDetailsTableHeaderCell from "./TimeDimensionDetailsTableHeaderCell.svelte";
  import { writable } from "svelte/store";

  let store = writable<{
    highlightedCol: number | null;
    highlightedRow: number | null;
    scrubbedCols: [number, number] | null;
  }>({
    highlightedCol: null,
    highlightedRow: null,
    scrubbedCols: [8, 12],
  });

  // Mock state for now
  let state = {
    getRowSize: () => 35,
    getColumnWidth: (idx: number) => (idx < 6 ? 60 : 100),
    renderCell: (rowIdx, colIdx) =>
      flexRender(TimeDimensionDetailsTableCell, {
        rowIdx,
        colIdx,
        store,
        fixed: colIdx < 6,
      }),
    renderHeaderCell: (rowIdx, colIdx) => flexRender(TimeDimensionDetailsTableHeaderCell, {
      rowIdx,
      colIdx,
      store,
      fixed: colIdx < 6
    })
  };


  // Mock data that is fetched from backend
  let data = {
    data: [],
    metadata: {
      rowCt: 1000,
      fixedColumnCt: 6,
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

  const handleEvt = (evt) =>
    ($store.highlightedCol = parseInt(evt.target.value));
  const handleRow = (evt) =>
    ($store.highlightedRow = parseInt(evt.target.value));
</script>

<!-- @ts-ignore -->
<input class="border" on:input={handleRow} />
<input class="border" on:input={handleEvt} />

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
