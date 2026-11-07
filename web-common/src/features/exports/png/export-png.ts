import { getFontEmbedCSS } from "html-to-image";
import { needsCanvasWarmup, rasterizeNode } from "../rasterize";

// Rasterizes a single on-screen node to a PNG and triggers a browser download.
// Shares the PDF export's capture path (SVG style inlining, font embedding, the
// WebKit canvas warm-up), so a chart that exports correctly to PDF also exports
// correctly here. The node's own computed background fills the image.
export async function downloadNodeAsPng(
  node: HTMLElement,
  filename: string,
): Promise<void> {
  const warmUpCanvas = await needsCanvasWarmup();
  const fontEmbedCSS = await getFontEmbedCSS(node);
  await document.fonts.ready;
  const dataUrl = await rasterizeNode(node, {
    backgroundColor: getComputedStyle(node).backgroundColor,
    fontEmbedCSS,
    warmUpCanvas,
    format: "png",
  });

  const link = document.createElement("a");
  link.download = filename;
  link.href = dataUrl;
  link.click();
}
