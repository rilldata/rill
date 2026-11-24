import { describe, expect, it } from "vitest";
import { CAPTURE_FAILED_NOTE } from "../rasterize";
import { parseColor, renderPdf } from "./assemble";
import { paginate } from "./layout";

// A 4x2 grayscale JPEG.
const JPEG =
  "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDABsSFBcUERsXFhceHBsgKEIrKCUlKFE6PTBCYFVlZF9VXVtqeJmBanGQc1tdhbWGkJ6jq62rZ4C8ybqmx5moq6T/wAALCAACAAQBAREA/8QAFAABAAAAAAAAAAAAAAAAAAAABP/EABQQAQAAAAAAAAAAAAAAAAAAAAD/2gAIAQEAAD8AK//Z";

describe("renderPdf", () => {
  // A block without an image is one whose capture failed. Handing jsPDF its
  // missing image throws ("Invalid argument passed to jsPDF.scale") and loses
  // the whole document, so the page gets a note where the block sits instead.
  it("draws a note in place of a block whose capture failed", async () => {
    const pagination = paginate(
      [
        { id: "captured", dataUrl: JPEG, rowIndex: 0, ...box(0, 0) },
        { id: "failed", dataUrl: undefined, rowIndex: 1, ...box(0, 220) },
      ],
      { contentWidthPx: 800, format: "a4", orientation: "landscape" },
    );

    const doc = await renderPdf(pagination, {
      title: "Dashboard",
      filename: "dashboard.pdf",
      backgroundColor: "rgb(255, 255, 255)",
      generatedAt: "2026-01-01 00:00 UTC",
      dashboardUrl: "https://example.com/canvas",
    });

    const pdf = doc.output();
    expect(pdf).toContain(`(${CAPTURE_FAILED_NOTE}) Tj`);
    // Only the captured block is embedded as an image.
    expect(pdf.match(/\/Subtype \/Image/g)).toHaveLength(1);
  });
});

function box(xPx: number, yPx: number) {
  return { xPx, yPx, widthPx: 800, heightPx: 200 };
}

describe("parseColor", () => {
  it("parses rgb and rgba colors", () => {
    expect(parseColor("rgb(31, 35, 41)")).toEqual({ r: 31, g: 35, b: 41 });
    expect(parseColor("rgba(31, 35, 41, 1)")).toEqual({
      r: 31,
      g: 35,
      b: 41,
    });
  });

  it("treats transparent backgrounds as white", () => {
    expect(parseColor("rgba(0, 0, 0, 0)")).toEqual({
      r: 255,
      g: 255,
      b: 255,
    });
  });

  it("parses modern CSS colors used by theme tokens", () => {
    expect(parseColor("oklch(24.3% 0 0)")).toEqual({
      r: 32,
      g: 32,
      b: 32,
    });
    expect(parseColor("color(srgb 0.1 0.2 0.3)")).toEqual({
      r: 26,
      g: 51,
      b: 77,
    });
  });
});
