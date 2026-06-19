import { jsPDF } from "jspdf";
import type { PaginationResult, Placement } from "./layout";

export interface AssembleMeta {
  title: string;
  filename: string;
  backgroundColor: string;
  generatedAt: string;
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

    // Always stamp the generation time (UTC) in the footer.
    doc.setFontSize(8);
    doc.setTextColor(120, 120, 120);
    doc.text(
      `Generated ${meta.generatedAt}`,
      result.marginPt,
      result.pageHeightPt - 10,
    );
  }

  doc.save(meta.filename);
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

  if (placement.block.truncatedNote && placement.srcYPx === undefined) {
    const note = placement.block.truncatedNote;
    doc.setFontSize(7);
    doc.setTextColor(120, 120, 120);
    doc.text(note, placement.xPt + 4, placement.yPt + placement.hPt - 4);
  }
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

function parseColor(color: string): RGB | null {
  const match = color.match(/rgba?\(([^)]+)\)/);
  if (!match) return null;
  const parts = match[1].split(",").map((p) => parseFloat(p.trim()));
  if (parts.length < 3) return null;
  // Treat a fully transparent background as white (the canvas paints white).
  if (parts.length >= 4 && parts[3] === 0) return { r: 255, g: 255, b: 255 };
  return { r: parts[0], g: parts[1], b: parts[2] };
}
