<script lang="ts">
  import { createEventDispatcher, tick } from "svelte";
  import type { DimensionTableRow } from "./dimension-table-types";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
  import { portal } from "@rilldata/web-common/lib/actions/portal";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import { formatDataTypeAsDuckDbQueryString } from "@rilldata/web-common/lib/formatters";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import DimensionTableHeaderCell from "./DimensionTableHeaderCell.svelte";
  import FormattedDataType from "@rilldata/web-common/components/data-types/FormattedDataType.svelte";
  import { isClipboardApiSupported } from "@rilldata/web-common/lib/actions/shift-click-action";

  type HoveringData = {
    row: number;
    column: string;
    value: string | number | null;
    type: string;
    isHeader: boolean;
  };

  const dispatch = createEventDispatcher();

  const ROW_HEIGHT = 24;
  const COLUMN_WIDTH = 108;
  const SAMPLE_SIZE = 30;
  const MIN_COL_WIDTH = 60;
  const MAX_COL_WIDTH = 400;
  const MAX_COL_EM = 50;
  const MIN_COL_EM = 6;

  export let rows: DimensionTableRow[];
  export let dimensionName: string;
  export let selectedValues: string[];
  export let excludeMode = false;
  export let isBeingCompared = false;
  export let columns: VirtualizedTableColumns[];
  export let sortedColumn: string;
  export let sortedAscending: boolean;

  let container: HTMLDivElement;
  let showTooltip = false;
  let resizing = {
    columnIndex: -1,
    initialPixelWidth: 0,
    initialCursorPosition: 0,
  };

  let hoverPosition: DOMRect;

  let hovering: HoveringData | null = null;
  let scrollTop = 0;
  let scrollLeft = 0;
  let clientHeight = 0;
  let columnWidths = initColumnWidths();
  let clientWidth = 0;

  $: selected = new Set(selectedValues);
  $: rowCount = rows.length;
  $: columnCount = columns.length - 1;

  $: totalRowSize = rowCount * ROW_HEIGHT;
  $: totalColumnSize = columnWidths
    .slice(1)
    .reduce((acc, width) => acc + Number(width.replace("px", "")), 0);

  $: [renderedRows, rowBuffer] = getRenderCount(
    80,
    rowCount,
    clientHeight,
    ROW_HEIGHT,
  );

  $: [renderedColumns, halfColOverScan] = getRenderCount(
    6,
    columnCount,
    clientWidth,
    COLUMN_WIDTH,
  );

  $: startRow = getStart(
    rowCount,
    renderedRows,
    rowBuffer,
    scrollTop,
    ROW_HEIGHT,
  );

  $: startColumn = getStart(
    columnCount,
    renderedColumns,
    halfColOverScan,
    scrollLeft,
    COLUMN_WIDTH,
  );

  import { writable } from "svelte/store";

  // function createVirtualizer(
  //   container: HTMLElement,
  //   rowCount: number,
  //   columnCount: number,
  //   minRowHeight: number,
  //   minColumnWidth: number,
  //   rowBuffer: number,
  //   columnBuffer: number,
  // ) {
  //   const startRow = writable(0);
  //   const startColumn = writable(0);
  //   const renderedColumns = writable(0);
  //   const renderedRows = writable(0);
  //   const paddingTop = writable(0);
  //   const paddingBottom = writable();
  //   const paddingLeft = writable(0);
  //   const paddingRight = writable(0);
  //   const totalWidth = writable(0);
  //   const totalHeight = writable(0);

  //   let previousWidth = 0;
  //   let previousHeight = 0;

  //   let visibleRows = 0;
  //   let visibleColumns = 0;
  //   let previousScrollTop = 0;
  //   let previousScrollLeft = 0;
  //   let totalRowsRendered = 0;
  //   let totalColumnsRendered = 0;
  //   let maxRowStart = 0;
  //   let maxColStart = 0;

  //   let previousStartCol = 0;
  //   let previousStartRow = 0;

  //   const observer = new ResizeObserver((entries) => {
  //     const { width, height } = entries[0].contentRect;
  //     if (previousWidth !== width) {
  //       previousWidth = width;
  //       const newVisibleColumns = Math.floor(width / minColumnWidth);
  //       if (newVisibleColumns === visibleColumns) return;

  //       visibleColumns = Math.floor(width / minColumnWidth);
  //       totalColumnsRendered = Math.min(
  //         visibleColumns + columnBuffer,
  //         columnCount,
  //       );
  //       maxColStart = columnCount - totalColumnsRendered;
  //       paddingRight.set((columnCount - totalColumnsRendered) * minColumnWidth);
  //       renderedColumns.set(totalColumnsRendered);
  //     }

  //     if (previousHeight !== height) {
  //       previousHeight = height;
  //       const newVisibleRows = Math.floor(height / minRowHeight);
  //       if (newVisibleRows === visibleRows) return;
  //       visibleRows = Math.floor(height / minRowHeight);
  //       totalRowsRendered = Math.min(visibleRows + rowBuffer, rowCount);
  //       maxRowStart = rowCount - totalRowsRendered;
  //       renderedRows.set(totalRowsRendered);
  //     }
  //   });

  // function handleScroll(e: MouseEvent & { currentTarget: HTMLElement }) {
  //   const target = e.currentTarget as HTMLElement;
  //   const scrollTop = target.scrollTop;
  //   const scrollLeft = target.scrollLeft;

  // if (scrollTop && scrollTop !== previousScrollTop) {
  //   const scrolledMinRows = Math.floor(scrollTop / minRowHeight);
  //   const turnover = Math.floor(scrolledMinRows / rowBuffer / 2);
  //   const newStart = clamp(0, (turnover * rowBuffer) / 2, maxRowStart);
  //   if (newStart !== previousStartRow) {
  //     paddingTop.set(newStart * minRowHeight);
  //     paddingBottom.set(
  //       (rowCount - totalRowsRendered - newStart) * minRowHeight,
  //     );
  //     startRow.set(newStart);
  //   }
  // }

  // if (scrollLeft && scrollLeft !== previousScrollLeft) {
  //   const scrolledMinCols = Math.floor(scrollLeft / minColumnWidth);
  //   const turnover = Math.floor(scrolledMinCols / columnBuffer / 2);
  //   const newStart = clamp(0, (turnover * columnBuffer) / 2, maxColStart);

  //   if (newStart !== previousStartCol) {
  //     paddingLeft.set(newStart * minColumnWidth);
  //     paddingRight.set(
  //       (columnCount - totalColumnsRendered - newStart) * minColumnWidth,
  //     );
  //     startColumn.set(newStart);
  //   }
  // }
  // }

  //   if (container) {
  //     observer.observe(container);
  //     container.addEventListener("scroll", handleScroll);
  //   }

  //   return [
  //     startRow,
  //     startColumn,
  //     renderedColumns,
  //     renderedRows,
  //     paddingTop,
  //     paddingBottom,
  //     paddingLeft,
  //     paddingRight,
  //     totalWidth,
  //     totalHeight,
  //   ];
  // }

  // $: [
  //   startRow,
  //   startColumn,
  //   renderedColumns,
  //   renderedRows,
  //   paddingTop,
  //   paddingBottom,
  //   paddingLeft,
  //   paddingRight,
  // ] = createVirtualizer(
  //   container,
  //   rowCount,
  //   columnCount,
  //   ROW_HEIGHT,
  //   COLUMN_WIDTH,
  //   1,
  //   1,
  // );

  // Size of "virtualized" table height before and after the real rows
  $: totalRowPadding = (rowCount - renderedRows) * ROW_HEIGHT;
  $: paddingTop = startRow * ROW_HEIGHT;
  $: paddingBottom = totalRowPadding - paddingTop;

  $: totalColumnPadding = (columnCount - renderedColumns) * COLUMN_WIDTH;
  $: paddingRight = startColumn * COLUMN_WIDTH;
  $: paddingLeft = totalColumnPadding - paddingLeft;

  function getRenderCount(
    overscan: number,
    totalCount: number,
    containerSize: number,
    itemSize: number,
  ) {
    const visibleCount = Math.floor(containerSize / itemSize);

    return [Math.min(visibleCount + overscan, totalCount), overscan / 2];
  }

  function getStart(
    totalCount: number,
    renderCount: number,
    halfOverScan: number,
    scrollOffset: number,
    itemSize: number,
  ) {
    const scrolledItems = Math.floor(scrollOffset / itemSize);
    const turnover = Math.floor(scrolledItems / halfOverScan);
    return clamp(0, turnover * halfOverScan, totalCount - renderCount);
  }

  async function handleScroll(
    event: Event & { currentTarget: HTMLDivElement },
  ) {
    showTooltip = false;
    scrollTop = event.currentTarget.scrollTop;
    scrollLeft = event.currentTarget.scrollLeft;

    // This is to fix a weird quirk in Chrome
    await tick();
    container.scrollTo({ top: scrollTop, left: scrollLeft });
  }

  function handleStartResize(
    event: MouseEvent & { currentTarget: HTMLButtonElement },
  ) {
    const parent = event.currentTarget.parentElement;
    if (!parent) return;
    const index = Number(event.currentTarget.dataset.index);

    resizing = {
      columnIndex: index,
      initialPixelWidth: parent.clientWidth,
      initialCursorPosition: event.clientX,
    };

    window.addEventListener("mousemove", handleResize);
    window.addEventListener("mouseup", handleEndResize);
  }

  function clamp(min: number, value: number, max: number) {
    return Math.min(Math.max(min, value), max);
  }

  function handleResize(event: MouseEvent) {
    const delta = event.clientX - resizing.initialCursorPosition;
    columnWidths[resizing.columnIndex] =
      clamp(MIN_COL_WIDTH, resizing.initialPixelWidth + delta, MAX_COL_WIDTH) +
      "px";
  }

  function handleEndResize() {
    window.removeEventListener("mousemove", handleResize);
    window.removeEventListener("mouseup", handleEndResize);
  }

  function handleHover(
    e: MouseEvent & {
      currentTarget: EventTarget & HTMLTableCellElement;
    },
  ) {
    hoverPosition = e.currentTarget.getBoundingClientRect();
    const isHeader = Boolean(e.currentTarget.dataset.header);
    const description = e.currentTarget.dataset.description;

    const columnName = String(e.currentTarget.dataset.column);
    const rowIndex = Number(e.currentTarget.dataset.index);
    const value =
      description ?? (rows[rowIndex][columnName] as string | number | null);
    const type = columns.find((c) => c.name === columnName)?.type ?? "string";

    hovering = {
      row: rowIndex,
      column: columnName,
      type,
      value,
      isHeader,
    };

    showTooltip = true;
  }

  function handleLeave() {
    showTooltip = false;
    hovering = null;
  }

  async function handleMouseDown(
    e: MouseEvent & {
      currentTarget: EventTarget & HTMLTableSectionElement;
    },
  ) {
    if (!hovering) return;
    console.log("click");
    if (e.shiftKey) {
      let exportedValue = formatDataTypeAsDuckDbQueryString(
        hovering.value,
        hovering.type,
      );
      await navigator.clipboard.writeText(exportedValue);
      notifications.send({
        message: `copied value "${exportedValue}" to clipboard`,
      });
      return;
    }

    if (hovering.isHeader) {
      dispatch("column-click", hovering.column);
    } else {
      dispatch("select-item", {
        index: hovering.row,
        meta: e.ctrlKey || e.metaKey,
      });
    }
  }

  function initColumnWidths() {
    let sampleIndexes: number[] = [];

    if (rows.length < 30) {
      sampleIndexes = rows.map((_, i) => i);
    } else {
      // Deterministic pseudo-random sample
      for (let i = 0; i < SAMPLE_SIZE; i++) {
        if (i <= 10) {
          sampleIndexes.push(i);
        } else if (i >= SAMPLE_SIZE - 10) {
          sampleIndexes.push(rows.length - i - 1);
        } else {
          sampleIndexes.push(Math.floor(rows.length / 2) - 5 + i);
        }
      }
    }

    return columns.map(({ label, name }, i) => {
      if (!label) return "auto";
      if (i !== 0) return "108px";

      const columnNameLength = label.length;
      const formattedValueKey = name.startsWith("measure")
        ? `__formatted_${name}`
        : name;

      let totalValueCharacters = 0;

      sampleIndexes.forEach((index) => {
        totalValueCharacters += String(rows[index][formattedValueKey]).length;
      });

      const averageCharacterLength = Math.ceil(
        totalValueCharacters / SAMPLE_SIZE,
      );

      const max = Math.max(columnNameLength, averageCharacterLength);

      if (columnNameLength > averageCharacterLength) {
        return clamp(MIN_COL_EM, max * 0.9, MAX_COL_EM) + "em";
      } else {
        return clamp(MIN_COL_EM, max, MAX_COL_EM) + "em";
      }
    });
  }

  // $: console.log(
  //   $startColumn,
  //   $renderedColumns,
  //   $startRow,
  //   $renderedRows,
  //   $paddingTop,
  //   $paddingBottom,
  //   $paddingLeft,
  //   $paddingRight,
  // );
</script>

<div
  role="presentation"
  class="overflow-scroll w-fit max-w-full flex flex-row gap-x-2"
  bind:clientHeight
  bind:clientWidth
  bind:this={container}
  on:scroll={handleScroll}
  on:mousedown={handleMouseDown}
  on:mouseleave={handleLeave}
>
  <table
    class="ui-copy"
    class:has-selection={selected.size}
    style:width="{totalColumnSize}px"
    style:height="{totalRowSize}px"
  >
    <thead>
      <tr class="h-fit">
        <DimensionTableHeaderCell
          index={0}
          {dimensionName}
          name={dimensionName}
          width={columnWidths[0]}
          label={String(columns[0].label)}
          description={columns[0].description}
          sorted={sortedColumn === dimensionName}
          {sortedAscending}
          on:column-click
          on:mouseenter={handleHover}
        >
          <button
            slot="resizer"
            data-index={0}
            class="absolute h-full -right-2 top-0 w-4 cursor-col-resize"
            on:mousedown={handleStartResize}
          />
        </DimensionTableHeaderCell>

        {#if paddingRight}
          <th style:width="{paddingRight}px" />
        {/if}

        {#each { length: renderedColumns } as _, i (i)}
          {@const index = startColumn + i + 1}
          {@const { name, label, description } = columns[index]}
          {@const sorted = sortedColumn === name}
          <DimensionTableHeaderCell
            {name}
            {index}
            {sorted}
            {description}
            {dimensionName}
            {sortedAscending}
            label={String(label)}
            width={columnWidths[index]}
            on:column-click
            on:mouseenter={handleHover}
          >
            <button
              slot="resizer"
              data-index={index}
              class="absolute h-full right-0 top-0 w-2 cursor-col-resize"
              on:mousedown={handleStartResize}
            />
          </DimensionTableHeaderCell>
        {/each}
        <th style:width="{paddingLeft}px" />
      </tr>
    </thead>

    <tbody>
      {#if paddingTop}
        <tr style:height="{paddingTop}px" />
      {/if}

      {#each { length: renderedRows } as _, index (index)}
        {@const rowIndex = index + startRow}
        {@const cells = rows[rowIndex]}
        {@const isSelected = selected.has(String(cells[columns[0].name]))}

        <tr
          class:selected={isSelected}
          class="row"
          style:height="{ROW_HEIGHT}px"
        >
          <td
            class="dimension-name-cell"
            data-index={rowIndex}
            data-column={dimensionName}
            style:max-width={columnWidths[0]}
            on:mouseenter={handleHover}
          >
            <div
              class:sorted={sortedColumn === dimensionName}
              class="flex flex-row gap-x-1 size-full text-right items-center"
            >
              <span class="h-full aspect-square bg-white items-center flex">
                {#if isSelected}
                  {#if !excludeMode && isBeingCompared}
                    <CheckCircle className={"bg-red-400"} size="18px" />
                  {:else if !excludeMode}
                    <Check size="20px" />
                  {:else if excludeMode}
                    <Cancel size="20px" />
                  {/if}
                {/if}
              </span>
              <p class="truncate size-full text-left label items-center flex">
                {cells[columns[0].name]}
              </p>
            </div>
          </td>

          {#if paddingRight}
            <td />
          {/if}

          {#each { length: renderedColumns } as _, i (i)}
            {@const index = startColumn + i + 1}
            {@const { name, type, total } = columns[index]}
            {@const value = cells[name]}
            {@const formattedValue = cells[`__formatted_${name}`]}
            <td
              data-index={rowIndex}
              data-column={name}
              class="cell"
              class:ui-copy-number={type === "INT"}
              class:sorted={sortedColumn === name}
              style:max-width={columnWidths[i]}
              class:no-data={!value}
              on:mouseenter={handleHover}
            >
              {value ? formattedValue : "no data"}

              {#if total}
                <span
                  class:selected={isSelected}
                  class="bar"
                  style:--bar-size="{(Number(cells[name]) / total) * 100}%"
                >
                </span>
              {/if}
            </td>
          {/each}

          {#if paddingLeft}
            <td />
          {/if}
        </tr>
      {/each}

      <tr style:height="{paddingBottom}px"> </tr>
    </tbody>
  </table>
</div>

{#if showTooltip && hovering}
  <aside
    class="w-fit h-fit absolute -translate-x-1/2 -translate-y-full z-[1000]"
    use:portal
    style:top="{hoverPosition.top - 8}px"
    style:left="{hoverPosition.left + hoverPosition.width / 2}px"
  >
    <TooltipContent maxWidth="360px">
      <TooltipTitle>
        <svelte:fragment slot="name">
          {#if hovering.isHeader}
            {hovering.value}
          {:else}
            <FormattedDataType
              dark
              type={hovering?.type}
              value={hovering?.value}
            />
          {/if}
        </svelte:fragment>
      </TooltipTitle>

      <TooltipShortcutContainer>
        {#if hovering.isHeader}
          <div>Sort column</div>
          <Shortcut>Click</Shortcut>
        {/if}
        {#if isClipboardApiSupported()}
          <div>
            <StackingWord key="shift">Copy</StackingWord> this value to clipboard
          </div>
          <Shortcut>
            <span style="font-family: var(--system);">⇧</span> + Click
          </Shortcut>
        {/if}
      </TooltipShortcutContainer>
    </TooltipContent>
  </aside>
{/if}

<style lang="postcss">
  table {
    @apply p-0 m-0 border-spacing-0 border-separate w-fit;
    @apply font-normal table-fixed cursor-pointer select-none;
    --bar-color: var(--color-primary-100);
  }

  thead {
    @apply sticky top-0 z-20;
  }

  .cell {
    @apply px-2;
  }

  .dimension-name-cell {
    @apply sticky left-0;
    @apply z-10 text-left;
    @apply truncate border-r;
  }

  .sorted {
    @apply bg-gray-50;
  }

  td {
    @apply text-right p-0 relative bg-white;
    font-weight: inherit;
  }

  .row:hover > .cell,
  .row:hover > .dimension-name-cell {
    @apply bg-gray-100;
  }

  .no-data {
    @apply text-gray-400 italic;
    font-size: 0.925em;
  }

  .cell:hover,
  .label:hover {
    filter: brightness(0.95) !important;
  }

  thead {
    @apply outline outline-1 outline-gray-200;
  }

  .has-selection tbody {
    @apply text-gray-400;
    --bar-color: #f0f0f0;
  }

  .selected {
    @apply font-semibold text-black;
    --bar-color: var(--color-primary-200);
  }

  .bar {
    @apply absolute top-0 left-0 z-0 size-full mix-blend-multiply;
    background: linear-gradient(
      to right,
      var(--bar-color) var(--bar-size),
      transparent 0%
    );
  }
</style>
