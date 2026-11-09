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

  it("calls the callback with the observer as `this`, like the native ResizeObserver", () => {
    const callback = vi.fn();
    const observer = new ResizeObserver(callback);

    lastObserverCallback?.(resizeEntries(), observer);

    expect(callback.mock.contexts[0]).toBe(observer);
  });

  it("returns the most recent callbacks first and only those within the window", () => {
    const now = vi.spyOn(performance, "now").mockReturnValue(10_000);
    const observer = new ResizeObserver(() => {});
    const first = document.createElement("header");
    const second = document.createElement("footer");

    lastObserverCallback?.(resizeEntries(first), observer);
    now.mockReturnValue(10_050);
    lastObserverCallback?.(resizeEntries(second), observer);

    const recent = getRecentResizeObserverCallbacks();
    expect(recent[0].targets).toEqual(["footer"]);
    expect(recent[1].targets).toEqual(["header"]);

    now.mockReturnValue(10_200);
    expect(getRecentResizeObserverCallbacks()).toEqual([]);
    now.mockRestore();
  });

  it("describes elements as they are when the notice is reported", () => {
    const observer = new ResizeObserver(() => {});
    const panel = document.createElement("section");

    lastObserverCallback?.(resizeEntries(panel), observer);
    panel.id = "filters";

    expect(getRecentResizeObserverCallbacks()[0].targets).toEqual([
      "section#filters",
    ]);
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
