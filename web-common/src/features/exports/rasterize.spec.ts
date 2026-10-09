// @vitest-environment jsdom
import { toJpeg, toPng } from "html-to-image";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { inlineSvgStyles, rasterizeNode } from "./rasterize";

vi.mock("html-to-image", () => ({
  toJpeg: vi.fn(() => Promise.resolve("data:image/jpeg;base64,")),
  toPng: vi.fn(() => Promise.resolve("data:image/png;base64,")),
  getFontEmbedCSS: vi.fn(() => Promise.resolve("")),
}));

describe("rasterizeNode", () => {
  beforeEach(() => {
    vi.mocked(toJpeg).mockClear();
    vi.mocked(toPng).mockClear();
  });

  function cardWith(inner: string): HTMLElement {
    const card = document.createElement("div");
    card.innerHTML = inner;
    return card;
  }

  // WebKit hands back a blank raster the first time it captures a <canvas>, so
  // affected browsers capture those nodes twice and discard the first result.
  it("captures a canvas-backed node twice when the warm-up is required", async () => {
    await rasterizeNode(cardWith("<canvas></canvas>"), {
      backgroundColor: "#fff",
      fontEmbedCSS: "",
      warmUpCanvas: true,
    });
    expect(vi.mocked(toJpeg)).toHaveBeenCalledTimes(2);
  });

  it("captures once when the browser does not need the warm-up", async () => {
    await rasterizeNode(cardWith("<canvas></canvas>"), {
      backgroundColor: "#fff",
      fontEmbedCSS: "",
      warmUpCanvas: false,
    });
    expect(vi.mocked(toJpeg)).toHaveBeenCalledTimes(1);
  });

  // Only charts render to a canvas; the other blocks must not pay for the pass.
  it("captures a node without a canvas once even on affected browsers", async () => {
    await rasterizeNode(cardWith("<svg></svg>"), {
      backgroundColor: "#fff",
      fontEmbedCSS: "",
      warmUpCanvas: true,
    });
    expect(vi.mocked(toJpeg)).toHaveBeenCalledTimes(1);
  });

  it("captures both passes with identical options", async () => {
    await rasterizeNode(cardWith("<canvas></canvas>"), {
      backgroundColor: "#fff",
      fontEmbedCSS: "",
      warmUpCanvas: true,
    });
    const [first, second] = vi.mocked(toJpeg).mock.calls;
    expect(first[1]).toStrictEqual(second[1]);
  });

  // The warm-up's result is discarded, so a failure in it must not cost the
  // block the real pass would have captured.
  it("captures for real even when the warm-up pass throws", async () => {
    vi.mocked(toJpeg)
      .mockRejectedValueOnce(new Error("warm-up failed"))
      .mockResolvedValueOnce("data:image/jpeg;base64,real");
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {});

    const dataUrl = await rasterizeNode(cardWith("<canvas></canvas>"), {
      backgroundColor: "#fff",
      fontEmbedCSS: "",
      warmUpCanvas: true,
    });

    expect(dataUrl).toBe("data:image/jpeg;base64,real");
    expect(warn).toHaveBeenCalled();
    warn.mockRestore();
  });

  it("hands the caller's font CSS to every pass", async () => {
    const fontEmbedCSS = "@font-face{src:url(data:font/woff2;base64,AA)}";
    await rasterizeNode(cardWith("<canvas></canvas>"), {
      backgroundColor: "#fff",
      fontEmbedCSS,
      warmUpCanvas: true,
    });
    for (const [, options] of vi.mocked(toJpeg).mock.calls) {
      expect(options).toMatchObject({ fontEmbedCSS });
    }
  });

  it("emits a PNG, warm-up pass included, when asked for one", async () => {
    const dataUrl = await rasterizeNode(cardWith("<canvas></canvas>"), {
      backgroundColor: "#fff",
      fontEmbedCSS: "",
      warmUpCanvas: true,
      format: "png",
    });
    expect(dataUrl).toBe("data:image/png;base64,");
    expect(vi.mocked(toPng)).toHaveBeenCalledTimes(2);
    expect(vi.mocked(toJpeg)).not.toHaveBeenCalled();
  });
});

describe("inlineSvgStyles", () => {
  it("restores original SVG style attributes", () => {
    const root = document.createElement("div");
    root.innerHTML = `
      <svg style="color: red">
        <path style="stroke-width: 2" />
        <circle />
      </svg>
    `;

    const svg = root.querySelector("svg")!;
    const path = root.querySelector("path")!;
    const circle = root.querySelector("circle")!;

    const restore = inlineSvgStyles(root);
    expect(svg.getAttribute("style")).not.toBe("color: red");
    expect(path.getAttribute("style")).not.toBe("stroke-width: 2");
    expect(circle.hasAttribute("style")).toBe(true);

    restore();
    expect(svg.getAttribute("style")).toBe("color: red");
    expect(path.getAttribute("style")).toBe("stroke-width: 2");
    expect(circle.hasAttribute("style")).toBe(false);
  });
});
