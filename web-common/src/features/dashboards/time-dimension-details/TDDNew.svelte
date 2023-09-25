<script lang="ts">
  import { get, writable } from "svelte/store";
  import Pivot from "@rilldata/web-common/features/dashboards/pivot/Pivot.svelte";
  import {
    MOCK_COL_CT,
    MOCK_ROW_CT,
    getBodyData,
    getColumnHeaderData,
    getRowHeaderData,
  } from "../../../../../web-local/src/routes/dev/table_pivot/examples/tdd/data-providers";
  import type { PivotRenderCallback } from "@rilldata/web-common/features/dashboards/pivot/types";

  let pivot;

  const PieChart = `<svg height="1em" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"> <path  d="M14.7497 1.01468C14.4741 0.998446 14.25 1.22384 14.25 1.49998V9.49998L22.25 9.49998C22.5261 9.49998 22.7515 9.27593 22.7353 9.00027C22.6796 8.0549 22.4663 7.12426 22.103 6.24717C21.6758 5.21591 21.0497 4.27887 20.2604 3.48957C19.4711 2.70027 18.5341 2.07417 17.5028 1.647C16.6257 1.2837 15.6951 1.07035 14.7497 1.01468Z"  fill="currentColor" /> <path  fill-rule="evenodd"  clip-rule="evenodd"  d="M12.5 4.49998C12.5 3.82296 11.9419 3.22085 11.2035 3.26706C9.66504 3.36333 8.17579 3.86466 6.88876 4.72462C5.44983 5.68608 4.32832 7.05264 3.66606 8.6515C3.00379 10.2504 2.83051 12.0097 3.16813 13.707C3.50575 15.4044 4.33911 16.9635 5.56282 18.1872C6.78653 19.4109 8.34563 20.2442 10.043 20.5819C11.7403 20.9195 13.4996 20.7462 15.0985 20.0839C16.6973 19.4217 18.0639 18.3001 19.0254 16.8612C19.8853 15.5742 20.3867 14.0849 20.4829 12.5465C20.5291 11.8081 19.927 11.25 19.25 11.25H12.5V4.49998ZM7.72212 5.97182C8.64349 5.35618 9.68996 4.96237 10.7797 4.8152C10.8973 4.79932 11 4.89211 11 5.01075V12.75H18.7392C18.8579 12.75 18.9507 12.8527 18.9348 12.9703C18.7876 14.06 18.3938 15.1065 17.7782 16.0279C16.9815 17.2201 15.8492 18.1494 14.5245 18.6981C13.1997 19.2468 11.742 19.3904 10.3356 19.1107C8.92924 18.8309 7.63741 18.1404 6.62348 17.1265C5.60955 16.1126 4.91905 14.8207 4.63931 13.4144C4.35957 12.008 4.50314 10.5503 5.05188 9.22552C5.60061 7.90076 6.52986 6.76846 7.72212 5.97182Z"  fill="currentColor" /></svg>`;

  const shortFormatDate = new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
  }).format;

  const renderCell: PivotRenderCallback = (data) => {
    data.element.classList.add("text-right");
  };

  const renderColumnHeader: PivotRenderCallback = (data) => {
    data.element.classList.add("font-medium", "text-right");
    return shortFormatDate(data.value.value);
  };

  // Visible line list
  const visible = writable([]);
  const toggleVisible = (n: number) =>
    visible.update((v) => {
      if (v.includes(n)) return v.filter((i) => i !== n);
      else v.push(n);
      return v;
    });
  const visibleColors = [
    "bg-blue-500",
    "bg-orange-500",
    "bg-green-500",
    "bg-red-500",
    "bg-purple-500",
    "bg-pink-500",
    "bg-cyan-500",
    "bg-slate-500",
    "bg-yellow-500",
    "bg-lime-500",
    "bg-violet-500",
    "bg-gray-500",
  ];

  // Any time visible line list changes changes, redraw the table
  $: {
    $visible;
    pivot?.draw();
  }

  const renderRowHeader: PivotRenderCallback = ({ value, x, y, element }) => {
    element.classList.toggle("bg-slate-50", x > 0);
    if (x === 0) {
      element.classList.add("pl-0");
      const visibleRows = get(visible);
      const visibleIdx = visibleRows.indexOf(y);
      const marker =
        visibleIdx > -1
          ? `<div class="rounded-full ${visibleColors[visibleIdx]}" style="width: 13px; height: 13px;"></div>`
          : ``;
      return `<div class="flex items-center w-full h-full overflow-hidden pr-2 gap-1">
        <div class="w-5 shrink-0 hover:bg-gray-100 h-full flex items-center justify-center" toggle-visible="${y}">${marker}</div>
        <div class="truncate text-xs font-normal text-gray-700 ">${value.value}</div></div>`;
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
      return `<div class="truncate text-xs font-bold text-blue-800 bg-blue-50 rounded-lg px-2 text-left flex items-center">Dimension A</div>`;
    if (x === 1)
      return `<div class="text-right font-medium text-gray-700">Measure A</div>`;
    if (x === 2)
      return `<div class="flex items-center justify-end text-gray-700">${PieChart} %</div>`;
  };

  const getColumnWidth = (x: number) => {
    return 60;
  };

  const getRowHeaderWidth = (x: number) => [172, 110, 46][x];

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

<div class="h-full w-full relative mb-2">
  <Pivot
    bind:this={pivot}
    {getRowHeaderData}
    {getColumnHeaderData}
    {getBodyData}
    rowCount={MOCK_ROW_CT}
    columnCount={MOCK_COL_CT}
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
