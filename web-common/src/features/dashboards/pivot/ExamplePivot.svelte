<script lang="ts">
  import { createQuery, useQueryClient } from "@tanstack/svelte-query";
  import Pivot from "./Pivot.svelte";
  import { getBlock } from "../time-dimension-details/util";
  import { createPivotDataProvider } from "./pivot-data-provider";
  import type { PivotPos } from "./types";
  import {
    fetchMockBodyData,
    fetchMockColumnData,
    fetchMockRowData,
    getMetadata,
  } from "./api";
  import { basicPivot } from "./configs";

  let pos = { x0: 0, x1: 0, y0: 0, y1: 0 };
  const handlePos = (evt) => {
    pos = evt.detail;
    // console.log("POS", pos, getBlock(50, pos.y0, pos.y1));
  };

  const config = basicPivot;
  $: metadata = getMetadata(config);

  // TODO: block size? where to get from. maybe pivot can provide

  // Row Headers
  const getRowHeaderKeyFromPos = (pos: PivotPos) => {
    const block = getBlock(50, pos.y0, pos.y1);
    return ["example-pivot-row-header", block[0], block[1]];
  };
  $: rowQuery = createQuery({
    queryKey: getRowHeaderKeyFromPos(pos),
    queryFn: fetchMockRowData(getBlock(50, pos.y0, pos.y1), 1000),
  });

  $: rowHeaderDataProvider = createPivotDataProvider(
    rowQuery,
    getRowHeaderKeyFromPos
  );

  // Column Headers
  const getColHeaderKeyFromPos = (pos: PivotPos) => {
    const block = getBlock(50, pos.x0, pos.x1);
    return ["example-pivot-col-header", block[0], block[1]];
  };
  $: colQuery = createQuery({
    queryKey: getColHeaderKeyFromPos(pos),
    queryFn: fetchMockColumnData(getBlock(50, pos.x0, pos.x1), config, 1000),
  });
  $: columnHeaderDataProvider = createPivotDataProvider(
    colQuery,
    getColHeaderKeyFromPos
  );

  // Table Body
  const getTableBodyKeyFromPos = (pos: PivotPos) => {
    const blockX = getBlock(50, pos.x0, pos.x1);
    const blockY = getBlock(50, pos.y0, pos.y1);
    return [
      "example-pivot-table-body",
      blockX[0],
      blockX[1],
      blockY[0],
      blockY[1],
    ];
  };
  $: bodyQuery = createQuery({
    queryKey: getTableBodyKeyFromPos(pos),
    queryFn: fetchMockBodyData(pos, config, 1000),
  });
  $: bodyDataProvider = createPivotDataProvider(
    bodyQuery,
    getTableBodyKeyFromPos
  );
</script>

<Pivot
  on:pos={handlePos}
  {rowHeaderDataProvider}
  {columnHeaderDataProvider}
  {bodyDataProvider}
/>
