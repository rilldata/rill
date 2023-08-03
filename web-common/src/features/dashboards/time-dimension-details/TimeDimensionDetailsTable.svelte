<script lang="ts">
  import { flexRender } from "@tanstack/svelte-table";
  import Pivot from "./Pivot.svelte";
  import TimeDimensionDetailsTableCell from "./TimeDimensionDetailsTableCell.svelte";

  // Mock state for now
  let state = {
    getRowSize: () => 35,
    getColumnWidth: (idx: number) => (idx < 6 ? 60 : 100),
    renderCell: (rowIdx, colIdx) =>
      flexRender(TimeDimensionDetailsTableCell, { rowIdx, colIdx }),
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
</script>

<Pivot
  height={400}
  rowCt={data.metadata.rowCt}
  colCt={data.metadata.fixedColumnCt + data.metadata.pivotColumnCt}
  fixedColCt={data.metadata.fixedColumnCt}
  getColumnWidth={state.getColumnWidth}
  getRowSize={state.getRowSize}
  renderCell={state.renderCell}
/>
