import { jsPDF } from "jspdf";
import chroma from "chroma-js";
import type { PaginationResult, Placement } from "./layout";

export interface AssembleMeta {
  title: string;
  filename: string;
  backgroundColor: string;
  generatedAt: string;
  dashboardUrl: string;
}

interface RGB {
  r: number;
  g: number;
  b: number;
}

// Renders the paginated placements into a PDF and triggers a download.
export async function assemblePdf(
  result: PaginationResult,
  meta: AssembleMeta,
): Promise<void> {
  const doc = new jsPDF({
    unit: "pt",
    format: [result.pageWidthPt, result.pageHeightPt],
    orientation: result.orientation,
  });
  doc.setProperties({ title: meta.title });

  const background = parseColor(meta.backgroundColor) ?? {
    r: 255,
    g: 255,
    b: 255,
  };
  const imageCache = new Map<string, HTMLImageElement>();

  for (let page = 0; page < result.pageCount; page++) {
    if (page > 0) {
      doc.addPage(
        [result.pageWidthPt, result.pageHeightPt],
        result.orientation,
      );
    }

    doc.setFillColor(background.r, background.g, background.b);
    doc.rect(0, 0, result.pageWidthPt, result.pageHeightPt, "F");

    for (const placement of result.placements) {
      if (placement.page !== page) continue;
      await drawPlacement(doc, placement, imageCache);
    }

    drawFooter(doc, result, meta);
  }

  doc.save(meta.filename);
}

function drawFooter(
  doc: jsPDF,
  result: PaginationResult,
  meta: AssembleMeta,
): void {
  const yPt = result.pageHeightPt - 10;
  const generatedText = `Generated ${meta.generatedAt}`;
  const linkPrefix = "Open the live dashboard: ";
  const linkText = "View in Rill";

  doc.setFontSize(8);
  doc.setTextColor(120, 120, 120);
  doc.text(generatedText, result.marginPt, yPt);

  const linkXPt =
    result.pageWidthPt -
    result.marginPt -
    doc.getTextWidth(`${linkPrefix}${linkText}`);
  doc.text(linkPrefix, linkXPt, yPt);
  doc.setTextColor(37, 99, 235);
  doc.textWithLink(linkText, linkXPt + doc.getTextWidth(linkPrefix), yPt, {
    url: meta.dashboardUrl,
  });
}

async function drawPlacement(
  doc: jsPDF,
  placement: Placement,
  imageCache: Map<string, HTMLImageElement>,
): Promise<void> {
  let dataUrl = placement.block.dataUrl;

  // Sliced blocks: crop the source image to the requested vertical band.
  if (placement.srcHeightPx !== undefined && placement.srcYPx !== undefined) {
    const img = await loadImage(placement.block.dataUrl, imageCache);
    const ratio = img.naturalHeight / placement.block.heightPx;
    dataUrl = cropImage(
      img,
      placement.srcYPx * ratio,
      placement.srcHeightPx * ratio,
    );
  }

  doc.addImage(
    dataUrl,
    "JPEG",
    placement.xPt,
    placement.yPt,
    placement.wPt,
    placement.hPt,
  );
}

function loadImage(
  dataUrl: string,
  cache: Map<string, HTMLImageElement>,
): Promise<HTMLImageElement> {
  const cached = cache.get(dataUrl);
  if (cached) return Promise.resolve(cached);
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => {
      cache.set(dataUrl, img);
      resolve(img);
    };
    img.onerror = reject;
    img.src = dataUrl;
  });
}

function cropImage(
  img: HTMLImageElement,
  srcY: number,
  srcHeight: number,
): string {
  const canvas = document.createElement("canvas");
  canvas.width = img.naturalWidth;
  canvas.height = Math.round(srcHeight);
  const ctx = canvas.getContext("2d");
  if (!ctx) return img.src;
  ctx.drawImage(
    img,
    0,
    srcY,
    img.naturalWidth,
    srcHeight,
    0,
    0,
    img.naturalWidth,
    srcHeight,
  );
  // Source blocks are already opaque JPEGs, so re-encoding as JPEG is safe and
  // keeps sliced pages as small as the rest of the document.
  return canvas.toDataURL("image/jpeg", 0.85);
}

export function parseColor(color: string): RGB | null {
  const trimmed = color.trim();

  const rgb = parseRgbColor(trimmed);
  if (rgb) return rgb;

  const srgb = parseSrgbColor(trimmed);
  if (srgb) return srgb;

  try {
    const [r, g, b] = chroma(trimmed).rgb();
    return { r, g, b };
  } catch {
    return null;
  }
}

function parseRgbColor(color: string): RGB | null {
  const match = color.match(/rgba?\(([^)]+)\)/);
  if (!match) return null;
  const parts = match[1]
    .split(/[,\s/]+/)
    .filter(Boolean)
    .map((p) => parseFloat(p.trim()));
  if (parts.length < 3) return null;
  // Treat a fully transparent background as white (the canvas paints white).
  if (parts.length >= 4 && parts[3] === 0) return { r: 255, g: 255, b: 255 };
  return { r: parts[0], g: parts[1], b: parts[2] };
}

function parseSrgbColor(color: string): RGB | null {
  const match = color.match(/^color\(\s*srgb\s+([^)]+)\)$/i);
  if (!match) return null;

  const parts = match[1].split(/[,\s/]+/).filter(Boolean);
  if (parts.length < 3) return null;

  const alpha = parts[3] ? parseCssUnit(parts[3]) : 1;
  if (alpha === 0) return { r: 255, g: 255, b: 255 };

  return {
    r: toByte(parseCssUnit(parts[0])),
    g: toByte(parseCssUnit(parts[1])),
    b: toByte(parseCssUnit(parts[2])),
  };
}

function parseCssUnit(value: string): number {
  if (value.endsWith("%")) return parseFloat(value) / 100;
  return parseFloat(value);
}

function toByte(value: number): number {
  if (!Number.isFinite(value)) return 0;
  return Math.round(Math.min(1, Math.max(0, value)) * 255);
}
