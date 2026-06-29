import { describe, expect, it } from "vitest";
import { paginate, resolveOrientation } from "./layout";
import type { CapturedBlock } from "./types";

function block(
  partial: Partial<CapturedBlock> & { id: string },
): CapturedBlock {
  return {
    dataUrl: "data:image/png;base64,xxx",
    xPx: 0,
    yPx: 0,
    widthPx: 1000,
    heightPx: 200,
    rowIndex: 0,
    ...partial,
  };
}

// A4 portrait content area: 841.89 - 48 ≈ 793.89pt tall, 595.28 - 48 ≈ 547.28pt wide.
const A4 = { format: "a4" as const, orientation: "portrait" as const };

describe("resolveOrientation", () => {
  it("keeps explicit orientations", () => {
    expect(resolveOrientation("portrait", 2000)).toBe("portrait");
    expect(resolveOrientation("landscape", 100)).toBe("landscape");
  });

  it("auto picks landscape for wide canvases, portrait otherwise", () => {
    expect(resolveOrientation("auto", 1200)).toBe("landscape");
    expect(resolveOrientation("auto", 600)).toBe("portrait");
  });
});

describe("paginate", () => {
  it("places a single short row on one page, scaled to content width", () => {
    const result = paginate(
      [block({ id: "a", widthPx: 1000, heightPx: 200 })],
      {
        ...A4,
        contentWidthPx: 1000,
      },
    );
    expect(result.pageCount).toBe(1);
    expect(result.orientation).toBe("portrait");
    const p = result.placements[0];
    // Scaled to the full content width.
    expect(p.wPt).toBeCloseTo(result.pageWidthPt - 2 * result.marginPt, 1);
    expect(p.page).toBe(0);
    expect(p.xPt).toBeCloseTo(result.marginPt, 1);
    expect(p.yPt).toBeCloseTo(result.marginPt, 1);
  });

  it("keeps two columns of one row on the same page side by side", () => {
    const result = paginate(
      [
        block({ id: "left", xPx: 0, widthPx: 500, heightPx: 300, rowIndex: 0 }),
        block({
          id: "right",
          xPx: 500,
          widthPx: 500,
          heightPx: 300,
          rowIndex: 0,
        }),
      ],
      { ...A4, contentWidthPx: 1000 },
    );
    expect(result.pageCount).toBe(1);
    const [left, right] = result.placements;
    expect(left.page).toBe(0);
    expect(right.page).toBe(0);
    expect(right.xPt).toBeGreaterThan(left.xPt);
  });

  it("paginates multiple rows without splitting a component across pages", () => {
    // Each row fits within one page, so rows move page-by-page without slicing
    // individual components.
    const blocks = [0, 1, 2, 3].map((i) =>
      block({ id: `r${i}`, rowIndex: i, widthPx: 1000, heightPx: 900 }),
    );
    const result = paginate(blocks, { ...A4, contentWidthPx: 1000 });
    // No placement should be a slice.
    expect(result.placements.every((p) => p.srcHeightPx === undefined)).toBe(
      true,
    );
    // Each block appears exactly once.
    expect(result.placements).toHaveLength(4);
    // More than one page used.
    expect(result.pageCount).toBeGreaterThan(1);
    // Every block fits within a single page (height <= content height).
    const contentHeight = result.pageHeightPt - 2 * result.marginPt;
    for (const p of result.placements) {
      expect(p.hPt).toBeLessThanOrEqual(contentHeight + 0.5);
      expect(p.yPt + p.hPt).toBeLessThanOrEqual(
        result.pageHeightPt - result.marginPt + 0.5,
      );
    }
  });

  it("slices a single component taller than a full page across pages", () => {
    // 5000px tall at scale ~0.547 => ~2735pt, content height ~794pt => 4 slices.
    const result = paginate(
      [block({ id: "tall-table", widthPx: 1000, heightPx: 5000 })],
      { ...A4, contentWidthPx: 1000 },
    );
    const slices = result.placements.filter((p) => p.block.id === "tall-table");
    expect(slices.length).toBeGreaterThan(1);
    // Slices have complementary, non-overlapping source crops covering the image.
    const sorted = [...slices].sort(
      (a, b) => (a.srcYPx ?? 0) - (b.srcYPx ?? 0),
    );
    let covered = 0;
    for (const s of sorted) {
      expect(s.srcYPx).toBeCloseTo(covered, 0);
      covered += s.srcHeightPx ?? 0;
    }
    expect(covered).toBeCloseTo(5000, 0);
    // Each slice lives on its own page.
    expect(new Set(slices.map((s) => s.page)).size).toBe(slices.length);
  });

  it("slices every component in a multi-component row taller than a full page", () => {
    const result = paginate(
      [
        block({
          id: "left-table",
          xPx: 0,
          widthPx: 500,
          heightPx: 5000,
          rowIndex: 0,
        }),
        block({
          id: "right-table",
          xPx: 500,
          widthPx: 500,
          heightPx: 5000,
          rowIndex: 0,
        }),
      ],
      { ...A4, contentWidthPx: 1000 },
    );

    const leftSlices = result.placements.filter(
      (p) => p.block.id === "left-table",
    );
    const rightSlices = result.placements.filter(
      (p) => p.block.id === "right-table",
    );
    expect(leftSlices.length).toBeGreaterThan(1);
    expect(leftSlices).toHaveLength(rightSlices.length);
    expect(new Set(leftSlices.map((s) => s.page))).toEqual(
      new Set(rightSlices.map((s) => s.page)),
    );
    for (const placement of result.placements) {
      expect(placement.srcHeightPx).toBeDefined();
      expect(placement.yPt + placement.hPt).toBeLessThanOrEqual(
        result.pageHeightPt - result.marginPt + 0.5,
      );
    }
  });

  it("offsets the first page's content by titleReservePt", () => {
    const withTitle = paginate(
      [block({ id: "a", widthPx: 1000, heightPx: 200 })],
      { ...A4, contentWidthPx: 1000, titleReservePt: 30 },
    );
    // The first block starts one title band below the margin.
    expect(withTitle.placements[0].yPt).toBeCloseTo(withTitle.marginPt + 30, 1);
  });

  it("places the filter bar (rowIndex -1) before content rows", () => {
    const result = paginate(
      [
        block({ id: "comp", rowIndex: 0, heightPx: 200 }),
        block({ id: "__filter_bar__", rowIndex: -1, heightPx: 80 }),
      ],
      { ...A4, contentWidthPx: 1000 },
    );
    const filter = result.placements.find(
      (p) => p.block.id === "__filter_bar__",
    )!;
    const comp = result.placements.find((p) => p.block.id === "comp")!;
    expect(filter.yPt).toBeLessThan(comp.yPt);
  });
});
