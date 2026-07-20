import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { CanvasTocController } from "./toc-controller.svelte";

// Controllable IntersectionObserver mock: capture the callback so tests can drive intersections.
let ioCallback: IntersectionObserverCallback;
const observe = vi.fn();
const disconnect = vi.fn();

class MockIntersectionObserver {
  constructor(cb: IntersectionObserverCallback) {
    ioCallback = cb;
  }
  observe = observe;
  unobserve = vi.fn();
  disconnect = disconnect;
  takeRecords = vi.fn();
}

function fireIntersections(states: Record<string, boolean>) {
  const records = Object.entries(states).map(([id, isIntersecting]) => ({
    target: { id } as unknown as Element,
    isIntersecting,
  })) as IntersectionObserverEntry[];
  ioCallback(records, {} as IntersectionObserver);
}

/**
 * Build a scroll container holding markdown headings at the given viewport tops, matching the DOM
 * the controller derives from (`.row-container > .canvas-markdown > h2`).
 */
function buildContainer(headings: { id: string; top: number }[]): HTMLElement {
  const container = document.createElement("div");
  container.getBoundingClientRect = () => ({ top: 0 }) as DOMRect;

  const rowContainer = document.createElement("div");
  rowContainer.className = "row-container";
  const markdown = document.createElement("div");
  markdown.className = "canvas-markdown";
  rowContainer.appendChild(markdown);
  container.appendChild(rowContainer);

  for (const { id, top } of headings) {
    const heading = document.createElement("h2");
    heading.id = id;
    heading.textContent = id;
    heading.getBoundingClientRect = () => ({ top }) as DOMRect;
    markdown.appendChild(heading);
  }

  return container;
}

describe("CanvasTocController", () => {
  beforeEach(() => {
    vi.stubGlobal("IntersectionObserver", MockIntersectionObserver);
    vi.stubGlobal(
      "MutationObserver",
      class {
        observe = vi.fn();
        disconnect = vi.fn();
        takeRecords = vi.fn();
      },
    );
    HTMLElement.prototype.scrollIntoView = vi.fn();
    observe.mockClear();
    disconnect.mockClear();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.useRealTimers();
  });

  it("derives entries from the rendered markdown headings", () => {
    const toc = new CanvasTocController(
      buildContainer([
        { id: "a", top: 0 },
        { id: "b", top: 100 },
      ]),
    );
    expect(toc.entries.map((e) => e.id)).toEqual(["a", "b"]);
    expect(observe).toHaveBeenCalledTimes(2);
  });

  it("keeps the last heading above the band when none are visible", () => {
    const toc = new CanvasTocController(
      buildContainer([
        { id: "a", top: -10 },
        { id: "b", top: 50 },
        { id: "c", top: 400 },
      ]),
    );
    // 'a' is above the band (top <= root top); 'b' and 'c' are below → active is 'a'.
    expect(toc.activeId).toBe("a");
  });

  it("activates the topmost heading currently in the band", () => {
    const toc = new CanvasTocController(
      buildContainer([
        { id: "a", top: -100 },
        { id: "b", top: 20 },
        { id: "c", top: 60 },
      ]),
    );

    fireIntersections({ b: true, c: true });
    expect(toc.activeId).toBe("b");

    fireIntersections({ b: false });
    expect(toc.activeId).toBe("c");
  });

  it("pins the target while a click-scroll is in flight, then settles once scrolling stops", () => {
    vi.useFakeTimers();
    const container = buildContainer([
      { id: "a", top: -100 },
      { id: "b", top: 20 },
      { id: "c", top: 60 },
    ]);
    const toc = new CanvasTocController(container);

    fireIntersections({ b: true });
    expect(toc.activeId).toBe("b");

    // Click 'c': it's pinned, and intermediate intersections must not move it.
    const event = {
      preventDefault: vi.fn(),
      currentTarget: document.createElement("a"),
    } as unknown as MouseEvent;
    toc.jumpTo(event, toc.entries[2]);
    expect(toc.activeId).toBe("c");
    fireIntersections({ b: true });
    expect(toc.activeId).toBe("c");

    // Scrolling stops → unlock and settle on the section actually in view.
    container.dispatchEvent(new Event("scroll"));
    vi.advanceTimersByTime(100);
    expect(toc.activeId).toBe("b");
  });

  it("disconnects observers on destroy", () => {
    const toc = new CanvasTocController(buildContainer([{ id: "a", top: 0 }]));
    toc.destroy();
    expect(disconnect).toHaveBeenCalled();
  });
});
