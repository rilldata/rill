import { beforeAll, describe, expect, it, vi } from "vitest";
import {
  describeElement,
  getRecentResizeObserverCallbacks,
  installResizeObserverTracker,
} from "./resize-observer-tracker";

let lastObserverCallback: ResizeObserverCallback | undefined;

function resizeEntries(...targets: Element[]): ResizeObserverEntry[] {
  return targets.map((target) => ({
    target,
  })) as unknown as ResizeObserverEntry[];
}

beforeAll(() => {
  // jsdom has no ResizeObserver: a stub whose callback the test can invoke
  class StubResizeObserver {
    constructor(callback: ResizeObserverCallback) {
      lastObserverCallback = callback;
    }
    observe() {}
    unobserve() {}
    disconnect() {}
  }
  Object.defineProperty(window, "ResizeObserver", {
    configurable: true,
    writable: true,
    value: StubResizeObserver,
  });
  installResizeObserverTracker();
});

describe("installResizeObserverTracker", () => {
  it("records the targets of each callback and still runs it", () => {
    const callback = vi.fn();
    const observer = new ResizeObserver(callback);
    const chart = document.createElement("div");
    chart.id = "chart";
    const legend = document.createElement("ul");
    legend.className = "legend";
    chart.append(legend);
    document.body.append(chart);
    const entries = resizeEntries(chart, legend);

    lastObserverCallback?.(entries, observer);

    expect(callback).toHaveBeenCalledWith(entries, observer);
    const [latest] = getRecentResizeObserverCallbacks();
    expect(latest.targets).toEqual(["div#chart", "ul.legend in #chart"]);
    expect(latest.at).toBeLessThanOrEqual(performance.now());
    chart.remove();
  });

  it("returns the most recent callbacks first and only those within the window", () => {
    const observer = new ResizeObserver(() => {});
    const first = document.createElement("header");
    const second = document.createElement("footer");

    lastObserverCallback?.(resizeEntries(first), observer);
    lastObserverCallback?.(resizeEntries(second), observer);

    const recent = getRecentResizeObserverCallbacks();
    expect(recent[0].targets).toEqual(["footer"]);
    expect(recent[1].targets).toEqual(["header"]);
    expect(getRecentResizeObserverCallbacks(-1)).toEqual([]);
  });

  it("installs only once", () => {
    const Wrapped = window.ResizeObserver;
    installResizeObserverTracker();
    expect(window.ResizeObserver).toBe(Wrapped);
  });
});

describe("describeElement", () => {
  it("keeps descriptions short", () => {
    const element = document.createElement("div");
    element.className = Array.from({ length: 10 }, (_, i) => `c${i}`).join(" ");
    expect(describeElement(element)).toBe("div.c0.c1.c2");

    element.id = "x".repeat(100);
    expect(describeElement(element)).toHaveLength(80);
  });
});
