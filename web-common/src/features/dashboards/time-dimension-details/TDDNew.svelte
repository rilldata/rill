<script lang="ts">
  import { CHECKMARK_COLORS } from "@rilldata/web-common/features/dashboards/config";
  import Pivot from "@rilldata/web-common/features/dashboards/pivot/Pivot.svelte";
  import type {
    PivotPos,
    PivotRenderCallback,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import {
    CheckCirlce,
    ExcludeIcon,
    MeasureArrow,
    PieChart,
  } from "@rilldata/web-common/features/dashboards/time-dimension-details/constants";

  export let metricViewName: string;
  export let dimensionName: string;
  export let excludeMode: boolean;
  export let comparing;
  export let data;

  /** Formatter for the time axis in the table*/
  export let timeFormatter;

  let pivot;
  let containerWidth;

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
    // if (data.y === 0) {
    //   data.element.classList.add("font-semibold");
    // }
    data.element.classList.add("text-right");
  };

  const renderColumnHeader: PivotRenderCallback = (data) => {
    data.element.classList.add("font-medium", "text-right");
    return timeFormatter(data.value.value);
  };

  // Visible line list
  const toggleVisible = (n: number) => {
    if (comparing != "dimension" || n == 0) return;
    const label = data?.rowHeaderData[n][0].value;
    metricsExplorerStore.toggleFilter(metricViewName, dimensionName, label);
  };

  // Any time visible line list changes changes, redraw the table
  $: {
    data?.selectedValues;
    pivot?.draw();
  }

  const getMarker = (value, y) => {
    // TODO: Use the colors returned by multiple-dim-query to set the marker color

    if (y === 0) return ``;
    const visibleIdx = data?.selectedValues.indexOf(value.value);

    if (visibleIdx > -1) {
      if (excludeMode) return ExcludeIcon;
      else return CheckCirlce("fill-" + CHECKMARK_COLORS[visibleIdx]);
    } else if (!data?.selectedValues.length && y < 4) {
      return `<div class="rounded-full bg-${
        CHECKMARK_COLORS[y - 1]
      }" style="width: 13px; height: 13px;"></div>`;
    }

    return ``;
  };

  const renderRowHeader: PivotRenderCallback = ({ value, x, y, element }) => {
    element.classList.toggle("bg-slate-50", x > 0);
    if (x === 0) {
      element.classList.add("pl-0");
      const marker = getMarker(value, y);
      const justifyTotal = y === 0 ? "justify-end" : "";
      const fontWeight = y === 0 ? "font-semibold" : "font-normal";
      return `<div class="flex items-center w-full h-full overflow-hidden pr-2 gap-1 ${justifyTotal}">
        <div class="w-5 shrink-0 hover:bg-gray-100 h-full flex items-center justify-center" toggle-visible="${y}">${marker}</div>
        <div class="truncate text-xs ${fontWeight} text-gray-700">${value.value}</div></div>`;
    } else if (x === 1)
      return `<div class="text-xs font-semibold text-right text-gray-700 flex items-center justify-end gap-2" >${value.value}
        ${value.spark}

        </div>`;
    else
      return `<div class="text-xs font-normal text-right text-gray-700" >${value.value}</div>`;
  };

  const renderRowCorner: PivotRenderCallback = ({ value, x, y, element }) => {
    element.classList.add("bg-white", "z-10");
    if (x === 0)
      return `<div class="truncate font-medium text-gray-700 text-left">Dimension A</div>`;
    if (x === 1)
      return `<div class="truncate text-right font-medium text-gray-700">Measure A</div>`;
    if (x === 2)
      return `<div class="flex items-center justify-end text-gray-700">${PieChart} % ${MeasureArrow(
        "down"
      )}</div>`;
  };

  const getColumnWidth = (x: number) => {
    return 75;
  };

  const getRowHeaderWidth = (x: number) => {
    // const dimensionColWidth =
    //   containerWidth - data?.columnCount * getColumnWidth(x) - 120 - 50 - 20;
    return [250, 120, 70][x];
  };

  const handleMouseDown = (evt, table) => {
    let currentNode = evt.target;
    let found = currentNode.hasAttribute("toggle-visible");
    while (!found && currentNode !== table) {
      currentNode = currentNode.parentNode;
      found = currentNode.hasAttribute("toggle-visible");
    }
    if (found) {
      const idx = parseInt(currentNode.getAttribute("toggle-visible"));
      toggleVisible(idx);
    }
  };

  // Hack: for some reason, not enough columns are being drawn on first render.
  // Force a second initial render to workaround it.
  $: {
    if (pivot) {
      setTimeout(pivot.draw, 0);
    }
  }
</script>

<div bind:clientWidth={containerWidth} class="h-full w-full relative mb-2">
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
  />
</div>
