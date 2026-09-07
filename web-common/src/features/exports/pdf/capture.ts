import { getFontEmbedCSS, toJpeg } from "html-to-image";
import {
  FILTER_BAR_ID,
  FILTER_BAR_ROW_INDEX,
  type CapturedBlock,
} from "./types";

// Properties that don't reliably serialize from <svg> subtrees during cloning,
// so we pin their computed values inline before capture. Mirrors the approach in
// time-series/ScreenshotContainer.svelte.
const SVG_PROPS = [
  "fill",
  "fill-opacity",
  "stroke",
  "stroke-width",
  "stroke-opacity",
  "stroke-dasharray",
  "stroke-linecap",
  "opacity",
  "font-family",
  "font-size",
  "font-weight",
  "color",
];

export function inlineSvgStyles(root: HTMLElement): () => void {
  const previousStyles: Array<{ el: Element; style: string | null }> = [];
  root.querySelectorAll("svg, svg *").forEach((el) => {
    const cs = getComputedStyle(el);
    const inline = SVG_PROPS.map((p) => `${p}: ${cs.getPropertyValue(p)}`).join(
      "; ",
    );
    previousStyles.push({ el, style: el.getAttribute("style") });
    el.setAttribute("style", `${inline}; ${el.getAttribute("style") ?? ""}`);
  });

  return () => {
    for (const { el, style } of previousStyles) {
      if (style === null) el.removeAttribute("style");
      else el.setAttribute("style", style);
    }
  };
}

const PIXEL_RATIO = 2;
// JPEG keeps PDFs an order of magnitude smaller than lossless PNG while staying
// crisp for dashboard charts/text. JPEG has no alpha, so we supply a background.
const JPEG_QUALITY = 0.85;

// Sized and captured like a real block: clear of the smallest chart block a
// canvas can lay out, at the real pixel ratio. A probe easier to decode than
// the blocks it stands in for wins the race below and reports a browser that
// needs no warm-up, which ships blank charts with no error.
const PROBE_WIDTH_PX = 400;
const PROBE_HEIGHT_PX = 360;

// html-to-image clones a <canvas> into an <img> nested inside the <foreignObject>
// it serializes, and WebKit paints that SVG before the nested image is ready, so
// the first capture of a node containing a canvas comes out blank (Safari 26 on
// macOS and iOS; Chrome and Firefox are unaffected). A second pass over the same
// node is correct. The behaviour is known upstream and still unfixed, so the
// workaround lives here until a html-to-image release carries one.
//
// Rather than pay the extra pass everywhere, or key it off the user agent,
// capture a canvas once and see whether it survives. Memoized at module scope:
// the answer is a property of the browser, so it holds for the page's lifetime.
let canvasWarmupProbe: Promise<boolean> | undefined;

function needsCanvasWarmup(): Promise<boolean> {
  canvasWarmupProbe ??= probeCanvasWarmup();
  return canvasWarmupProbe;
}

async function probeCanvasWarmup(): Promise<boolean> {
  const host = document.createElement("div");
  host.setAttribute("aria-hidden", "true");
  host.style.cssText = "position:fixed;left:-99999px;top:0;pointer-events:none";

  const canvas = document.createElement("canvas");
  canvas.width = PROBE_WIDTH_PX;
  canvas.height = PROBE_HEIGHT_PX;
  canvas.style.display = "block";
  const ctx = canvas.getContext("2d");
  if (!ctx) return true;
  ctx.fillStyle = "#fff";
  ctx.fillRect(0, 0, PROBE_WIDTH_PX, PROBE_HEIGHT_PX);
  // A repeated payload comes back from WebKit's decode cache, which outlives
  // the page.
  ctx.fillStyle = "#000";
  ctx.fillRect(Date.now() % PROBE_WIDTH_PX, 0, 1, 1);

  host.appendChild(canvas);
  document.body.appendChild(host);
  try {
    // White on black: any bright pixel means the canvas reached the raster.
    return await isBlank(
      await toJpeg(host, {
        pixelRatio: PIXEL_RATIO,
        backgroundColor: "#000",
        // The probe asks a question about the browser, so there is no reason to
        // walk the document's stylesheets and inline the app's faces to answer it.
        skipFonts: true,
      }),
    );
  } catch {
    // Assume the warm-up is needed: guessing "no" ships blank charts, guessing
    // "yes" only costs a second pass.
    return true;
  } finally {
    host.remove();
  }
}

async function isBlank(dataUrl: string): Promise<boolean> {
  const img = new Image();
  img.src = dataUrl;
  await img.decode();

  const canvas = document.createElement("canvas");
  canvas.width = img.naturalWidth;
  canvas.height = img.naturalHeight;
  const ctx = canvas.getContext("2d");
  if (!ctx) return true;
  ctx.drawImage(img, 0, 0);

  const { data } = ctx.getImageData(0, 0, canvas.width, canvas.height);
  for (let i = 0; i < data.length; i += 4) {
    if (data[i] > 128) return false;
  }
  return true;
}

export interface RasterizeOptions {
  backgroundColor: string;
  // Web fonts, already resolved to data URIs, shared by every capture. Letting
  // html-to-image re-resolve them per capture pushes the two passes below far
  // enough apart that WebKit drops the decoded canvas between them, and the
  // warm-up stops working.
  fontEmbedCSS: string;
  // Comes from needsCanvasWarmup(). Both fields are required: a caller that
  // forgot the warm-up would ship blank charts on WebKit with nothing to show
  // for it, no error and no failed capture.
  warmUpCanvas: boolean;
}

// Rasterizes a single element to a JPEG data URL. On browsers that need it, a
// node holding a <canvas> is captured twice and the first result discarded; the
// warm-up has to run at the real pixel ratio, as a smaller one does not prime
// the second pass.
export async function rasterizeNode(
  node: HTMLElement,
  { backgroundColor, fontEmbedCSS, warmUpCanvas }: RasterizeOptions,
): Promise<string> {
  const restoreSvgStyles = inlineSvgStyles(node);
  const options = {
    cacheBust: true,
    pixelRatio: PIXEL_RATIO,
    quality: JPEG_QUALITY,
    backgroundColor,
    fontEmbedCSS,
  };
  try {
    if (warmUpCanvas && node.querySelector("canvas")) {
      try {
        await toJpeg(node, options);
      } catch (e) {
        // The warm-up's own result is thrown away, so a failure here is no
        // reason to lose the block: fall through and capture for real.
        console.warn("Canvas warm-up pass failed", e);
      }
    }
    return await toJpeg(node, options);
  } finally {
    restoreSvgStyles();
  }
}

export interface CaptureResult {
  blocks: CapturedBlock[];
  contentWidthPx: number;
  backgroundColor: string;
}

export interface CaptureOptions {
  instanceId: string;
  canvasName: string;
  includeFilters: boolean;
  onProgress?: (ratio: number) => void;
}

// Rasterizes the filter bar (optional) and each canvas block into images
// positioned relative to the canvas content area: every component card plus the
// label band above each exported tab (see captureTargetsIn). Per-block failures
// degrade to a skipped block rather than aborting the whole export.
export async function captureCanvasBlocks(
  opts: CaptureOptions,
): Promise<CaptureResult> {
  // The off-screen export render (see CanvasPdfExportView), mounted only while
  // exporting. Capturing a dedicated tree keeps the live dashboard untouched.
  // Scope the lookup to this canvas store (keyed by instance + canvas name) so a
  // second export view (if another is mounted on the page) can't be captured by
  // mistake.
  const exportView = Array.from(
    document.querySelectorAll<HTMLElement>("#canvas-pdf-export-view"),
  ).find(
    (el) =>
      el.dataset.instanceId === opts.instanceId &&
      el.dataset.canvasName === opts.canvasName,
  );
  const rowContainer = exportView?.querySelector<HTMLElement>(".row-container");

  if (!exportView || !rowContainer) {
    throw new Error(
      "Canvas content is not available to export. Make sure all required filters are set.",
    );
  }

  const contentRect = rowContainer.getBoundingClientRect();
  const contentWidthPx = rowContainer.clientWidth;
  const backgroundColor = getComputedStyle(exportView).backgroundColor;

  const targets = captureTargetsIn(rowContainer);

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
    const rect = target.getBoundingClientRect();
    try {
      const dataUrl = await rasterizeNode(target, {
        backgroundColor,
        fontEmbedCSS,
        warmUpCanvas,
      });
      blocks.push({
        id: target.id,
        dataUrl,
        xPx: rect.left - contentRect.left,
        yPx: rect.top - contentRect.top,
        widthPx: rect.width,
        heightPx: rect.height,
        rowIndex: rowIndexFor(target, rowContainer),
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
