<script lang="ts">
  import { CHECKMARK_COLORS } from "@rilldata/web-common/features/dashboards/config";
  import Pivot from "@rilldata/web-common/features/dashboards/pivot/Pivot.svelte";
  import type {
    PivotPos,
    PivotRenderCallback,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import {
    SelectedCheckmark,
    ExcludeIcon,
    MeasureArrow,
    PieChart,
  } from "@rilldata/web-common/features/dashboards/time-dimension-details/TDDIcons";
  import type { TableData, TablePosition, TDDComparison } from "./types";
  import { getClassForCell } from "@rilldata/web-common/features/dashboards/time-dimension-details/util";
  import { createEventDispatcher } from "svelte";
  import { lastKnownPosition } from "./time-dimension-data-store";

  export let metricViewName: string;
  export let dimensionName: string;
  export let dimensionLabel: string;
  export let measureLabel: string;
  export let excludeMode: boolean;
  export let sortDirection: boolean;
  export let highlightedCol: number;
  export let scrubPos: { start: number; end: number };
  export let comparing: TDDComparison;
  export let data: TableData;

  /** Formatter for the time axis in the table*/
  export let timeFormatter: (date: Date) => string;

  const dispatch = createEventDispatcher();

  let pivot;

  let rowIdxHover: number | undefined;
  let colIdxHover: number | undefined;

  function getBodyData(pos: PivotPos) {
    return data?.body
      ?.slice(pos.x0, pos.x1)
      .map((row) => row.slice(pos.y0, pos.y1));
  }

  export function getRowHeaderData(pos: PivotPos) {
    return data?.rowHeaderData?.slice(pos.y0, pos.y1);
  }

  export function getColumnHeaderData(pos: PivotPos) {
    return data?.columnHeaderData?.slice(pos.x0, pos.x1);
  }

  const renderCell: PivotRenderCallback = (data) => {
    data.element.classList.toggle("font-semibold", Boolean(data.y == 0));
    data.element.classList.add("text-right");

    if (data.y === 2) {
      if (comparing === "time")
        data.element.classList.add("border-b", "border-gray-200");
      else {
        data.element.classList.remove("border-b", "border-gray-200");
      }
    }

    const isScrubbed =
      scrubPos?.start !== undefined &&
      data.x >= scrubPos?.start &&
      data.x <= scrubPos?.end - 1;

    const cellBgColor = getClassForCell(
      isScrubbed ? "scrubbed" : "default",
      rowIdxHover,
      colIdxHover === undefined ? highlightedCol : colIdxHover,
      data.y,
      data.x
    );
    data.element.classList.remove(
      "!bg-white",
      "!bg-gray-100",
      "!bg-gray-200",
      "!bg-blue-50",
      "!bg-blue-100",
      "!bg-blue-200"
    );
    data.element.classList.add(cellBgColor);
  };

  const renderColumnHeader: PivotRenderCallback = (data) => {
    data.element.classList.add("font-medium", "text-right");
    return timeFormatter(data.value.value);
  };

  // Visible line list
  const toggleVisible = (n) => {
    n = parseInt(n);
    if (comparing != "dimension" || n == 0) return;
    const label = data?.rowHeaderData[n][0].value;
    metricsExplorerStore.toggleFilter(metricViewName, dimensionName, label);
  };

  // Any time visible line list changes, redraw the table
  $: {
    scrubPos;
    highlightedCol;
    data?.selectedValues;
    pivot?.draw();
  }

  let noSelectionMarkerCount = 0;
  const getMarker = (value, y) => {
    if (y === 0) {
      noSelectionMarkerCount = 0;
      return ``;
    }
    const visibleIdx = data?.selectedValues.indexOf(value.value);

    if (comparing === "time") {
      if (y == 1) return SelectedCheckmark("fill-blue-500");
      else if (y == 2) return SelectedCheckmark("fill-gray-300");
      else return ``;
    }

    if (visibleIdx > -1) {
      if (excludeMode) return ExcludeIcon;
      // Only show colored markers for first 11 selected values
      else
        return SelectedCheckmark(
          "fill-" +
            (visibleIdx < 11 ? CHECKMARK_COLORS[visibleIdx] : "gray-300")
        );
    } else if (noSelectionMarkerCount < 3) {
      if (excludeMode || !data?.selectedValues.length) {
        noSelectionMarkerCount += 1;
        return `<div class="rounded-full bg-${
          CHECKMARK_COLORS[noSelectionMarkerCount - 1]
        }" style="width: 13px; height: 13px;"></div>`;
      }
    }

    return ``;
  };

  const renderRowHeader: PivotRenderCallback = ({ value, x, y, element }) => {
    if (y === 2) {
      if (comparing === "time") {
        element.classList.add("border-b", "border-gray-200");
      } else {
        element.classList.remove("border-b", "border-gray-200");
      }
    }

    const cellBgColor = getClassForCell(
      "fixed",
      rowIdxHover,
      colIdxHover === undefined ? highlightedCol : colIdxHover,
      y,
      x - data?.fixedColCount
    );
    if (x > 0) {
      element.classList.remove("bg-slate-50", "bg-slate-100", "bg-slate-200");
      element.classList.add(cellBgColor);
    }
    if (x === 0) {
      element.classList.add("pl-0");
      const marker = getMarker(value, y);
      const justifyTotal = y === 0 ? "justify-end" : "";
      const fontWeight = y === 0 ? "font-semibold" : "font-normal";
      return `<div class="flex items-center w-full h-full overflow-hidden pr-2 gap-1 ${justifyTotal}">
        <div class="w-5 shrink-0 h-full flex items-center justify-center" toggle-visible="${y}">${marker}</div>
        <div class="truncate text-xs ${fontWeight} text-gray-700">${value.value}</div></div>`;
    } else if (x === 1)
      return `<div class="text-xs font-semibold text-right text-gray-700 flex items-center justify-end gap-2" >${value.value}
        ${value.spark}

        </div>`;
    else
      return `<div class="text-xs font-normal text-right text-gray-700" >${value.value}</div>`;
  };

  const renderRowCorner: PivotRenderCallback = (data) => {
    data.element.classList.add("bg-white", "z-10");
    if (data.x === 0)
      return `<div class="truncate font-medium text-gray-700 text-left">${dimensionLabel}</div>`;
    if (data.x === 1)
      return `<div class="truncate text-right font-medium text-gray-700" sortable="true">${measureLabel}</div>`;
    if (data.x === 2)
      return `<div class="flex items-center justify-end text-gray-700" sortable="true">${PieChart} % ${MeasureArrow(
        sortDirection
      )}</div>`;
  };

  const getColumnWidth = () => {
    return 75;
  };

  const getRowHeaderWidth = (x: number) => {
    return [250, 130, 50][x];
  };

  const handleEvent = (evt, table, attribute, callback) => {
    let currentNode = evt.target;

    let found = currentNode.hasAttribute(attribute);
    while (!found && currentNode !== table) {
      currentNode = currentNode.parentNode;
      found = currentNode.hasAttribute(attribute);
    }
    if (found) {
      const attributeValue = currentNode.getAttribute(attribute);
      callback(attributeValue);
    }
  };

  const handleMouseDown = (evt, table) => {
    handleEvent(evt, table, "__row", toggleVisible);
    handleEvent(evt, table, "sortable", () => dispatch("toggle-sort"));
  };

  const handleMouseHover = (evt, table) => {
    let newRowIdxHover;
    let newColIdxHover;
    if (evt.type === "mouseout") {
      newRowIdxHover = undefined;
      newColIdxHover = undefined;
    } else {
      handleEvent(evt, table, "__row", (n) => (newRowIdxHover = parseInt(n)));
      handleEvent(evt, table, "__col", (n) => (newColIdxHover = parseInt(n)));
    }

    if (newRowIdxHover !== rowIdxHover && newColIdxHover !== colIdxHover) {
      rowIdxHover = newRowIdxHover;
      colIdxHover = newColIdxHover;
      dispatch("highlight", { x: colIdxHover, y: rowIdxHover });
      pivot?.draw();
    }
  };

  function resetHighlight() {
    rowIdxHover = undefined;
    colIdxHover = undefined;
    dispatch("highlight", { x: colIdxHover, y: rowIdxHover });
    pivot?.draw();
  }

  // Scroll to previous position in case of dashboard refresh during reconcile
  let currentPosition: TablePosition;
  let hasScrolled = false;
  let isInitialized;
  function handlePos(e) {
    const pos = e.detail;
    currentPosition = pos;
    isInitialized = pivot?.isInitialized();
    if (!$lastKnownPosition || hasScrolled) lastKnownPosition.set(pos);
  }

  $: if (isInitialized && !hasScrolled && $lastKnownPosition) {
    pivot?.scrollToCell($lastKnownPosition?.x0, $lastKnownPosition?.y0);
    hasScrolled =
      currentPosition?.x0 === $lastKnownPosition?.x0 &&
      currentPosition?.y0 === $lastKnownPosition?.y0;
  }

  // Hack: for some reason, not enough columns are being drawn on first render.
  // Force a second initial render to workaround it.
  $: {
    if (pivot) {
      setTimeout(pivot.draw, 0);
    }
  }
  $: cssVarStyles = `--cursor: ${
    comparing === "dimension" ? "pointer" : "default"
  }`;
</script>

<div
  on:mouseleave={resetHighlight}
  style:height="calc(100% - 50px)"
  style={cssVarStyles}
  class="w-full relative"
>
  <Pivot
    bind:this={pivot}
    {getRowHeaderData}
    {getColumnHeaderData}
    {getBodyData}
    rowCount={data?.rowCount}
    columnCount={data?.columnCount}
    rowHeaderDepth={4}
    columnHeaderDepth={1}
    {renderCell}
    {renderColumnHeader}
    {renderRowHeader}
    {renderRowCorner}
    {getColumnWidth}
    {getRowHeaderWidth}
    onMouseDown={handleMouseDown}
    onMouseHover={handleMouseHover}
    on:pos={handlePos}
  />
</div>

<style>
  /* Define cursor styles */
  :global(regular-table table) {
    cursor: var(--cursor, default);
  }
  :global(regular-table table tbody tr:first-child, regular-table thead) {
    cursor: default;
  }
  :global(regular-table div[sortable="true"]) {
    cursor: pointer;
  }
</style>
