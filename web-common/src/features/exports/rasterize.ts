import { toJpeg, toPng } from "html-to-image";

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

export const PIXEL_RATIO = 2;
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

export function needsCanvasWarmup(): Promise<boolean> {
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

export type RasterFormat = "jpeg" | "png";

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
  // JPEG (the default) for blocks bound for a PDF; PNG for a standalone image.
  format?: RasterFormat;
}

// Rasterizes a single element to a data URL. On browsers that need it, a node
// holding a <canvas> is captured twice and the first result discarded; the
// warm-up has to run at the real pixel ratio, as a smaller one does not prime
// the second pass.
export async function rasterizeNode(
  node: HTMLElement,
  { backgroundColor, fontEmbedCSS, warmUpCanvas, format }: RasterizeOptions,
): Promise<string> {
  const restoreSvgStyles = inlineSvgStyles(node);
  const capture = format === "png" ? toPng : toJpeg;
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
        await capture(node, options);
      } catch (e) {
        // The warm-up's own result is thrown away, so a failure here is no
        // reason to lose the block: fall through and capture for real.
        console.warn("Canvas warm-up pass failed", e);
      }
    }
    return await capture(node, options);
  } finally {
    restoreSvgStyles();
  }
}
