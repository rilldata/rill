<script lang="ts">
  import {
    createQueries,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";
  import Pivot from "./Pivot.svelte";
  import {
    get2DBlocks,
    getBlock,
    getBlocks,
  } from "../time-dimension-details/util";
  import { createPivotDataProvider } from "./pivot-data-provider";
  import type { PivotPos } from "./types";
  import {
    fetchMockBodyData,
    fetchMockColumnData,
    fetchMockRowData,
    getMetadata,
  } from "./api";
  import { basicPivot } from "./configs";
  import { createColumnDataProvider } from "./column-data-provider";
  import { createRowDataProvider } from "./row-data-provider";
  import { createBodyDataProvider } from "./body-data-provider";

  let pos = { x0: 0, x1: 0, y0: 0, y1: 0 };
  const handlePos = (evt) => {
    pos = evt.detail;
  };

  const config = basicPivot;
  const ROW_BLOCK_SIZE = 50;
  const COL_BLOCK_SIZE = 50;
  $: metadata = getMetadata(config);

  // TODO: block size? where to get from. maybe pivot can provide
  // Row Headers
  const getRowHeaderKeysFromPos = (pos: PivotPos) => {
    const blocks = getBlocks(ROW_BLOCK_SIZE, pos.y0, pos.y1);
    return blocks.map((b) => ["example-pivot-row-header", b[0], b[1]]);
  };
  let uniqueRowKeys = [];
  $: {
    const nextKeys = getRowHeaderKeysFromPos(pos);
    if (JSON.stringify(uniqueRowKeys) !== JSON.stringify(nextKeys))
      uniqueRowKeys = nextKeys;
  }
  $: rowQueries = createQueries(
    uniqueRowKeys.map((queryKey) => ({
      queryKey,
      queryFn: fetchMockRowData(queryKey.slice(1), 1000),
    }))
  );
  const rowHeaderDataProvider = createRowDataProvider(
    rowQueries,
    getRowHeaderKeysFromPos
  );
  $: {
    rowHeaderDataProvider.setQueries(rowQueries);
  }

  // Column Headers
  const getColHeaderKeysFromPos = (pos: PivotPos) => {
    const blocks = getBlocks(COL_BLOCK_SIZE, pos.x0, pos.x1);
    return blocks.map((b) => ["example-pivot-col-header", b[0], b[1]]);
  };
  let uniqueColumnKeys = [];
  $: {
    const nextKeys = getColHeaderKeysFromPos(pos);
    if (JSON.stringify(uniqueColumnKeys) !== JSON.stringify(nextKeys))
      uniqueColumnKeys = nextKeys;
  }
  $: columnQueries = createQueries(
    uniqueColumnKeys.map((queryKey) => ({
      queryKey,
      queryFn: fetchMockColumnData(queryKey.slice(1), config, 1000),
    }))
  );
  const columnHeaderDataProvider = createColumnDataProvider(
    columnQueries,
    getColHeaderKeysFromPos
  );
  $: {
    columnHeaderDataProvider.setQueries(columnQueries);
  }

  // Table Body
  const getBodyKeysFromPos = (pos: PivotPos) => {
    const blocks = get2DBlocks({
      blockSizeX: COL_BLOCK_SIZE,
      blockSizeY: ROW_BLOCK_SIZE,
      ...pos,
    });
    return blocks.map((b) => ["example-pivot-body", b]);
  };
  let uniqueBodyKeys = [];
  $: {
    const nextKeys = getBodyKeysFromPos(pos);
    if (JSON.stringify(uniqueBodyKeys) !== JSON.stringify(nextKeys))
      uniqueBodyKeys = nextKeys;
  }
  $: bodyQueries = createQueries(
    uniqueBodyKeys.map((queryKey) => ({
      queryKey,
      queryFn: fetchMockBodyData(queryKey.at(1), 1000),
    }))
  );
  const bodyDataProvider = createBodyDataProvider(
    bodyQueries,
    getBodyKeysFromPos
  );
  $: {
    bodyDataProvider.setQueries(bodyQueries);
  }
</script>

<Pivot
  on:pos={handlePos}
  {rowHeaderDataProvider}
  {columnHeaderDataProvider}
  {bodyDataProvider}
/>
