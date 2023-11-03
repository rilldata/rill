<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { CHECKMARK_COLORS } from "@rilldata/web-common/features/dashboards/config";
  import Pivot from "@rilldata/web-common/features/dashboards/pivot/Pivot.svelte";
  import type {
    PivotPos,
    PivotRenderCallback,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import {
    SelectedCheckmark,
    ExcludeIcon,
    MeasureArrow,
    PieChart,
    PinIcon,
    PinUnsetIcon,
  } from "@rilldata/web-common/features/dashboards/time-dimension-details/TDDIcons";
  import type { TableData, TablePosition, TDDComparison } from "./types";
  import { getClassForCell } from "@rilldata/web-common/features/dashboards/time-dimension-details/util";
  import { lastKnownPosition } from "./time-dimension-data-store";

  export let dimensionLabel: string;
  export let measureLabel: string;
  export let excludeMode: boolean;
  export let sortDirection: boolean;
  export let highlightedCol: number;
  export let scrubPos: { start: number; end: number };
  export let pinIndex: number;
  export let comparing: TDDComparison;
  export let tableData: TableData;

  /** Formatter for the time axis in the table*/
  export let timeFormatter: (date: Date) => string;

  const dispatch = createEventDispatcher();

  let pivot;

  let rowIdxHover: number | undefined;
  let colIdxHover: number | undefined;

  function getBodyData(pos: PivotPos) {
    return tableData?.body
      ?.slice(pos.x0, pos.x1)
      .map((row) => row.slice(pos.y0, pos.y1));
  }

  export function getRowHeaderData(pos: PivotPos) {
    return tableData?.rowHeaderData?.slice(pos.y0, pos.y1);
  }

  export function getColumnHeaderData(pos: PivotPos) {
    return tableData?.columnHeaderData?.slice(pos.x0, pos.x1);
  }

  const renderCell: PivotRenderCallback = (data) => {
    const classesToAdd = ["text-right"];
    const classesToRemove = [
      "border-b",
      "border-gray-200",
      "!bg-white",
      "!bg-gray-100",
      "!bg-gray-200",
      "!bg-blue-50",
      "!bg-blue-100",
      "!bg-blue-200",
      "!bg-slate-50",
      "!bg-slate-100",
      "!bg-slate-200",
    ];

    if (pinIndex > -1 && comparing === "dimension" && data.y === pinIndex + 1) {
      classesToAdd.push("border-b", "border-gray-200");
    }

    if (comparing === "time" && data.y === 2) {
      classesToAdd.push("border-b", "border-gray-200");
    }

    const isScrubbed =
      scrubPos?.start !== undefined &&
      data.x >= scrubPos.start &&
      data.x <= scrubPos.end - 1;

    const palette = isScrubbed
      ? "scrubbed"
      : data.y === 0
      ? "fixed"
      : "default";

    classesToAdd.push(
      getClassForCell(
        palette,
        rowIdxHover,
        colIdxHover ?? highlightedCol,
        data.y,
        data.x
      )
    );
    // Update DOM with consolidated class operations
    data.element.classList.toggle("font-semibold", Boolean(data.y == 0));
    data.element.classList.remove(...classesToRemove);
    data.element.classList.add(...classesToAdd);
  };

  const renderColumnHeader: PivotRenderCallback = (data) => {
    data.element.classList.add("font-medium", "text-right");
    return timeFormatter(data.value.value);
  };

  // Any time visible line list changes, redraw the table
  $: {
    scrubPos;
    highlightedCol;
    tableData?.selectedValues;
    pivot?.draw();
  }

  let noSelectionMarkerCount = 0;
  const getMarker = (value, y) => {
    if (y === 0) {
      noSelectionMarkerCount = 0;

      if (pinIndex === tableData?.selectedValues.length - 1) return PinIcon;
      else return PinUnsetIcon;
    }
    const visibleIdx = tableData?.selectedValues.indexOf(value.value);

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
      if (excludeMode || !tableData?.selectedValues.length) {
        noSelectionMarkerCount += 1;
        return `<div class="rounded-full bg-${
          CHECKMARK_COLORS[noSelectionMarkerCount - 1]
        }" style="width: 13px; height: 13px;"></div>`;
      }
    }

    return ``;
  };

  const renderRowHeader: PivotRenderCallback = ({ value, x, y, element }) => {
    const showBorder =
      (pinIndex > -1 && comparing === "dimension" && y === pinIndex + 1) ||
      (comparing === "time" && y === 2);
    if (showBorder) {
      element.classList.add("border-b", "border-gray-200");
    } else {
      element.classList.remove("border-b", "border-gray-200");
    }

    const cellBgColor = getClassForCell(
      "fixed",
      rowIdxHover,
      colIdxHover ?? highlightedCol,
      y,
      x - tableData?.fixedColCount
    );
    if (x > 0) {
      element.classList.remove(
        "!bg-slate-50",
        "!bg-slate-100",
        "!bg-slate-200"
      );
      element.classList.add(cellBgColor);
    }
    if (x === 0) {
      element.classList.add("pl-0");
      const marker = getMarker(value, y);
      const pinClass = y === 0 ? "pin" : "";
      const fontWeight = y === 0 ? "font-semibold" : "font-normal";
      return `<div class="flex items-center w-full h-full overflow-hidden pr-2 gap-1">
        <div class="${pinClass} w-5 shrink-0 h-full flex items-center justify-center" marker="${y}">${marker}</div>
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
      return `<div class="text-right font-medium text-gray-700 flex items-center" sortable="true">
        <span class="truncate">${measureLabel} </span>
        ${
          comparing === "dimension" && tableData?.fixedColCount === 2
            ? `<span>${MeasureArrow(sortDirection)}</span>`
            : ``
        }
      </div>`;
    if (data.x === 2)
      return `<div class="flex items-center justify-end text-gray-700" sortable="true">${PieChart} % ${MeasureArrow(
        sortDirection
      )}</div>`;
  };

  let containerWidth;

  /**
   * Compute available width for table columns by subtracting fixed widths
   * from container width along with extra 50px for padding
   */
  $: colWidth = Math.floor(
    (containerWidth - 250 - 130 - 50 - 50) / tableData?.columnCount
  );

  const getColumnWidth = () => {
    if (colWidth) {
      if (colWidth < 75) return 75;
      if (colWidth > 150) return 150;
      else return colWidth;
    }
    return 75;
  };

  const getRowHeaderWidth = (x: number) => {
    if (colWidth > 160) {
      if (x === 0) {
        const dimWidth = 220 + tableData?.columnCount * (colWidth - 150);
        return Math.min(dimWidth, 500);
      } else if (x === 1) {
        return 160;
      }
    }
    return [250, 130, 50][x];
  };

  // Visible line list
  const toggleVisible = (n) => {
    n = parseInt(n);
    if (comparing != "dimension" || n == 0) return;
    const label = tableData?.rowHeaderData[n][0].value;
    dispatch("toggle-filter", label);
  };

  const togglePin = (n) => {
    n = parseInt(n);
    if (n > 0) return;
    dispatch("toggle-pin");
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
    handleEvent(evt, table, "marker", togglePin);
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

  // Add a CSS variable to control the cursor style as table element is not
  // directly available
  $: cssVarStyles = `--cursor: ${
    comparing === "dimension" ? "pointer" : "default"
  }`;
</script>

<div
  bind:clientWidth={containerWidth}
  on:mouseleave={resetHighlight}
  style:height={comparing === "none" ? "80px" : "calc(100% - 50px)"}
  style={cssVarStyles}
  class="w-full relative"
>
  <Pivot
    bind:this={pivot}
    {getRowHeaderData}
    {getColumnHeaderData}
    {getBodyData}
    rowCount={tableData?.rowCount}
    columnCount={tableData?.columnCount}
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

  :global(.pin) {
    cursor: pointer;
    margin-top: 2px;
  }

  :global(regular-table thead th:first-child div) {
    margin-left: 20px; /* w-5 */
  }
</style>
