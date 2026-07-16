import { describe, expect, it } from "vitest";
import { deriveTocEntries, slugify } from "./toc";

describe("slugify", () => {
  it("lowercases and replaces non-alphanumerics with hyphens", () => {
    expect(slugify("Hello, World!")).toBe("hello-world");
  });

  it("trims leading and trailing hyphens", () => {
    expect(slugify("  --Overview--  ")).toBe("overview");
  });

  it("falls back to 'section' for empty/symbol-only text", () => {
    expect(slugify("   ")).toBe("section");
    expect(slugify("!!!")).toBe("section");
  });
});

function buildRoot(html: string): HTMLElement {
  const root = document.createElement("div");
  root.innerHTML = html;
  return root;
}

describe("deriveTocEntries", () => {
  it("returns nothing when there is no content", () => {
    expect(deriveTocEntries(null)).toEqual([]);
    expect(deriveTocEntries(buildRoot("<p>no headings</p>"))).toEqual([]);
  });

  it("only picks up headings inside .canvas-markdown", () => {
    const root = buildRoot(`
      <h2>Outside</h2>
      <div class="canvas-markdown"><h2>Inside</h2></div>
    `);
    const entries = deriveTocEntries(root);
    expect(entries.map((e) => e.text)).toEqual(["Inside"]);
  });

  it("assigns slugified ids and scroll-margin-top to headings without an id", () => {
    const root = buildRoot(
      `<div class="canvas-markdown"><h2>Getting Started</h2></div>`,
    );
    const [entry] = deriveTocEntries(root);
    expect(entry.id).toBe("getting-started");
    expect(entry.el.id).toBe("getting-started");
    expect(entry.el.style.scrollMarginTop).toBe("16px");
  });

  it("preserves an existing id", () => {
    const root = buildRoot(
      `<div class="canvas-markdown"><h2 id="custom">Title</h2></div>`,
    );
    const [entry] = deriveTocEntries(root);
    expect(entry.id).toBe("custom");
  });

  it("dedupes colliding slugs with a numeric suffix", () => {
    const root = buildRoot(`
      <div class="canvas-markdown">
        <h2>Overview</h2>
        <h2>Overview</h2>
        <h2>Overview</h2>
      </div>
    `);
    expect(deriveTocEntries(root).map((e) => e.id)).toEqual([
      "overview",
      "overview-2",
      "overview-3",
    ]);
  });

  it("treats the shallowest heading level as top-level and nests one level deeper", () => {
    const root = buildRoot(`
      <div class="canvas-markdown">
        <h2>Section A</h2>
        <h3>Sub A1</h3>
        <h2>Section B</h2>
      </div>
    `);
    expect(deriveTocEntries(root).map((e) => e.depth)).toEqual([0, 1, 0]);
  });

  it("normalizes nesting when h1 is the shallowest heading used", () => {
    const root = buildRoot(`
      <div class="canvas-markdown">
        <h1>Title</h1>
        <h2>Subsection</h2>
      </div>
    `);
    const entries = deriveTocEntries(root);
    expect(entries.map((e) => e.depth)).toEqual([0, 1]);
  });

  it("gives each heading level its own indent depth (h2 and h3 differ)", () => {
    const root = buildRoot(`
      <div class="canvas-markdown">
        <h1>Title</h1>
        <h2>Section</h2>
        <h3>Subsection</h3>
      </div>
    `);
    expect(deriveTocEntries(root).map((e) => e.depth)).toEqual([0, 1, 2]);
  });

  it("ignores blank headings", () => {
    const root = buildRoot(`
      <div class="canvas-markdown">
        <h2>   </h2>
        <h2>Real</h2>
      </div>
    `);
    expect(deriveTocEntries(root).map((e) => e.text)).toEqual(["Real"]);
  });
});
