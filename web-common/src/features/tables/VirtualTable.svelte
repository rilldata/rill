<script lang="ts">
  import { ComponentType, createEventDispatcher } from "svelte";
  import { portal } from "@rilldata/web-common/lib/actions/portal";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import { formatDataTypeAsDuckDbQueryString } from "@rilldata/web-common/lib/formatters";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import FormattedDataType from "@rilldata/web-common/components/data-types/FormattedDataType.svelte";
  import { isClipboardApiSupported } from "@rilldata/web-common/lib/actions/shift-click-action";
  import type {
    V1MetricsViewColumn,
    V1MetricsViewRowsResponseDataItem,
  } from "@rilldata/web-common/runtime-client";
  import { tick } from "svelte";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
  import { writable } from "svelte/store";
  import { initColumnWidths } from "./init-widths";
  import { clamp } from "@rilldata/web-common/lib/clamp";

  type HoveringData = {
    row: number;
    column: string;
    value: string | number | null;
    type: string;
    isHeader: boolean;
  };

  const dispatch = createEventDispatcher();

  const ROW_HEIGHT = 24;
  const MIN_COL_WIDTH = 108;
  const MAX_COL_WIDTH = 400;
  const HEADER_HEIGHT = 36;
  const columnBuffer = 4;
  const rowBuffer = 8;

  export let rows: V1MetricsViewRowsResponseDataItem[];
  export let columns: (VirtualizedTableColumns | V1MetricsViewColumn)[];
  export let valueAccessor: (name: string) => string = (name: string) => name;
  export let columnAccessor: keyof VirtualizedTableColumns = "label";
  export let HeaderCell: ComponentType | null = null;
  export let Cell: ComponentType | null = null;
  export let RowHeader: ComponentType | null = null;
  export let PinnedCell: ComponentType | null = null;
  export let pinnedColumns = new Map<number, number>();
  export let cellBorders = false;
  export let stickyBorders = false;
  export let headerBorders = false;
  export let resizableColumns = false;
  export let minColWidth = MIN_COL_WIDTH;
  export let rowHeight = ROW_HEIGHT;
  export let headerHeight = HEADER_HEIGHT;
  export let selectedIndexes: number[] = [];
  export let sortedColumn: string | null = null;
  export let sortedAscending = true;

  let container: HTMLDivElement;
  let showTooltip = false;
  let resizing: null | {
    columnIndex: number;
    initialPixelWidth: number;
    initialCursorPosition: number;
  } = null;
  let hoverPosition: DOMRect;
  let hovering: HoveringData | null = null;

  let columnWidths = initColumnWidths({
    columns,
    rows,
    maxWidth: MAX_COL_WIDTH,
    minWidth: minColWidth,
    nameAccessor: "name",
    resizableColumns,
  });

  let previousStartCol = 0;

  let previousPaddingLeft = 0;
  let newColStart = 0;
  let previousColumnWidth = 0;
  let firstColumnWidth = columnWidths[newColStart];
  let previousScrollLeft = 0;

  let startColumn = 0;
  let paddingLeft = 0;

  let clientWidth = 0;
  let clientHeight = 0;

  $: rowCount = rows.length;
  $: columnCount = columns.length - 1;

  $: rowChunk = Math.ceil(rowBuffer / 2);

  $: visibleRows = Math.ceil(clientHeight / ROW_HEIGHT);
  $: renderedRows = Math.min(visibleRows + rowBuffer, rowCount);
  $: maxRowStart = rowCount - renderedRows;
  $: scrolledRows = Math.floor(scrollTop / ROW_HEIGHT);
  $: rowTurnover = Math.floor((scrolledRows - rowChunk) / rowChunk);
  $: startRow = clamp(0, rowTurnover * rowChunk, maxRowStart);
  $: paddingTop = startRow * ROW_HEIGHT;

  $: visibleColumns = Math.ceil(clientWidth / MIN_COL_WIDTH);
  $: renderedColumns = Math.min(visibleColumns + columnBuffer, columnCount);
  $: maxColStart = columnCount - renderedColumns;

  $: totalRowSize = rowCount * ROW_HEIGHT;
  $: totalColumnSize = columnWidths
    .slice(1)
    .reduce((acc, width) => acc + width, 0);

  const getWidth = writable((index: number) => columnWidths[index]);

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

  function handleResize(event: MouseEvent) {
    if (!resizing) return;
    const delta = event.clientX - resizing.initialCursorPosition;
    requestAnimationFrame(() => {
      columnWidths[resizing.columnIndex] = clamp(
        MIN_COL_WIDTH,
        resizing.initialPixelWidth + delta,
        MAX_COL_WIDTH,
      );
    });
  }

  function handleEndResize() {
    getWidth.set((index) => columnWidths[index]);

    resizing = null;

    window.removeEventListener("mousemove", handleResize);
    window.removeEventListener("mouseup", handleEndResize);
  }

  function handleHover(
    e: MouseEvent & {
      currentTarget: EventTarget & HTMLTableCellElement;
    },
  ) {
    hoverPosition = e.currentTarget.getBoundingClientRect();
    const isHeader = Boolean(e.currentTarget.tagName === "TH");
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

  let scrollTop = 0;
  let scrollLeft = 0;

  async function handleScroll(
    e: MouseEvent & { currentTarget: HTMLDivElement },
  ) {
    hovering = null;
    const target = e.currentTarget;
    scrollTop = target.scrollTop;
    scrollLeft = target.scrollLeft;

    let newColStart = previousStartCol;

    const xDelta = scrollLeft - previousScrollLeft;

    previousScrollLeft = scrollLeft;

    if (xDelta > 0) {
      while (
        scrollLeft - previousPaddingLeft >= firstColumnWidth &&
        newColStart < maxColStart
      ) {
        previousPaddingLeft += firstColumnWidth;
        newColStart++;
        previousColumnWidth = firstColumnWidth;
        firstColumnWidth = columnWidths[newColStart];
      }
    } else if (xDelta < 0) {
      while (scrollLeft - previousPaddingLeft <= 0 && newColStart > 0) {
        previousPaddingLeft -= previousColumnWidth;
        newColStart--;
        firstColumnWidth = previousColumnWidth;
        previousColumnWidth = columnWidths[newColStart - 1];
      }
    }

    if (newColStart !== startColumn) {
      previousStartCol = newColStart;
      paddingLeft = previousPaddingLeft;

      startColumn = newColStart;
    }

    // // This is to fix a weird quirk in Chrome
    await tick();
    container.scrollTo({ top: scrollTop, left: scrollLeft });
  }
</script>

<div class="size-full overflow-hidden">
  <div
    role="presentation"
    class="table-wrapper"
    class:cursor-col-resize={resizing}
    class:has-selection={selectedIndexes.length > 0}
    style:--row-height="{rowHeight}px"
    style:--header-height="{headerHeight}px"
    style:--cell-borders="{cellBorders ? 1 : 0}px"
    style:--sticky-borders="{stickyBorders ? 1 : 0}px"
    style:--header-borders="{headerBorders ? 1 : 0}px"
    bind:clientWidth
    bind:clientHeight
    bind:this={container}
    on:mousedown={handleMouseDown}
    on:mouseleave={handleLeave}
    on:scroll={handleScroll}
  >
    <table
      class:pointer-events-none={resizing}
      class="relative"
      style:width="{totalColumnSize}px"
      style:height="{totalRowSize}px"
    >
      <thead>
        <tr>
          {#if RowHeader}
            <th style:width="{72}px" class="row-number">
              <RowHeader index={"#"} />
            </th>
          {/if}

          {#each pinnedColumns as [index, position], i (index)}
            {@const sorted = columns[index].name === sortedColumn}
            <th
              id="header-{index}"
              data-index={index}
              data-column={columns[index].name}
              class:last-pinned={i === pinnedColumns.size - 1}
              style:width="{columnWidths[index]}px"
              class="group pinned pinned-header"
              style:left="{position}px"
              on:mouseenter={handleHover}
            >
              {#if HeaderCell}
                <HeaderCell {sorted} {...columns[index]} />
              {:else}
                {columns[index].name}
              {/if}
            </th>
          {/each}

          {#if paddingLeft}
            <th style:width="{paddingLeft}px" />
          {/if}

          {#each { length: renderedColumns } as _, i (i)}
            {@const index = startColumn + i}
            {@const sorted = columns[index].name === sortedColumn}
            <th
              id="header-{i}"
              data-index={index}
              data-column={columns[index].name}
              style:width="{columnWidths[index]}px"
              class="group relative overflow-hidden"
              on:mouseenter={handleHover}
            >
              {#if HeaderCell}
                <HeaderCell {sorted} {...columns[index]} />
              {:else}
                {columns[index].name}
              {/if}
              {#if resizableColumns}
                <button
                  class="absolute top-0 -right-1 w-2 z-10 cursor-col-resize"
                  style:height="{HEADER_HEIGHT}px"
                  data-index={index}
                  on:mousedown={handleStartResize}
                />
              {/if}
            </th>
          {/each}

          <th class="auto-right-pad" />
        </tr>
      </thead>

      <tbody>
        {#if paddingTop}
          <tr style:height="{paddingTop}px" />
        {/if}

        {#each { length: renderedRows } as _, index (index)}
          {@const rowIndex = index + startRow}
          {@const cells = rows[rowIndex]}
          {@const selected = selectedIndexes.includes(rowIndex)}

          <tr class:selected>
            {#if RowHeader}
              <td class="row-number">
                <RowHeader index={rowIndex + 1} />
              </td>
            {/if}

            {#each pinnedColumns as [index, position], i (i)}
              {@const column = columns[index]}
              {@const columnLabel = String(column[columnAccessor])}
              {@const sorted = columns[index].name === sortedColumn}
              <td
                class="pinned"
                class:last-pinned={i === pinnedColumns.size - 1}
                data-index={rowIndex}
                data-column={columnLabel}
                style:left="{position}px"
                on:mouseenter={handleHover}
              >
                {#if PinnedCell}
                  <PinnedCell
                    {sorted}
                    {selected}
                    value={cells[valueAccessor(columnLabel)] ??
                      cells[columnLabel]}
                    type={columns[index].type}
                  />
                {:else}
                  {cells[valueAccessor(columnLabel)]}
                {/if}
              </td>
            {/each}

            {#if paddingLeft}
              <td />
            {/if}

            {#each { length: renderedColumns } as _, i (i)}
              {@const index = startColumn + i}
              {@const column = columns[index]}
              {@const columnLabel = String(column[columnAccessor])}
              {@const sorted = columns[index].name === sortedColumn}

              <td
                data-index={rowIndex}
                data-column={columnLabel}
                on:mouseenter={handleHover}
              >
                {#if Cell}
                  <Cell
                    {...column}
                    {sorted}
                    {selected}
                    value={cells[columnLabel]}
                    formattedValue={cells[valueAccessor(columnLabel)]}
                  />
                {:else}
                  {cells[valueAccessor(columnLabel)]}
                {/if}
              </td>
            {/each}

            <td class="auto-right-pad" />
          </tr>
        {/each}

        <tr class="auto-bottom-pad" />
      </tbody>
    </table>
  </div>
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
    @apply font-normal cursor-pointer select-none;
    @apply table-fixed;
  }

  .table-wrapper {
    @apply overflow-scroll w-fit max-w-full h-fit max-h-full relative bg-white;
  }

  td,
  th {
    @apply truncate p-0 bg-white;
  }

  td {
    border-right-width: var(--cell-borders);
    border-bottom-width: var(--cell-borders);
  }

  th {
    border-bottom-width: var(--sticky-borders);
    border-right-width: var(--header-borders);
  }

  td:first-of-type {
    border-bottom-width: var(--header-borders);
    border-right-width: var(--sticky-borders);
  }

  thead tr {
    height: var(--header-height);
  }

  tbody tr {
    height: var(--row-height);
  }

  .pinned {
    @apply sticky;
  }

  td.pinned {
    @apply z-10;
  }

  th.pinned {
    @apply z-50;
  }

  .row-number {
    @apply sticky left-0 z-10 text-center;
  }

  thead {
    @apply sticky top-0 z-20;
  }

  tr:hover > td {
    @apply bg-gray-100;
  }

  td:not(:first-of-type):hover {
    filter: brightness(0.95) !important;
  }

  .auto-right-pad {
    width: 100%;
  }

  .auto-bottom-pad {
    height: 100%;
  }

  .last-pinned {
    box-shadow: 2px 0 0 0px gray;
  }

  .has-selection tbody {
    @apply text-gray-400;
    --bar-color: #f0f0f0;
  }

  .selected {
    @apply text-black font-bold;
  }
</style>
