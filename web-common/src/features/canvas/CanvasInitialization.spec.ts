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
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type {
  V1CanvasSpec,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import {
  RUNTIME_CONTEXT_KEY,
  RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import { getQueryServiceResolveCanvasQueryKey } from "@rilldata/web-common/runtime-client/v2/gen/query-service";
import { render, screen, waitFor } from "@testing-library/svelte";
import { beforeEach, describe, expect, it, vi } from "vitest";

// The SvelteKit mocks have to be declared in the spec file, since `vi.mock` is hoisted per file.
const hoistedPage: HoistedPageForComponentTests = vi.hoisted(() => ({}) as any);

vi.mock("$app/navigation", () => {
  return {
    goto: (url, opts) => hoistedPage.goto(url, opts),
    afterNavigate: (cb) => hoistedPage.afterNavigate(cb),
    onNavigate: () => {},
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

    mocks.mockCanvas(
      CANVAS_NAME,
      CANVAS_SPEC,
      {},
      {
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
      },
    );
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
});

/**
 * Covers components that are filtered by the canvas time range, which they wait for before they query.
 * The time range only resolves when a metrics view on the canvas has a time dimension and data,
 * so a component must not wait for it otherwise, and must say why when it can't be resolved.
 */
describe("Canvas components and the time range", () => {
  mockResizeObserverForComponentTesting();
  const mocks = DashboardFetchMocks.useDashboardFetchMocks();

  const TIME_RANGE_CANVAS = "time_range_canvas";
  const CHART_SQL = "SELECT publisher, total_records FROM ads_metrics";
  const START = "2026-01-01T00:00:00.000Z";
  const END = "2026-02-01T00:00:00.000Z";
  const LOOKUP_FAILURE = "dial tcp 127.0.0.1:9000: connect: connection refused";
  const METRICS_VIEW: V1MetricsViewSpec = {
    dimensions: [{ name: "publisher" }],
    measures: [{ name: "total_records" }],
  };

  beforeEach(() => {
    setupCanvasPage(TIME_RANGE_CANVAS);
    mocks.connectRequests = [];
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

  // The properties of the metrics SQL queries the chart has sent.
  function chartQueries(): Record<string, unknown>[] {
    return mocks.connectRequests
      .filter((request) => request.method === "RuntimeService/QueryResolver")
      .map((request) => request.body.resolverProperties);
  }

  it("Should query a custom chart without a time range when no metrics view has a time dimension", async () => {
    renderChartCanvas(METRICS_VIEW);

    await waitFor(() =>
      expect(chartQueries()).toContainEqual({ sql: CHART_SQL }),
    );
  });

  it("Should wait for the time range before querying a custom chart on a metrics view with a time dimension", async () => {
    mocks.mockConnect("QueryService/MetricsViewTimeRanges", {
      fullTimeRange: { min: START, max: END },
      resolvedTimeRanges: [{ start: START, end: END }],
    });
    renderChartCanvas({ ...METRICS_VIEW, timeDimension: "timestamp" });

    await waitFor(() =>
      expect(chartQueries()).toContainEqual(
        expect.objectContaining({
          sql: CHART_SQL,
          additional_time_range: expect.objectContaining({
            start: START,
            end: END,
          }),
        }),
      ),
    );
    expect(chartQueries().every((query) => query.additional_time_range)).toBe(
      true,
    );
  });

  it("Should show why a component can't load when the time range lookup fails", async () => {
    // The lookup logs the failure it reports.
    const consoleError = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});
    mocks.mockConnect(
      "QueryService/MetricsViewTimeRanges",
      new Error(LOOKUP_FAILURE),
    );
    renderChartCanvas({ ...METRICS_VIEW, timeDimension: "timestamp" });

    await waitFor(() =>
      expect(
        screen.getByText(
          m.canvas_time_range_error({ message: LOOKUP_FAILURE }),
        ),
      ).toBeInTheDocument(),
    );
    consoleError.mockRestore();
  });

  it("Should show that there is no data when the metrics view has none to take a time range from", async () => {
    // What the runtime returns for a metrics view on an empty table.
    mocks.mockConnect("QueryService/MetricsViewTimeRanges", {
      fullTimeRange: {},
      resolvedTimeRanges: [
        { start: "0001-01-01T00:00:00Z", end: "0001-01-02T00:00:00Z" },
      ],
    });
    renderChartCanvas({ ...METRICS_VIEW, timeDimension: "timestamp" });

    await waitFor(() =>
      expect(
        screen.getByText(m.canvas_time_range_no_data()),
      ).toBeInTheDocument(),
    );
  });
});
