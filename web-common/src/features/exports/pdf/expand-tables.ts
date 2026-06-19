// Canvas tables/pivots (web-common/.../pivot/PivotTable.svelte) use TanStack
// Virtual, so only the rows in the visible viewport exist in the DOM. To capture
// them we temporarily grow each table's scroll container so the virtualizer
// materializes up to `rowCap` rows, then restore the original styles afterwards.

const FALLBACK_ROW_HEIGHT_PX = 24;

export interface ExpandTablesResult {
  // Restores the original inline styles. Always call this in a `finally`.
  restore: () => void;
  // componentId -> footnote, for tables clipped to the row cap.
  truncatedNotes: Map<string, string>;
}

export function expandTablesForCapture(
  root: ParentNode,
  rowCap: number,
): ExpandTablesResult {
  const restorers: Array<() => void> = [];
  const truncatedNotes = new Map<string, string>();

  root.querySelectorAll<HTMLElement>(".table-wrapper").forEach((wrapper) => {
    const styles = getComputedStyle(wrapper);
    const rowHeight = pxVar(styles, "--row-height") || FALLBACK_ROW_HEIGHT_PX;
    const headerHeight = pxVar(styles, "--total-header-height");

    // Full rendered content height (header + all fetched rows).
    const fullContentHeight = wrapper.scrollHeight;
    const cappedHeight = headerHeight + rowCap * rowHeight;
    const targetHeight = Math.min(fullContentHeight, cappedHeight);
    const clipped = fullContentHeight > cappedHeight + 1;

    const prev = {
      maxHeight: wrapper.style.maxHeight,
      height: wrapper.style.height,
      overflow: wrapper.style.overflow,
    };
    // A taller scroll element makes the virtualizer treat more rows as visible,
    // so it renders them; overflow:visible prevents the cap from clipping.
    wrapper.style.maxHeight = "none";
    wrapper.style.height = `${targetHeight}px`;
    wrapper.style.overflow = "visible";
    restorers.push(() => {
      wrapper.style.maxHeight = prev.maxHeight;
      wrapper.style.height = prev.height;
      wrapper.style.overflow = prev.overflow;
    });

    if (clipped) {
      const totalRows = Math.max(
        0,
        Math.round((fullContentHeight - headerHeight) / rowHeight),
      );
      const componentId = wrapper.closest("article.component-card")?.id;
      if (componentId) {
        truncatedNotes.set(
          componentId,
          `Showing first ${rowCap} of ${totalRows} rows`,
        );
      }
    }
  });

  return {
    restore: () => restorers.forEach((r) => r()),
    truncatedNotes,
  };
}

function pxVar(styles: CSSStyleDeclaration, name: string): number {
  const value = parseFloat(styles.getPropertyValue(name));
  return Number.isFinite(value) ? value : 0;
}
