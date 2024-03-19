import { Writable, get, writable } from "svelte/store";
import { tick } from "svelte";
import { clamp } from "@rilldata/web-common/lib/clamp";

export function createVirtualizer(
  container: HTMLElement,
  rowCount: number,
  columnCount: number,
  minRowHeight: number,
  minColumnWidth: number,
  rowBuffer: number,
  columnBuffer: number,
  widthGetter: Writable<(index: number) => number>,
) {
  const startRow = writable(0);
  const startColumn = writable(0);
  const renderedColumns = writable(Math.min(10, columnCount));
  const renderedRows = writable(Math.min(40, rowCount));
  const paddingTop = writable(0);
  const paddingLeft = writable(0);
  const totalWidth = writable(0);
  const totalHeight = writable(0);

  let getWidth = get(widthGetter);

  let previousWidth = 0;
  let previousHeight = 0;

  let visibleRows = 0;
  let visibleColumns = 0;

  let totalRowsRendered = 0;
  let totalColumnsRendered = 0;

  let maxRowStart = 0;
  let maxColStart = 0;

  let previousScrollTop = 0;
  let previousScrollLeft = 0;

  let previousPaddingLeft = 0;
  let previousPaddingTop = 0;

  let previousStartCol = 0;
  let previousStartRow = 0;

  let firstColumnWidth = 0;
  let previousColumnWidth = 0;

  const chunk = rowBuffer / 2;

  widthGetter.subscribe((v) => {
    getWidth = v;
    firstColumnWidth = getWidth(previousStartCol);
  });

  const observer = new ResizeObserver((entries) => {
    const { width, height } = entries[0].contentRect;

    if (previousWidth !== width) {
      previousWidth = width;
      const newVisibleColumns = Math.floor(width / minColumnWidth);
      if (newVisibleColumns === visibleColumns) return;

      visibleColumns = Math.floor(width / minColumnWidth);
      totalColumnsRendered = Math.min(
        visibleColumns + columnBuffer,
        columnCount,
      );
      maxColStart = columnCount - totalColumnsRendered;

      renderedColumns.set(totalColumnsRendered);
    }

    if (previousHeight !== height) {
      previousHeight = height;
      const newVisibleRows = Math.floor(height / minRowHeight);
      if (newVisibleRows === visibleRows) return;
      visibleRows = Math.floor(height / minRowHeight);
      totalRowsRendered = Math.min(visibleRows + rowBuffer, rowCount);
      maxRowStart = rowCount - totalRowsRendered;

      renderedRows.set(totalRowsRendered);
    }
  });

  async function handleScroll(e: MouseEvent & { currentTarget: HTMLElement }) {
    const target = e.currentTarget;
    const { scrollTop, scrollLeft } = target;

    let newColStart = previousStartCol;
    let newRowStart = previousStartRow;

    const yDelta = scrollTop - previousScrollTop;
    const xDelta = scrollLeft - previousScrollLeft;

    previousScrollTop = scrollTop;
    previousScrollLeft = scrollLeft;

    if (yDelta > 0) {
      while (
        scrollTop - previousPaddingTop > minRowHeight * rowBuffer &&
        newRowStart <= maxRowStart
      ) {
        previousPaddingTop += minRowHeight * chunk;
        newRowStart += chunk;
      }
    } else if (yDelta < 0) {
      while (
        scrollTop - previousPaddingTop < (minRowHeight * rowBuffer) / 2 &&
        newRowStart > 0
      ) {
        previousPaddingTop -= minRowHeight * chunk;
        newRowStart -= chunk;
      }
    }

    if (xDelta > 0) {
      while (
        scrollLeft - previousPaddingLeft > firstColumnWidth &&
        newColStart < maxColStart
      ) {
        previousPaddingLeft += firstColumnWidth;
        newColStart++;
        previousColumnWidth = firstColumnWidth;
        firstColumnWidth = getWidth(newColStart);
      }
    } else if (xDelta < 0) {
      while (scrollLeft - previousPaddingLeft < 0 && newColStart >= 0) {
        previousPaddingLeft -= previousColumnWidth;
        newColStart--;
        firstColumnWidth = previousColumnWidth;
        previousColumnWidth = getWidth(newColStart - 1);
      }
    }

    if (newColStart !== previousStartCol) {
      previousStartCol = newColStart;
      paddingLeft.set(previousPaddingLeft);
      startColumn.set(clamp(0, newColStart, maxColStart));
    }

    if (newRowStart !== previousStartRow) {
      previousStartRow = newRowStart;
      paddingTop.set(previousPaddingTop);
      startRow.set(clamp(0, newRowStart, maxRowStart));
    }

    // // This is to fix a weird quirk in Chrome
    await tick();
    container.scrollTo({ top: scrollTop, left: scrollLeft });
  }

  if (container) {
    observer.observe(container);
    container.addEventListener("scroll", handleScroll);
  }

  return [
    startRow,
    startColumn,
    renderedColumns,
    renderedRows,
    paddingTop,
    paddingLeft,
    totalWidth,
    totalHeight,
  ];
}
