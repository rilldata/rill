// @vitest-environment jsdom
import { getFontEmbedCSS, toJpeg } from "html-to-image";
import { beforeAll, beforeEach, describe, expect, it, vi } from "vitest";
import {
  captureCanvasBlocks,
  captureTargetsIn,
  inlineSvgStyles,
  rasterizeNode,
  rowIndexFor,
} from "./capture";

vi.mock("html-to-image", () => ({
  toJpeg: vi.fn(() => Promise.resolve("data:image/jpeg;base64,")),
  getFontEmbedCSS: vi.fn(() => Promise.resolve("")),
}));

describe("rasterizeNode", () => {
  beforeEach(() => vi.mocked(toJpeg).mockClear());

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

describe("captureTargetsIn", () => {
  // Mirrors the export render: free rows and tab rows are top-level <section>s
  // holding component cards, and each exported tab is preceded by its label
  // band (a section.pdf-tab-label-row; see CanvasPdfExportTab).
  function renderExportDom(): HTMLElement {
    const container = document.createElement("div");
    container.innerHTML = `
      <section id="canvas-row-0">
        <article id="free-a" class="component-card"></article>
        <article id="free-b" class="component-card"></article>
      </section>
      <section id="pdf-tab-label-overview-first" class="pdf-tab-label-row"></section>
      <section id="canvas-row-overview-first-0">
        <article id="tab-a" class="component-card"></article>
        <article id="tab-b" class="component-card"></article>
      </section>
      <section id="pdf-tab-label-overview-second" class="pdf-tab-label-row"></section>
      <section id="canvas-row-overview-second-0">
        <article id="tab-c" class="component-card"></article>
      </section>
      <section id="canvas-row-2">
        <article id="free-c" class="component-card"></article>
      </section>
    `;
    return container;
  }

  it("captures every card plus each tab's label band, in document order", () => {
    const container = renderExportDom();

    const targets = captureTargetsIn(container);

    expect(targets.map((el) => el.id)).toStrictEqual([
      "free-a",
      "free-b",
      "pdf-tab-label-overview-first",
      "tab-a",
      "tab-b",
      "pdf-tab-label-overview-second",
      "tab-c",
      "free-c",
    ]);
  });

  it("groups each tab's label with the tab's first row, between the free rows", () => {
    const container = renderExportDom();
    const indexOf = (selector: string) =>
      rowIndexFor(container.querySelector<HTMLElement>(selector)!, container);

    expect(indexOf("#free-a")).toBe(0);
    expect(indexOf("#free-b")).toBe(0);
    // A label shares its rowIndex with the row that follows it, so the two
    // paginate as one unit and the label can't be orphaned by a page break.
    expect(indexOf("#pdf-tab-label-overview-first")).toBe(2);
    expect(indexOf("#tab-a")).toBe(2);
    expect(indexOf("#tab-b")).toBe(2);
    expect(indexOf("#pdf-tab-label-overview-second")).toBe(4);
    expect(indexOf("#tab-c")).toBe(4);
    expect(indexOf("#free-c")).toBe(5);
  });

  it("keeps an empty tab's label on its own row", () => {
    const container = document.createElement("div");
    container.innerHTML = `
      <section id="pdf-tab-label-overview-empty" class="pdf-tab-label-row"></section>
      <section id="pdf-tab-label-overview-second" class="pdf-tab-label-row"></section>
      <section id="canvas-row-overview-second-0">
        <article id="tab-a" class="component-card"></article>
      </section>
    `;
    const indexOf = (selector: string) =>
      rowIndexFor(container.querySelector<HTMLElement>(selector)!, container);

    // No row follows the empty tab's label (the next section is another label),
    // so it keeps its own index instead of merging into an unrelated row.
    expect(indexOf("#pdf-tab-label-overview-empty")).toBe(0);
    expect(indexOf("#pdf-tab-label-overview-second")).toBe(2);
    expect(indexOf("#tab-a")).toBe(2);
  });
});

describe("captureCanvasBlocks", () => {
  // needsCanvasWarmup memoizes at module scope, so the probe runs once for the
  // whole file: take a single capture run and assert on what it did.
  let probeFills: number[][];
  let header: HTMLElement;
  let fontNode: HTMLElement;
  let probeNode: HTMLElement;
  let probeOptions: Record<string, unknown>;

  beforeAll(async () => {
    // jsdom cannot rasterize, so hand the probe a context it can paint on and
    // let the blankness check fail into its own catch.
    const fillRect =
      vi.fn<(x: number, y: number, w: number, h: number) => void>();
    vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue({
      fillStyle: "",
      fillRect,
    } as unknown as CanvasRenderingContext2D);

    const view = document.createElement("div");
    view.id = "canvas-pdf-export-view";
    view.dataset.instanceId = "inst";
    view.dataset.canvasName = "canvas";
    header = document.createElement("div");
    header.id = "canvas-pdf-export-header";
    const rows = document.createElement("div");
    rows.className = "row-container";
    view.append(header, rows);
    document.body.appendChild(view);

    vi.mocked(toJpeg).mockClear();
    vi.mocked(getFontEmbedCSS).mockClear();
    await captureCanvasBlocks({
      instanceId: "inst",
      canvasName: "canvas",
      includeFilters: true,
    });

    // Read here, not in the tests: the mocks are cleared between them.
    probeFills = fillRect.mock.calls.map((args) => [...args] as number[]);
    fontNode = vi.mocked(getFontEmbedCSS).mock.calls[0][0];
    [probeNode, probeOptions] = vi.mocked(toJpeg).mock.calls[0] as [
      HTMLElement,
      Record<string, unknown>,
    ];
  });

  // The export header is a sibling of the row container, and getFontEmbedCSS
  // keeps only the @font-face rules used inside the node it is handed, so
  // collecting from the rows alone drops any face only the header uses.
  it("collects the font CSS from a node that covers the header too", () => {
    expect(fontNode.contains(header)).toBe(true);
  });

  // A square small enough to decode before WebKit paints would report a browser
  // that needs no warm-up, and the export would go quietly blank.
  it("probes with a canvas the size of a chart card", () => {
    const canvas = probeNode.querySelector("canvas")!;
    expect(canvas.width).toBeGreaterThanOrEqual(300);
    expect(canvas.height).toBeGreaterThanOrEqual(200);
    expect(probeOptions.pixelRatio).toBe(2);
  });

  // WebKit's decode cache outlives the page, so a probe that serializes the same
  // canvas twice would have its answer handed back from the cache.
  it("signs the probe so it is never the same image twice", () => {
    expect(probeFills.some(([, , w, h]) => w === 1 && h === 1)).toBe(true);
  });

  // The probe carries no text, so resolving the app's web fonts for it is pure
  // latency on the first export in every browser, affected or not.
  it("probes without resolving web fonts", () => {
    expect(probeOptions.skipFonts).toBe(true);
  });
});
