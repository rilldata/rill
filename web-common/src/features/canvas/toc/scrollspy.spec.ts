import { get } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createScrollSpy } from "./scrollspy";
import type { TocEntry } from "./toc";

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

/** Build an entry whose element reports a fixed viewport top. */
function entry(id: string, top: number): TocEntry {
  const el = document.createElement("h2");
  el.id = id;
  el.getBoundingClientRect = () => ({ top }) as DOMRect;
  return { id, text: id, depth: 0, el };
}

describe("createScrollSpy", () => {
  let root: HTMLElement;

  beforeEach(() => {
    vi.stubGlobal("IntersectionObserver", MockIntersectionObserver);
    root = document.createElement("div");
    root.getBoundingClientRect = () => ({ top: 0 }) as DOMRect;
    observe.mockClear();
    disconnect.mockClear();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("starts with no active section when there are no entries", () => {
    const spy = createScrollSpy(root);
    spy.setEntries([]);
    expect(get(spy.activeId)).toBeNull();
  });

  it("observes every entry and, with none visible, keeps the last heading above the band", () => {
    const spy = createScrollSpy(root);
    spy.setEntries([entry("a", -10), entry("b", 50), entry("c", 400)]);

    expect(observe).toHaveBeenCalledTimes(3);
    // 'a' is above the band (top <= root top); 'b' and 'c' are below → active is 'a'.
    expect(get(spy.activeId)).toBe("a");
  });

  it("activates the topmost heading currently in the band", () => {
    const spy = createScrollSpy(root);
    const entries = [entry("a", -100), entry("b", 20), entry("c", 60)];
    spy.setEntries(entries);

    fireIntersections({ b: true, c: true });
    expect(get(spy.activeId)).toBe("b");

    // Scroll further: only 'c' remains in the band.
    fireIntersections({ b: false });
    expect(get(spy.activeId)).toBe("c");
  });

  it("ignores observer updates while locked, then settles on unlock", () => {
    const spy = createScrollSpy(root);
    spy.setEntries([entry("a", -100), entry("b", 20), entry("c", 60)]);

    // Lock to 'c' (as a click-to-scroll would): intermediate intersections must not move it.
    spy.lock("c");
    fireIntersections({ b: true });
    expect(get(spy.activeId)).toBe("c");

    // Once unlocked, it recomputes from the current intersection state.
    spy.unlock();
    expect(get(spy.activeId)).toBe("b");
  });

  it("setActive overrides the computed section immediately", () => {
    const spy = createScrollSpy(root);
    spy.setEntries([entry("a", -10), entry("b", 50)]);
    spy.setActive("b");
    expect(get(spy.activeId)).toBe("b");
  });

  it("disconnects the observer on destroy", () => {
    const spy = createScrollSpy(root);
    spy.setEntries([entry("a", 0)]);
    spy.destroy();
    expect(disconnect).toHaveBeenCalled();
  });
});
