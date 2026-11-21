// @vitest-environment jsdom
import { getFontEmbedCSS, toJpeg } from "html-to-image";
import { beforeAll, describe, expect, it, vi } from "vitest";
import {
  captureCanvasBlocks,
  captureTargetsIn,
  measureCanvasBlocks,
  rowIndexFor,
} from "./capture";

vi.mock("html-to-image", () => ({
  toJpeg: vi.fn(() => Promise.resolve("data:image/jpeg;base64,")),
  toPng: vi.fn(() => Promise.resolve("data:image/png;base64,")),
  getFontEmbedCSS: vi.fn(() => Promise.resolve("")),
}));

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

  // A raster over the browser's canvas limits comes back as an empty data URL.
  // The block keeps its place without an image, so the PDF can say a widget is
  // missing there, and the other blocks are still captured.
  it("keeps a block whose capture produced no image, without one", async () => {
    const view = document.createElement("div");
    view.id = "canvas-pdf-export-view";
    view.dataset.instanceId = "inst";
    view.dataset.canvasName = "failing";
    view.innerHTML = `
      <div class="row-container">
        <section id="canvas-row-0">
          <article id="fine" class="component-card"></article>
          <article id="broken" class="component-card"></article>
        </section>
      </div>
    `;
    document.body.appendChild(view);
    layOut(view.querySelector("#fine")!, { width: 400, height: 200 });
    layOut(view.querySelector("#broken")!, {
      left: 400,
      width: 400,
      height: 200,
    });
    vi.mocked(toJpeg)
      .mockResolvedValueOnce("data:image/jpeg;base64,")
      .mockResolvedValueOnce("data:,");
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {});

    const { blocks } = await captureCanvasBlocks({
      instanceId: "inst",
      canvasName: "failing",
      includeFilters: false,
    });

    expect(
      blocks.map((b) => [b.id, b.dataUrl, b.xPx, b.widthPx]),
    ).toStrictEqual([
      ["fine", "data:image/jpeg;base64,", 0, 400],
      ["broken", undefined, 400, 400],
    ]);
    expect(warn).toHaveBeenCalled();
    warn.mockRestore();
  });
});

describe("measureCanvasBlocks", () => {
  it("reports each target's row, kind and tab in document order", () => {
    const container = document.createElement("div");
    container.innerHTML = `
      <section id="canvas-row-0">
        <article id="free-a" class="component-card"></article>
      </section>
      <section id="pdf-tab-label-g-overview" class="pdf-tab-label-row"></section>
      <section id="canvas-row-g-overview-0" data-tab-group="g" data-tab="overview">
        <article id="tab-a" class="component-card"></article>
      </section>
    `;
    for (const target of captureTargetsIn(container)) layOut(target);
    // The export render sits far off-screen; blocks are positioned within it.
    layOut(container, { left: -99999, top: 40 });
    layOut(container.querySelector("#free-a")!, {
      left: -99989,
      top: 50,
      width: 580,
      height: 300,
    });

    const blocks = measureCanvasBlocks(container);

    expect(
      blocks.map((b) => [b.id, b.rowIndex, b.isTabLabel, b.tabGroup, b.tab]),
    ).toStrictEqual([
      ["free-a", 0, false, undefined, undefined],
      ["pdf-tab-label-g-overview", 2, true, undefined, undefined],
      ["tab-a", 2, false, "g", "overview"],
    ]);
    expect(blocks[0].element).toBe(container.querySelector("#free-a"));
    expect(blocks[0]).toMatchObject({
      xPx: 10,
      yPx: 10,
      widthPx: 580,
      heightPx: 300,
    });
  });

  // Canvas markdown can carry a <style> block, and the export render's rows and
  // cards have the same ids as the dashboard's, so a rule that hides a widget
  // on the dashboard hides it here too. Such a widget has nothing to capture
  // (a zero-size raster is not an image) and must not show up in an export.
  it("leaves out the targets the dashboard doesn't draw", () => {
    const container = document.createElement("div");
    container.innerHTML = `
      <section id="canvas-row-0">
        <article id="shown" class="component-card"></article>
        <article id="collapsed" class="component-card"></article>
      </section>
      <section id="canvas-row-1">
        <article id="in-hidden-row" class="component-card"></article>
      </section>
      <section id="canvas-row-2">
        <article id="no-height" class="component-card"></article>
        <article id="invisible" class="component-card"></article>
      </section>
      <section id="canvas-row-3">
        <article id="in-invisible-row" class="component-card"></article>
      </section>
      <section id="canvas-row-4">
        <article id="in-transparent-row" class="component-card"></article>
      </section>
      <section id="canvas-row-5">
        <article id="in-skipped-row" class="component-card"></article>
      </section>
      <section id="pdf-tab-label-g-hidden" class="pdf-tab-label-row"></section>
      <section id="canvas-row-6">
        <article id="after" class="component-card"></article>
      </section>
      <style>
        #invisible, #canvas-row-3 { visibility: hidden; }
        #canvas-row-4 { opacity: 0; }
        #canvas-row-5 { content-visibility: hidden; }
      </style>
    `;
    // The rules of a <style> block apply once it is in the document.
    document.body.append(container);
    const el = (id: string) => container.querySelector<HTMLElement>(`#${id}`)!;
    layOut(el("shown"));
    layOut(el("after"));
    // `width: 0; height: 0` on the card, and `display: none` on a row: no box.
    layOut(el("collapsed"), { width: 0, height: 0 });
    layOut(el("in-hidden-row"), { width: 0, height: 0 });
    layOut(el("pdf-tab-label-g-hidden"), { width: 0, height: 0 });
    // A card in a row too short for its padding: a width but no height.
    layOut(el("no-height"), { width: 1180, height: 0 });
    // `visibility: hidden`, `opacity: 0` and a row's `content-visibility:
    // hidden` keep the card's box, so only the styles tell that nothing is drawn.
    layOut(el("invisible"));
    layOut(el("in-invisible-row"));
    layOut(el("in-transparent-row"));
    layOut(el("in-skipped-row"));

    try {
      const blocks = measureCanvasBlocks(container);

      // The blocks that remain keep the row they had.
      expect(blocks.map((b) => [b.id, b.rowIndex])).toStrictEqual([
        ["shown", 0],
        ["after", 7],
      ]);
    } finally {
      container.remove();
    }
  });
});

// jsdom has no layout: every element measures 0x0, which is how a browser
// reports an element that isn't drawn. Gives an element the box a browser
// would have laid out for it.
function layOut(
  element: Element,
  box: { left?: number; top?: number; width?: number; height?: number } = {},
): void {
  const { left = 0, top = 0, width = 400, height = 200 } = box;
  element.getBoundingClientRect = () =>
    ({
      left,
      top,
      width,
      height,
      right: left + width,
      bottom: top + height,
      x: left,
      y: top,
    }) as DOMRect;
}
