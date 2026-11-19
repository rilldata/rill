import type {
  V1CanvasItem,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import { get, writable, type Writable } from "svelte/store";
import { describe, expect, it, vi } from "vitest";
import type { AppliedControl } from "../controls/controls";
import type { CanvasEntity, ComponentPath } from "../stores/canvas-entity";
import type { KPIGridComponent } from "./kpi-grid";
import type { LeaderboardComponent } from "./leaderboard";
import {
  COMPONENT_CLASS_MAP,
  createComponent,
  getComponentInstanceType,
} from "./util";

function resource(
  renderer: string,
  rendererProperties: Record<string, unknown> = {},
  params: { name: string; type: string }[] = [],
): V1Resource {
  return {
    component: {
      state: {
        validSpec: { renderer, rendererProperties, params },
      },
    },
  } as V1Resource;
}

const externalItem = {
  component: "standalone",
  definedInCanvas: false,
} as V1CanvasItem;

describe("getComponentInstanceType", () => {
  it("keeps non-chart standalone components on their existing renderer", () => {
    expect(getComponentInstanceType(resource("markdown"), externalItem)).toBe(
      "markdown",
    );
  });

  it("keeps legacy Vega custom charts on the multi-query renderer", () => {
    expect(
      getComponentInstanceType(
        resource("custom_chart", { vega_spec: "{}", metrics_sql: ["a", "b"] }),
        externalItem,
      ),
    ).toBe("custom_chart");
  });

  it("routes Flint custom charts through the component reference wrapper", () => {
    expect(
      getComponentInstanceType(
        resource("custom_chart", { spec: { chartType: "Bar Chart" } }),
        externalItem,
      ),
    ).toBe("component_ref");
  });

  it("routes parameterized custom charts through the component reference wrapper", () => {
    expect(
      getComponentInstanceType(
        resource("custom_chart", { vega_spec: "{}" }, [
          { name: "metrics_view", type: "metrics_view" },
        ]),
        externalItem,
      ),
    ).toBe("component_ref");
  });
});

// A canvas entity with only the state the component constructors touch.
// Every store is a real store: `derived` rejects a falsy input when it is built.
// `controls` holds the viewer controls in effect for every component.
function viewerCanvasEntity(
  controls: Writable<AppliedControl[]> = writable([]),
): CanvasEntity {
  const store = () => writable(undefined);
  const localTimeControls = {
    metricsViewName: undefined,
    onUrlChange: vi.fn(),
    interval: store(),
    grainStore: store(),
    urlGrainStore: store(),
    rangeStore: store(),
  };
  const localExpressionFilters = {
    setParamForMetricsView: vi.fn(),
    metricsViewsProvider: { setMetricsViewNames: vi.fn(), cleanup: vi.fn() },
    yamlConfigProvider: { cleanup: vi.fn() },
  };
  return {
    allowUnvalidatedSpec: false,
    client: {},
    specStore: store(),
    exactMeasures: store(),
    exportMode: writable(false),
    currentExactMeasures: () => [],
    controls: { forPath: () => controls },
    metricsView: {
      getMetricsViewFromName: () => writable({ metricsView: undefined }),
    },
    timeManager: {
      createLocalTimeState: () => localTimeControls,
      largestMinTimeGrain: store(),
      hasTimeSeriesMap: writable(new Map()),
      availableGrainsMap: writable(new Map()),
      state: {
        interval: store(),
        grainStore: store(),
        showTimeComparisonStore: store(),
        comparisonRangeStore: store(),
        comparisonIntervalStore: store(),
        timeZoneStore: store(),
        rangeStore: store(),
      },
    },
    expressionFilterManager: {
      createLocalFilterStore: () => localExpressionFilters,
      getExprStoreForMetricsView: () => store(),
    },
  } as unknown as CanvasEntity;
}

describe("createComponent", () => {
  const path: ComponentPath = ["rows", 0, "items", 0, "table"];

  // The runtime clears a component's valid spec when its spec fails validation
  // (a dropped measure, an errored metrics view) but keeps the spec itself, and
  // a viewer is not served that unvalidated spec.
  it.each([
    { renderer: "table" as const, defaults: { metrics_view: "", columns: [] } },
    {
      renderer: "pivot" as const,
      defaults: {
        metrics_view: "",
        measures: [],
        row_dimensions: [],
        col_dimensions: [],
      },
    },
  ])(
    "builds a $renderer whose spec failed validation so it can render its error state",
    ({ renderer, defaults }) => {
      const invalid = {
        meta: { name: { name: "table_1" } },
        component: {
          spec: {
            renderer,
            rendererProperties: { metrics_view: "mv", columns: ["gone"] },
          },
        },
      } as V1Resource;

      const component = createComponent(invalid, viewerCanvasEntity(), path);

      expect(component).toBeInstanceOf(COMPONENT_CLASS_MAP[renderer]);
      expect(component.type).toBe(renderer);
      expect(get(component.specStore)).toEqual(defaults);
      component.destroy();
    },
  );

  it("builds the class of the valid spec's renderer when the YAML names another", () => {
    const changed = {
      meta: { name: { name: "chart_1" } },
      component: {
        state: {
          validSpec: {
            renderer: "bar_chart",
            rendererProperties: { metrics_view: "mv" },
          },
        },
        spec: {
          renderer: "table",
          rendererProperties: { metrics_view: "mv", columns: [] },
        },
      },
    } as V1Resource;

    const component = createComponent(changed, viewerCanvasEntity(), path);

    expect(component).toBeInstanceOf(COMPONENT_CLASS_MAP.bar_chart);
    expect(component.type).toBe("bar_chart");
    component.destroy();
  });
});

describe("components under a viewer control that swaps a measure", () => {
  const path: ComponentPath = ["rows", 0, "items", 0, "leaderboard"];

  function swap(to: string): AppliedControl {
    return {
      control: {
        name: "revenue",
        kind: "measure",
        target: "revenue",
        options: ["revenue", "profit"],
      },
      level: "canvas",
      scope: [],
      urlKey: "c.revenue",
      targetKey: "measure:revenue",
      values: [to],
    };
  }

  it("sorts a leaderboard by the measure viewers picked, and links to Explore with it", () => {
    const controls = writable<AppliedControl[]>([]);
    const component = createComponent(
      resource("leaderboard", {
        metrics_view: "mv",
        measures: ["revenue"],
        dimensions: ["publisher"],
      }),
      viewerCanvasEntity(controls),
      path,
    ) as unknown as LeaderboardComponent;
    expect(get(component.leaderboardState).leaderboardSortByMeasureName).toBe(
      "revenue",
    );

    controls.set([swap("profit")]);

    // The query now asks for profit, so sorting by revenue would fail.
    expect(get(component.leaderboardState).leaderboardSortByMeasureName).toBe(
      "profit",
    );
    expect(component.getExploreTransformerProperties()).toMatchObject({
      visibleMeasures: ["profit"],
      leaderboardSortByMeasureName: "profit",
      leaderboardMeasureNames: ["profit"],
    });
    // The YAML spec, which the editor writes to, is untouched.
    expect(get(component.specStore).measures).toEqual(["revenue"]);
    component.destroy();
  });

  it("links a KPI grid to Explore with the measure viewers picked", () => {
    const controls = writable<AppliedControl[]>([swap("profit")]);
    const component = createComponent(
      resource("kpi_grid", { metrics_view: "mv", measures: ["revenue"] }),
      viewerCanvasEntity(controls),
      ["rows", 0, "items", 0, "kpi_grid"],
    ) as unknown as KPIGridComponent;

    expect(component.getExploreTransformerProperties()).toMatchObject({
      visibleMeasures: ["profit"],
      leaderboardSortByMeasureName: "profit",
    });
    component.destroy();
  });
});
