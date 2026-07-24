// @vitest-environment jsdom
import { describe, expect, it } from "vitest";
import { captureTargetsIn, inlineSvgStyles, rowIndexFor } from "./capture";

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
