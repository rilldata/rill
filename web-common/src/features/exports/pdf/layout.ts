import type {
  CapturedBlock,
  PdfPageFormat,
  PdfOrientation,
  ResolvedOrientation,
} from "./types";

// Page dimensions in PostScript points (1/72 inch), portrait orientation.
// jsPDF measures in points by default, so these feed directly into it.
const PAGE_SIZES_PT: Record<PdfPageFormat, { width: number; height: number }> =
  {
    a4: { width: 595.28, height: 841.89 },
    letter: { width: 612, height: 792 },
  };

const DEFAULT_MARGIN_PT = 24;
// Vertical gap between canvas rows, in points.
const ROW_GAP_PT = 12;
// A canvas wider than this (in CSS px) is exported as landscape under "auto".
const AUTO_LANDSCAPE_WIDTH_PX = 900;

export interface PaginateOptions {
  contentWidthPx: number;
  format: PdfPageFormat;
  orientation: PdfOrientation;
  marginPt?: number;
}

export interface Placement {
  block: CapturedBlock;
  // 0-based page index.
  page: number;
  xPt: number;
  yPt: number;
  wPt: number;
  hPt: number;
  // For blocks sliced across pages: the source crop within the image, in image
  // pixels. Undefined when the whole block is drawn.
  srcYPx?: number;
  srcHeightPx?: number;
}

export interface PaginationResult {
  pageWidthPt: number;
  pageHeightPt: number;
  marginPt: number;
  pageCount: number;
  orientation: ResolvedOrientation;
  placements: Placement[];
}

export function resolveOrientation(
  orientation: PdfOrientation,
  contentWidthPx: number,
): ResolvedOrientation {
  if (orientation === "auto") {
    return contentWidthPx > AUTO_LANDSCAPE_WIDTH_PX ? "landscape" : "portrait";
  }
  return orientation;
}

// Groups blocks into canvas rows (preserving DOM order within a row) and walks
// them top-to-bottom, scaling the on-screen layout to the page content width.
// A row that would overflow the current page moves wholesale to the next page;
// a single-block row taller than a full page is sliced across pages.
export function paginate(
  blocks: CapturedBlock[],
  opts: PaginateOptions,
): PaginationResult {
  const orientation = resolveOrientation(opts.orientation, opts.contentWidthPx);
  const size = PAGE_SIZES_PT[opts.format];
  const pageWidthPt = orientation === "landscape" ? size.height : size.width;
  const pageHeightPt = orientation === "landscape" ? size.width : size.height;
  const marginPt = opts.marginPt ?? DEFAULT_MARGIN_PT;

  const contentWidthPt = pageWidthPt - 2 * marginPt;
  const contentHeightPt = pageHeightPt - 2 * marginPt;
  const scale =
    opts.contentWidthPx > 0 ? contentWidthPt / opts.contentWidthPx : 1;

  const rows = groupIntoRows(blocks);

  const placements: Placement[] = [];
  let page = 0;
  let cursorYPt = marginPt;

  for (const row of rows) {
    const rowTopPx = Math.min(...row.map((b) => b.yPx));
    const rowHeightPt = Math.max(...row.map((b) => b.heightPx)) * scale;

    // A row taller than a full page must be sliced. We only slice single-block
    // rows (the common case: a tall table); multi-block tall rows are rare and
    // fall through to normal placement on a fresh page.
    if (rowHeightPt > contentHeightPt && row.length === 1) {
      if (cursorYPt > marginPt) {
        page += 1;
        cursorYPt = marginPt;
      }
      const block = row[0];
      const wPt = block.widthPx * scale;
      const xPt = marginPt + block.xPx * scale;
      // How many source pixels fit in one page of content height.
      const pageSrcPx = contentHeightPt / scale;
      let srcYPx = 0;
      while (srcYPx < block.heightPx - 0.5) {
        const srcHeightPx = Math.min(pageSrcPx, block.heightPx - srcYPx);
        placements.push({
          block,
          page,
          xPt,
          yPt: marginPt,
          wPt,
          hPt: srcHeightPx * scale,
          srcYPx,
          srcHeightPx,
        });
        srcYPx += srcHeightPx;
        page += 1;
        cursorYPt = marginPt;
      }
      continue;
    }

    // Move the whole row to the next page if it doesn't fit and isn't already
    // at the top of a page.
    if (
      cursorYPt > marginPt &&
      cursorYPt + rowHeightPt > pageHeightPt - marginPt
    ) {
      page += 1;
      cursorYPt = marginPt;
    }

    for (const block of row) {
      placements.push({
        block,
        page,
        xPt: marginPt + block.xPx * scale,
        yPt: cursorYPt + (block.yPx - rowTopPx) * scale,
        wPt: block.widthPx * scale,
        hPt: block.heightPx * scale,
      });
    }

    cursorYPt += rowHeightPt + ROW_GAP_PT;
  }

  return {
    pageWidthPt,
    pageHeightPt,
    marginPt,
    pageCount: placements.length
      ? Math.max(...placements.map((p) => p.page)) + 1
      : 0,
    orientation,
    placements,
  };
}

// Orders blocks by rowIndex, then by horizontal position within the row.
function groupIntoRows(blocks: CapturedBlock[]): CapturedBlock[][] {
  const byRow = new Map<number, CapturedBlock[]>();
  for (const block of blocks) {
    const row = byRow.get(block.rowIndex);
    if (row) row.push(block);
    else byRow.set(block.rowIndex, [block]);
  }
  return [...byRow.keys()]
    .sort((a, b) => a - b)
    .map((rowIndex) => byRow.get(rowIndex)!.sort((a, b) => a.xPx - b.xPx));
}
