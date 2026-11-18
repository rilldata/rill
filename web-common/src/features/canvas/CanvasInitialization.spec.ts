import { getResolveCanvasRequest } from "@rilldata/web-common/features/canvas/selector";
import {
  getCanvasStore,
  removeCanvasStore,
} from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import { lastVisitedState } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import CanvasDashboardTest from "@rilldata/web-common/features/canvas/test/CanvasDashboardTest.svelte";
import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import { mockResizeObserverForComponentTesting } from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import {
  type HoistedPageForComponentTests,
  PageMockForComponentTests,
} from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForComponentTests.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type {
  V1CanvasSpec,
  V1ComponentSpec,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import {
  RUNTIME_CONTEXT_KEY,
  RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import { getQueryServiceResolveCanvasQueryKey } from "@rilldata/web-common/runtime-client/v2/gen/query-service";
import type { OnNavigate } from "@sveltejs/kit";
import { render, screen, waitFor } from "@testing-library/svelte";
import {
  afterEach,
  beforeEach,
  describe,
  expect,
  it,
  type MockInstance,
  vi,
} from "vitest";

// The SvelteKit mocks have to be declared in the spec file, since `vi.mock` is hoisted per file.
const hoistedPage: HoistedPageForComponentTests = vi.hoisted(() => ({}) as any);
const onNavigateCallbacks: ((navigation: OnNavigate) => void)[] = vi.hoisted(
  () => [],
);

vi.mock("$app/navigation", () => {
  return {
    goto: (url, opts) => hoistedPage.goto(url, opts),
    afterNavigate: (cb) => hoistedPage.afterNavigate(cb),
    onNavigate: (cb) => onNavigateCallbacks.push(cb),
    beforeNavigate: () => {},
  };
});
// The rune based url stores read `page` from here, and SvelteKit only populates it through its
// client router, which a component test does not boot.
vi.mock("$app/state", async () => {
  return {
    page: (
      await import(
        "@rilldata/web-common/features/dashboards/state-managers/loaders/test/page-state.mock.svelte"
      )
    ).pageStateMock,
  };
});
vi.mock("$app/stores", () => {
  return {
    page: hoistedPage,
    navigating: {
      subscribe: (run: (value: null) => void) => {
        run(null);
        return () => {};
      },
    },
  };
});

const INSTANCE_ID = "test";
const CANVAS_NAME = "revisited_canvas";
const COMPONENT_NAME = `${CANVAS_NAME}--component-0-0`;
const COMPONENT_TITLE = "Revisited image";
const CANVAS_SPEC: V1CanvasSpec = {
  displayName: "Revisited canvas",
  rows: [
    {
      height: 40,
      heightUnit: "px",
      items: [{ component: COMPONENT_NAME, definedInCanvas: true, width: 12 }],
    },
  ],
};
const COMPONENTS: Record<string, V1ComponentSpec> = {
  // Not a markdown component: its templated content query subscribes to the canvas spec
  // for good, which would keep the canvas query cached and hide what this covers.
  [COMPONENT_NAME]: {
    renderer: "image",
    rendererProperties: {
      title: COMPONENT_TITLE,
      url: "https://example.com/logo.png",
    },
    definedInCanvas: true,
  },
};

function setupCanvasPage(canvasName: string) {
  new PageMockForComponentTests(hoistedPage);
  // Components latch their lazy load on an IntersectionObserver, which jsdom does not have.
  vi.stubGlobal(
    "IntersectionObserver",
    class {
      observe() {}
      unobserve() {}
      disconnect() {}
    },
  );

  queryClient.clear();
  removeCanvasStore(canvasName, INSTANCE_ID);
  lastVisitedState.clear();
  onNavigateCallbacks.length = 0;
}

function renderCanvas(canvasName: string, runtimeClient: RuntimeClient) {
  return render(CanvasDashboardTest, {
    props: { canvasName },
    // TODO: we need to make sure every single query uses an explicit queryClient instead of the
    //       global one. Only then we can use a fresh client here.
    context: new Map<string | symbol, unknown>([
      ["$$_queryClient", queryClient],
      [RUNTIME_CONTEXT_KEY, runtimeClient],
    ]),
  });
}

/**
 * Covers coming back to a canvas after leaving it.
 * The canvas store outlives the page that built it, so the second visit must not be served
 * by what the first visit left behind.
 */
describe("CanvasInitialization", () => {
  mockResizeObserverForComponentTesting();
  const mocks = DashboardFetchMocks.useDashboardFetchMocks();

  beforeEach(() => {
    setupCanvasPage(CANVAS_NAME);

    mocks.mockCanvas(CANVAS_NAME, CANVAS_SPEC, {}, COMPONENTS);
  });

  async function assertComponentRendered() {
    await waitFor(() =>
      expect(screen.getByText(COMPONENT_TITLE)).toBeInTheDocument(),
    );
    expect(screen.queryByText(/No valid component/)).not.toBeInTheDocument();
  }

  it("Should render the components when revisited after the canvas query was dropped from the cache", async () => {
    const runtimeClient = new RuntimeClient({
      host: "http://localhost",
      instanceId: INSTANCE_ID,
    });
    const { unmount } = renderCanvas(CANVAS_NAME, runtimeClient);
    await assertComponentRendered();
    unmount();

    // The query cache drops a query once it has had no observers for `gcTime` (5 minutes),
    // so this is the canvas a user returns to after spending that long elsewhere in the project.
    queryClient.removeQueries({
      queryKey: getQueryServiceResolveCanvasQueryKey(
        INSTANCE_ID,
        getResolveCanvasRequest(CANVAS_NAME, false),
      ),
    });
    renderCanvas(CANVAS_NAME, runtimeClient);

    await assertComponentRendered();
  });

  it("Should load the canvas through the current runtime client when revisited", async () => {
    // Leaving the project disposes its runtime client, and coming back creates a new one.
    const firstClient = new RuntimeClient({
      host: "http://localhost",
      instanceId: INSTANCE_ID,
    });
    const { unmount } = renderCanvas(CANVAS_NAME, firstClient);
    await assertComponentRendered();
    unmount();
    firstClient.dispose();

    const secondClient = new RuntimeClient({
      host: "http://localhost",
      instanceId: INSTANCE_ID,
    });
    renderCanvas(CANVAS_NAME, secondClient);

    await assertComponentRendered();
    expect(getCanvasStore(CANVAS_NAME, INSTANCE_ID).canvasEntity.client).toBe(
      secondClient,
    );
  });

  it("Should show the banner with its inline markdown until the user leaves the canvas", async () => {
    mocks.mockCanvas(
      CANVAS_NAME,
      {
        ...CANVAS_SPEC,
        banner: "**Heads up:** see the [status page](https://example.com)",
      },
      {},
      COMPONENTS,
    );
    const navigate = (to: string) => {
      const navigation = {
        from: { url: new URL("http://localhost/canvas/revisited_canvas") },
        to: { url: new URL(to, "http://localhost") },
      } as OnNavigate;
      onNavigateCallbacks.forEach((callback) => callback(navigation));
    };

    renderCanvas(
      CANVAS_NAME,
      new RuntimeClient({ host: "http://localhost", instanceId: INSTANCE_ID }),
    );

    const link = await screen.findByRole("link", { name: "status page" });
    expect(link).toHaveAttribute("href", "https://example.com");
    expect(link).toHaveAttribute("target", "_blank");
    expect(screen.getByText("Heads up:").tagName).toBe("STRONG");

    // The canvas writes its filters and time range to the URL, which is a navigation too.
    navigate("/canvas/revisited_canvas?tr=P7D");
    await assertComponentRendered();
    expect(screen.getByRole("link", { name: "status page" })).toBeVisible();

    navigate("/explore/another_dashboard");
    await waitFor(() =>
      expect(
        screen.queryByRole("link", { name: "status page" }),
      ).not.toBeInTheDocument(),
    );
  });
});

/**
 * Covers custom charts, which wait for the canvas time range before they query when their metrics view has a time dimension.
 * A canvas only resolves a time range when one of its metrics views has a time dimension,
 * so a chart must not wait for it otherwise.
 */
describe("Custom charts and the time range", () => {
  mockResizeObserverForComponentTesting();
  const mocks = DashboardFetchMocks.useDashboardFetchMocks();

  const TIME_RANGE_CANVAS = "time_range_canvas";
  const CHART_SQL = "SELECT publisher, total_records FROM ads_metrics";
  const START = "2026-01-01T00:00:00.000Z";
  const END = "2026-02-01T00:00:00.000Z";
  const METRICS_VIEW: V1MetricsViewSpec = {
    dimensions: [{ name: "publisher" }],
    measures: [{ name: "total_records" }],
  };

  // The bodies of the metrics SQL queries the canvas sent.
  let chartQueries: Record<string, unknown>[] = [];
  let fetchSpy: MockInstance<typeof fetch>;

  beforeEach(() => {
    setupCanvasPage(TIME_RANGE_CANVAS);
    chartQueries = [];

    // Records the chart's queries, and answers the time range lookup, which the shared mocks don't serve.
    const serveMock = globalThis.fetch;
    fetchSpy = vi
      .spyOn(globalThis, "fetch")
      .mockImplementation(async (url, init) => {
        const method = String(url).split("/rill.runtime.v1.")[1];
        if (method === "RuntimeService/QueryResolver") {
          const body =
            typeof init?.body === "string"
              ? init.body
              : new TextDecoder().decode(init?.body as Uint8Array);
          chartQueries.push(JSON.parse(body).resolverProperties);
        }
        if (method === "QueryService/MetricsViewTimeRanges") {
          return Response.json({
            fullTimeRange: { min: START, max: END },
            resolvedTimeRanges: [{ start: START, end: END }],
          });
        }
        return serveMock(url, init);
      });
  });

  afterEach(() => {
    fetchSpy.mockRestore();
  });

  // Renders a canvas whose only component is a custom chart on the given metrics view.
  function renderChartCanvas(metricsView: V1MetricsViewSpec) {
    const componentName = `${TIME_RANGE_CANVAS}--component-0-0`;
    mocks.mockCanvas(
      TIME_RANGE_CANVAS,
      {
        rows: [
          {
            height: 320,
            heightUnit: "px",
            items: [
              { component: componentName, definedInCanvas: true, width: 12 },
            ],
          },
        ],
      },
      { ads_metrics: metricsView },
      {
        [componentName]: {
          renderer: "custom_chart",
          rendererProperties: {
            metrics_sql: [CHART_SQL],
            vega_spec: JSON.stringify({
              data: { name: "query1" },
              mark: "bar",
            }),
          },
          definedInCanvas: true,
        },
      },
    );

    return renderCanvas(
      TIME_RANGE_CANVAS,
      new RuntimeClient({ host: "http://localhost", instanceId: INSTANCE_ID }),
    );
  }

  it("Should query a custom chart without a time range when no metrics view has a time dimension", async () => {
    renderChartCanvas(METRICS_VIEW);

    await waitFor(() =>
      expect(chartQueries).toContainEqual({ sql: CHART_SQL }),
    );
  });

  it("Should query a custom chart with the time range when its metrics view has a time dimension", async () => {
    renderChartCanvas({ ...METRICS_VIEW, timeDimension: "timestamp" });

    await waitFor(() =>
      expect(chartQueries).toContainEqual(
        expect.objectContaining({
          sql: CHART_SQL,
          additional_time_range: expect.objectContaining({
            start: START,
            end: END,
          }),
        }),
      ),
    );
    // It waited for the time range rather than querying without it first.
    expect(chartQueries.every((query) => query.additional_time_range)).toBe(
      true,
    );
  });
});
