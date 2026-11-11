import { render, waitFor } from "@testing-library/svelte";
import * as vega from "vega";
import type { TopLevelSpec } from "vega-lite";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { stubCanvasContext } from "../../features/components/charts/cartesian/test-fixtures";

vi.mock("@rilldata/web-common/runtime-client/v2", () => ({
  tryUseRuntimeClient: () => ({
    host: "http://localhost:9009",
    instanceId: "default",
    getJwt: () => undefined,
  }),
}));

import VegaLiteRenderer from "./VegaLiteRenderer.svelte";

const rows = [
  { category: "a", value: 1 },
  { category: "b", value: 2 },
];

const encoding = {
  theta: { field: "value", type: "quantitative" },
  color: { field: "category", type: "nominal" },
} as const;

/** jsdom has no ResizeObserver; the renderer only embeds once it has measured its container. */
class FakeResizeObserver {
  constructor(private readonly callback: ResizeObserverCallback) {}
  observe(target: Element) {
    queueMicrotask(() =>
      this.callback(
        [{ target, contentRect: new DOMRect(0, 0, 400, 300) }] as never,
        this as unknown as ResizeObserver,
      ),
    );
  }
  unobserve() {}
  disconnect() {}
}

function renderChart(spec: TopLevelSpec) {
  const change = vi.spyOn(vega.View.prototype, "change");
  const rendered = render(VegaLiteRenderer, {
    props: {
      spec,
      data: { "metrics-view": rows },
      renderer: "svg",
      // The bound view; it starts unset, as it does behind a `bind:viewVL`.
      viewVL: undefined as unknown as vega.View,
    },
  });
  return { ...rendered, change };
}

/** Resolves once the embedded chart has drawn its arcs. */
async function waitForArcs(container: HTMLElement) {
  await waitFor(() =>
    expect(container.querySelector(".mark-arc path")).not.toBeNull(),
  );
}

describe("VegaLiteRenderer", () => {
  beforeEach(() => {
    stubCanvasContext();
    vi.stubGlobal("ResizeObserver", FakeResizeObserver);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("feeds the view the dataset its spec reads, on embed and on update", async () => {
    const { container, change, rerender } = renderChart({
      data: { name: "metrics-view" },
      mark: "arc",
      encoding,
    });

    await waitFor(() =>
      expect(change).toHaveBeenCalledWith("metrics-view", expect.anything()),
    );
    await waitForArcs(container);
    expect(container.querySelectorAll(".mark-arc path")).toHaveLength(2);

    await rerender({ data: { "metrics-view": rows.slice(0, 1) } });
    await waitFor(() => expect(change).toHaveBeenCalledTimes(2));
    await waitFor(() =>
      expect(container.querySelectorAll(".mark-arc path")).toHaveLength(1),
    );
  });

  it("skips a dataset the compiled spec does not define", async () => {
    // A spec whose only data is inline (as a donut with an "Other" slice used
    // to be) has no `metrics-view` source once compiled, and svelte-vega's
    // update of it would throw "Unrecognized data set" on every data change.
    const { container, change } = renderChart({
      data: { values: rows },
      mark: "arc",
      encoding,
    });

    await waitForArcs(container);
    // The rejected update would follow the first draw; give it time to show up.
    await new Promise((resolve) => setTimeout(resolve, 100));
    expect(change).not.toHaveBeenCalled();
    expect(container.querySelectorAll(".mark-arc path")).toHaveLength(2);
  });
});
