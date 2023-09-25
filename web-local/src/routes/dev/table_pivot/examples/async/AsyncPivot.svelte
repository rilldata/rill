<script lang="ts">
  import { createQueries, useQueryClient } from "@tanstack/svelte-query";
  import Pivot from "@rilldata/web-common/features/dashboards/pivot/Pivot.svelte";
  import type { PivotRenderCallback } from "@rilldata/web-common/features/dashboards/pivot/types";
  import {
    fetchMockBodyData,
    fetchMockColumnData,
    fetchMockRowData,
    getMetadata,
  } from "./mock-api";

  import { basicNestedPivot, basicPivot } from "../configs";
  import { writable } from "svelte/store";
  import {
    createBodyDataGetter,
    createColumnHeaderDataGetter,
    createRowHeaderDataGetter,
  } from "./data-providers";
  import {
    getBodyKeysFromPos,
    getColHeaderKeysFromPos,
    getRowHeaderKeysFromPos,
  } from "./query-keys";
  import { createMouseDownHandler } from "./mousedown";

  let pos = { x0: 0, x1: 0, y0: 0, y1: 0 };
  const handlePos = (evt) => {
    pos = evt.detail;
  };

  let config = writable(structuredClone(basicPivot));
  $: serializedConfig = JSON.stringify($config);
  const getConfig = () => serializedConfig;
  $: metadata = getMetadata($config);
  const queryClient = useQueryClient();
  const cache = queryClient.getQueryCache();

  // Row Headers
  let uniqueRowKeys = [];
  $: {
    const nextKeys = getRowHeaderKeysFromPos(pos, serializedConfig);
    if (JSON.stringify(uniqueRowKeys) !== JSON.stringify(nextKeys))
      uniqueRowKeys = nextKeys;
  }
  $: rowQueries = createQueries(
    uniqueRowKeys.map((queryKey) => ({
      queryKey,
      queryFn: fetchMockRowData(queryKey.slice(-2), $config, 1000),
    }))
  );
  const getRowHeaderData = createRowHeaderDataGetter({
    getConfig,
    cache,
  });

  let pivot;
  $: {
    if ($rowQueries) {
      // console.count("DRAW ROWS");
      pivot?.draw();
    }
  }

  // Column Headers
  let uniqueColumnKeys = [];
  $: {
    const nextKeys = getColHeaderKeysFromPos(pos, serializedConfig);
    if (JSON.stringify(uniqueColumnKeys) !== JSON.stringify(nextKeys))
      uniqueColumnKeys = nextKeys;
  }
  $: columnQueries = createQueries(
    uniqueColumnKeys.map((queryKey) => ({
      queryKey,
      queryFn: fetchMockColumnData(queryKey.slice(-2), $config, 1000),
    }))
  );
  const getColumnHeaderData = createColumnHeaderDataGetter({
    getConfig,
    cache,
  });
  $: {
    if (pivot && $columnQueries) {
      // console.count("DRAW COLUMNS");
      pivot.draw();
    }
  }

  // Table Body
  let uniqueBodyKeys = [];
  $: {
    const nextKeys = getBodyKeysFromPos(pos, serializedConfig);
    if (JSON.stringify(uniqueBodyKeys) !== JSON.stringify(nextKeys))
      uniqueBodyKeys = nextKeys;
  }
  $: bodyQueries = createQueries(
    uniqueBodyKeys.map((queryKey) => ({
      queryKey,
      queryFn: fetchMockBodyData(queryKey.at(-1), 1000),
    }))
  );
  const getBodyData = createBodyDataGetter({
    getConfig,
    cache,
  });
  $: {
    if (pivot && $bodyQueries) {
      // console.count("DRAW BODY");
      pivot.draw();
    }
  }

  const handleMouseDown = createMouseDownHandler({
    config,
    getPos: () => pos,
    queryClient,
    cache,
  });

  const renderFlatRowHeader: PivotRenderCallback = ({ element }) => {
    element.classList.add("truncate");
  };

  const renderNestedRowHeader: PivotRenderCallback = (data) => {
    data.element.classList.add("truncate");
    if (data.value === "") return data.value;
    if (data.value.expandable) {
      const marker = data.value.isExpanded ? "-" : "+";
      return `<div class='w-full h-full' data-expandable="${data.value.idx}">${marker} ${data.value.text}</div>`;
    }
    return data.value.text;
  };
  $: renderRowHeader =
    $config.rowJoinType === "flat"
      ? renderFlatRowHeader
      : renderNestedRowHeader;

  const getColumnWidth = (x: number) => {
    if (x === 0) return 100;
    return 70;
  };
</script>

<button on:click={() => config.set(structuredClone(basicPivot))}
  >Flat rows</button
>
<button on:click={() => config.set(structuredClone(basicNestedPivot))}
  >Nested rows</button
>
<div class="h-96 w-full relative">
  <Pivot
    bind:this={pivot}
    on:pos={handlePos}
    onMouseDown={handleMouseDown}
    {getRowHeaderData}
    {getColumnHeaderData}
    {getBodyData}
    {renderRowHeader}
    rowCount={metadata.rowCt}
    columnCount={metadata.colCt}
    {getColumnWidth}
    rowHeaderDepth={2}
    columnHeaderDepth={2}
  />
</div>
