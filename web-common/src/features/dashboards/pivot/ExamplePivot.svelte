<script lang="ts">
  import { createQueries, useQueryClient } from "@tanstack/svelte-query";
  import Pivot from "./Pivot.svelte";
  import { get2DBlocks, getBlocks } from "../time-dimension-details/util";
  import type { PivotPos } from "./types";
  import {
    fetchMockBodyData,
    fetchMockColumnData,
    fetchMockRowData,
    getMetadata,
  } from "./api";
  import { basicNestedPivot, basicPivot } from "./configs";
  import { transpose2DArray } from "./util";
  import { writable } from "svelte/store";
  import JsonIcon from "@rilldata/web-common/components/icons/JsonIcon.svelte";

  let pos = { x0: 0, x1: 0, y0: 0, y1: 0 };
  const handlePos = (evt) => {
    pos = evt.detail;
  };

  // const config = basicPivot;
  let config = writable(structuredClone(basicPivot));
  const ROW_BLOCK_SIZE = 50;
  const COL_BLOCK_SIZE = 50;
  $: metadata = getMetadata($config);
  let table;

  const queryClient = useQueryClient();
  const cache = queryClient.getQueryCache();

  $: serializedConfig = JSON.stringify($config);

  // Row Headers
  // const getRowHeaderKeysFromPos = (pos: PivotPos) => {
  //   const blocks = getBlocks(ROW_BLOCK_SIZE, pos.y0, pos.y1);
  //   return blocks.map((b) => ["example-pivot-row-header", b[0], b[1]]);
  // };
  const getRowHeaderKeysFromPos = (pos: PivotPos) => {
    const blocks = getBlocks(ROW_BLOCK_SIZE, pos.y0, pos.y1);
    return blocks.map((b) => [
      "example-pivot-row-header",
      serializedConfig,
      b[0],
      b[1],
    ]);
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
      queryFn: fetchMockRowData(queryKey.slice(-2), $config, 1000),
    }))
  );

  const getRowHeaderData = (pos: PivotPos) => {
    // Splice any cached data into this array
    let data = new Array(pos.y1 - pos.y0).fill(null);
    const keys = getRowHeaderKeysFromPos(pos);
    keys.forEach((key) => {
      const cachedBlock =
        cache.find(key)?.state?.data ??
        (getFromPlaceholderCache(key) as
          | {
              block: [number, number];
              data: any[];
            }
          | undefined);
      if (cachedBlock) {
        const b = cachedBlock.block;
        const startingValue = Math.max(b[0], pos.y0);
        const startingValueLocationInBlock = startingValue - b[0];
        const endingValue = Math.min(b[1], pos.y1);
        const endingValueLocationInBlock = endingValue - b[0];
        const valuesToInclude = cachedBlock.data.slice(
          startingValueLocationInBlock,
          endingValueLocationInBlock
        );
        const targetStartPt = Math.max(b[0], pos.y0) - pos.y0;
        data.splice(targetStartPt, valuesToInclude.length, ...valuesToInclude);
      }
    });
    return data;
  };

  $: {
    if (table && $rowQueries) {
      console.count("DRAW ROWS");
      table.draw();
    }
  }

  // Column Headers
  const getColHeaderKeysFromPos = (pos: PivotPos) => {
    const blocks = getBlocks(COL_BLOCK_SIZE, pos.x0, pos.x1);
    return blocks.map((b) => [
      "example-pivot-col-header",
      serializedConfig,
      b[0],
      b[1],
    ]);
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
      queryFn: fetchMockColumnData(queryKey.slice(-2), $config, 1000),
    }))
  );
  const getColumnHeaderData = (pos: PivotPos) => {
    // Splice any cached data into this array
    let data = new Array(pos.x1 - pos.x0).fill(null);

    /*
      Could do more sophisticated logic
      - I have the data for the parent dimension
      - keep that, add placeholders for expanded dimensions
      - replace any placeholders with actual data for expanded if we have it
      - setup queries for any missing expanded sections
      - stitch everything together
    */

    const keys = getColHeaderKeysFromPos(pos);
    keys.forEach((key) => {
      const cachedBlock = cache.find(key)?.state?.data as
        | {
            block: [number, number];
            data: any[];
          }
        | undefined;
      if (cachedBlock) {
        const b = cachedBlock.block;
        const startingValue = Math.max(b[0], pos.x0);
        const startingValueLocationInBlock = startingValue - b[0];
        const endingValue = Math.min(b[1], pos.x1);
        const endingValueLocationInBlock = endingValue - b[0];
        const valuesToInclude = cachedBlock.data.slice(
          startingValueLocationInBlock,
          endingValueLocationInBlock
        );
        const targetStartPt = Math.max(b[0], pos.x0) - pos.x0;
        data.splice(targetStartPt, valuesToInclude.length, ...valuesToInclude);
      }
    });
    return data;
  };
  $: {
    if (table && $columnQueries) {
      // console.count("DRAW COLUMNS");
      table.draw();
    }
  }

  const placeholderCache = new Map();
  const getFromPlaceholderCache = (k: any) =>
    placeholderCache.get(JSON.stringify(k));
  const setPlaceholderCache = (k: any, v: any) =>
    placeholderCache.set(JSON.stringify(k), v);

  // Table Body
  const getBodyKeysFromPos = (pos: PivotPos) => {
    const blocks = get2DBlocks({
      blockSizeX: COL_BLOCK_SIZE,
      blockSizeY: ROW_BLOCK_SIZE,
      ...pos,
    });
    return blocks.map((b) => ["example-pivot-body", serializedConfig, b]);
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
      queryFn: fetchMockBodyData(queryKey.at(-1), 1000),
    }))
  );
  const getBodyData = (pos: PivotPos) => {
    // Splice any cached data into this array. regular-table expects data to be columnar
    const sampleColData = new Array(pos.x1 - pos.x0).fill(null);
    let data = new Array(pos.y1 - pos.y0)
      .fill(null)
      .map(() => sampleColData.slice());
    const keys = getBodyKeysFromPos(pos);
    keys.forEach((key) => {
      const cachedBlock = cache.find(key)?.state?.data as
        | {
            block: {
              x: [number, number];
              y: [number, number];
            };
            data: any[];
          }
        | undefined;
      if (cachedBlock) {
        const rowBlock = cachedBlock.block.y;
        const colBlock = cachedBlock.block.x;
        const targetStartRowIndex = Math.max(rowBlock[0], pos.y0);
        const targetStartRowIndexInBlock = targetStartRowIndex - rowBlock[0];
        const targetEndRowIndex = Math.min(rowBlock[1], pos.y1);
        const targetEndRowIndexInBlock = targetEndRowIndex - rowBlock[0];
        for (
          var i = targetStartRowIndexInBlock;
          i < targetEndRowIndexInBlock;
          i++
        ) {
          const row = cachedBlock.data[i];

          // Determine target columns
          const targetStartColIndex = Math.max(colBlock[0], pos.x0);
          const targetStartColIndexInBlock = targetStartColIndex - colBlock[0];
          const targetEndColIndex = Math.min(colBlock[1], pos.x1);
          const targetEndColIndexInBlock = targetEndColIndex - colBlock[0];
          const colsToInclude = row.slice(
            targetStartColIndexInBlock,
            targetEndColIndexInBlock
          );
          // Splice them in
          const mergedDataRowIndex = i + rowBlock[0] - pos.y0;
          const targetColIndexInSource = targetStartColIndex - pos.x0;
          data[mergedDataRowIndex].splice(
            targetColIndexInSource,
            colsToInclude.length,
            ...colsToInclude
          );
        }
      }
    });
    // regular-table expects data to be columnar, so transpose it
    // TODO: This step can probably be combined with the previous merging step to save time
    // alternatively, the data can be transposed at the time its fetched
    return data.length > 0 ? transpose2DArray(data) : data;
  };
  $: {
    if (table && $bodyQueries) {
      // console.count("DRAW BODY");
      table.draw();
    }
  }

  function handleMouseDown(event: MouseEvent, table) {
    if (event.target.hasAttribute("pivot-expandable")) {
      const meta = table.getMeta(event.target.parentNode);
      config.update((c) => {
        const existingKeys = getRowHeaderKeysFromPos(pos);
        let action = {
          type: "",
          idx: -1,
        };
        if (meta.value.isExpanded) {
          c.expanded = c.expanded.filter((idx) => idx !== meta.value.idx);
          action.type = "collapse";
          action.idx = meta.value.idx;
        } else {
          c.expanded.push(meta.value.idx);
          action.type = "expand";
          action.idx = meta.value.idx;
        }

        // Attempt to use setQueryData to optimistically update, but this then blocks the actual data result. Want an _intermediate_ result in the cache
        const nextKeys = existingKeys.map((existingKey) => [
          ...existingKey.slice(0, 1),
          JSON.stringify(c),
          ...existingKey.slice(2),
        ]);

        const prevData = existingKeys.map((key) =>
          queryClient.getQueryData(key)
        );

        const nextData = prevData.map((cache: any) => {
          let data = structuredClone(cache.data);
          if (action.type === "collapse") {
            const parentRow = data.find((r) => r[0]?.idx === action.idx);
            if (parentRow) {
              parentRow[0].isExpanded = false;
            }
            data = data.filter((r) => r[1]?.parentIdx !== action.idx);
          } else if (action.type === "expand") {
            const targetIdx = data.findIndex((r) => r[0]?.idx === action.idx);
            if (targetIdx > -1) {
              data[targetIdx][0].isExpanded = true;
              data.splice(targetIdx + 1, 0, ["", "LOADING"]);
            }
          }
          return {
            ...cache,
            data,
          };
        });
        nextKeys.forEach((key, i) => {
          // queryClient.setQueryData(key, nextData[i]);
          setPlaceholderCache(key, nextData[i]);
          // queryClient.refetchQueries({ queryKey: key, type: "active" });
        });

        return c;
      });
    }
  }
</script>

<div class="max-h-64 overflow-auto w-fit">
  <pre>
  {JSON.stringify($config, null, 2)}
</pre>
</div>

<button on:click={() => config.set(structuredClone(basicPivot))}
  >Flat rows</button
>
<button on:click={() => config.set(structuredClone(basicNestedPivot))}
  >Nested rows</button
>
<button on:click={() => table.trigger()}>trigger</button>
<!-- {getColumnHeaderData}
  {getBodyData} -->
<Pivot
  bind:api={table}
  on:pos={handlePos}
  onMouseDown={handleMouseDown}
  {getRowHeaderData}
  {getColumnHeaderData}
  {getBodyData}
  rowCount={metadata.rowCt}
  columnCount={metadata.colCt}
  rowHeaderDepth={2}
  columnHeaderDepth={2}
/>
