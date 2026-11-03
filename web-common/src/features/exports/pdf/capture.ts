import { getFontEmbedCSS } from "html-to-image";
import {
  needsCanvasWarmup,
  rasterizeNode,
  type RasterFormat,
} from "../rasterize";
import {
  FILTER_BAR_ID,
  FILTER_BAR_ROW_INDEX,
  type CapturedBlock,
} from "./types";

export interface CaptureResult {
  blocks: CapturedBlock[];
  contentWidthPx: number;
  backgroundColor: string;
}

export interface CaptureOptions {
  instanceId: string;
  canvasName: string;
  includeFilters: boolean;
  // JPEG (the default) for blocks bound for a PDF; PNG for standalone images.
  format?: RasterFormat;
  onProgress?: (ratio: number) => void;
}

// The off-screen export render (see CanvasPdfExportView) of one canvas, mounted
// only while exporting. Capturing a dedicated tree keeps the live dashboard
// untouched. The lookup is scoped to the canvas store (keyed by instance +
// canvas name) so a second export view (if another is mounted on the page)
// can't be captured by mistake.
export function findExportView(
  instanceId: string,
  canvasName: string,
): { exportView: HTMLElement; rowContainer: HTMLElement } {
  const exportView = Array.from(
    document.querySelectorAll<HTMLElement>("#canvas-pdf-export-view"),
  ).find(
    (el) =>
      el.dataset.instanceId === instanceId &&
      el.dataset.canvasName === canvasName,
  );
  const rowContainer = exportView?.querySelector<HTMLElement>(".row-container");

  if (!exportView || !rowContainer) {
    throw new Error(
      "Canvas content is not available to export. Make sure all required filters are set.",
    );
  }
  return { exportView, rowContainer };
}

// A block of the export render measured but not yet rasterized: its element,
// position and size relative to the content area in CSS pixels, its row, and
// the tab it belongs to (see StaticCanvasRow's tab attributes).
export interface MeasuredBlock {
  id: string;
  element: HTMLElement;
  xPx: number;
  yPx: number;
  widthPx: number;
  heightPx: number;
  rowIndex: number;
  // The label band above an exported tab (see CanvasPdfExportTab) rather than a component card.
  isTabLabel: boolean;
  tabGroup?: string;
  tab?: string;
}

// Measures every capture target of the export render in document order.
export function measureCanvasBlocks(rowContainer: HTMLElement): MeasuredBlock[] {
  const contentRect = rowContainer.getBoundingClientRect();
  return captureTargetsIn(rowContainer).map((target) => {
    const rect = target.getBoundingClientRect();
    const section = target.closest("section");
    return {
      id: target.id,
      element: target,
      xPx: rect.left - contentRect.left,
      yPx: rect.top - contentRect.top,
      widthPx: rect.width,
      heightPx: rect.height,
      rowIndex: rowIndexFor(target, rowContainer),
      isTabLabel: target.classList.contains("pdf-tab-label-row"),
      tabGroup: section?.dataset.tabGroup,
      tab: section?.dataset.tab,
    };
  });
}

// Rasterizes the filter bar (optional) and each canvas block into images
// positioned relative to the canvas content area: every component card plus the
// label band above each exported tab (see captureTargetsIn). Per-block failures
// degrade to a skipped block rather than aborting the whole export.
export async function captureCanvasBlocks(
  opts: CaptureOptions,
): Promise<CaptureResult> {
  const { exportView, rowContainer } = findExportView(
    opts.instanceId,
    opts.canvasName,
  );

  const contentWidthPx = rowContainer.clientWidth;
  const backgroundColor = getComputedStyle(exportView).backgroundColor;

  const targets = measureCanvasBlocks(rowContainer);

  // Probed once per page load, not per block or per export: the answer is a
  // property of the browser, and the probe itself rasterizes.
  const warmUpCanvas = await needsCanvasWarmup();
  // Collected from the whole export view rather than the rows: the header is a
  // sibling of the row container, and getFontEmbedCSS keeps only the @font-face
  // rules whose family is used inside the node it is handed.
  const fontEmbedCSS = await getFontEmbedCSS(exportView);

  const blocks: CapturedBlock[] = [];
  const total = targets.length + (opts.includeFilters ? 1 : 0);
  let done = 0;
  const reportProgress = () => opts.onProgress?.(total ? done / total : 1);

  if (opts.includeFilters) {
    // Read-only summary block (title + exact time range + filter chips),
    // rendered inside the export view specifically for capture; see
    // CanvasPdfExportHeader.
    const header = exportView.querySelector<HTMLElement>(
      "#canvas-pdf-export-header",
    );
    if (header) {
      // Match the header's width to the content area so it scales consistently
      // with the component blocks during pagination.
      header.style.width = `${contentWidthPx}px`;
      if (header.scrollHeight > 0) {
        try {
          const dataUrl = await rasterizeNode(header, {
            backgroundColor,
            fontEmbedCSS,
            warmUpCanvas,
            format: opts.format,
          });
          blocks.push({
            id: FILTER_BAR_ID,
            dataUrl,
            xPx: 0,
            yPx: 0,
            widthPx: contentWidthPx,
            heightPx: header.scrollHeight,
            rowIndex: FILTER_BAR_ROW_INDEX,
          });
        } catch (e) {
          console.warn("Failed to capture canvas header for PDF export", e);
        }
      }
    }
    done += 1;
    reportProgress();
  }

  for (const target of targets) {
    try {
      const dataUrl = await rasterizeNode(target.element, {
        backgroundColor,
        fontEmbedCSS,
        warmUpCanvas,
        format: opts.format,
      });
      blocks.push({
        id: target.id,
        dataUrl,
        xPx: target.xPx,
        yPx: target.yPx,
        widthPx: target.widthPx,
        heightPx: target.heightPx,
        rowIndex: target.rowIndex,
      });
    } catch (e) {
      console.warn(`Failed to capture canvas block "${target.id}"`, e);
    }
    done += 1;
    reportProgress();
  }

  return { blocks, contentWidthPx, backgroundColor };
}

// The units to rasterize, in document order: every component card (top-level
// rows and exported tab rows alike, since the export view flattens tab groups
// into plain rows) plus the label band above each exported tab (see
// CanvasPdfExportTab). Capturing the bands keeps it visible in the PDF which
// tab the rows below belong to.
export function captureTargetsIn(rowContainer: HTMLElement): HTMLElement[] {
  return Array.from(
    rowContainer.querySelectorAll<HTMLElement>(
      "article.component-card, section.pdf-tab-label-row",
    ),
  );
}

// Canvas rows (and tab label bands) are <section> elements; use the nearest
// section's DOM order as the row index so blocks in the same row are grouped
// and laid out together. A tab label band reports the index of the row section
// that follows it, so the label and the tab's first row paginate as one unit
// and the label can never be stranded alone at the bottom of a page (paginate
// sizes a row by the vertical extent of its blocks). A label with no following
// row (an empty tab) keeps its own index.
export function rowIndexFor(
  target: HTMLElement,
  rowContainer: HTMLElement,
): number {
  let section: Element | null = target.closest("section");
  if (!section) return 0;
  if (section.classList.contains("pdf-tab-label-row")) {
    const next = section.nextElementSibling;
    if (
      next?.tagName === "SECTION" &&
      !next.classList.contains("pdf-tab-label-row")
    ) {
      section = next;
    }
  }
  const sections: Element[] = Array.from(
    rowContainer.querySelectorAll("section"),
  );
  const index = sections.indexOf(section);
  return index === -1 ? 0 : index;
}
