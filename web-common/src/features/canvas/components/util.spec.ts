import type {
  V1CanvasItem,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import { get, writable } from "svelte/store";
import { describe, expect, it, vi } from "vitest";
import type { CanvasEntity, ComponentPath } from "../stores/canvas-entity";
import {
  COMPONENT_CLASS_MAP,
  createComponent,
  getComponentInstanceType,
} from "./util";

// The components build their own metrics views provider and time filter manager,
// which would subscribe to runtime queries; stub them with the parts the constructors touch.
vi.mock(
  "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts",
  () => ({
    MetricsViewsProvider: vi.fn(function () {
      return { setMetricsViewNames: vi.fn(), cleanup: vi.fn() };
    }),
  }),
);
vi.mock(
  "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts",
  async (importOriginal) => ({
    ...(await importOriginal<object>()),
    TimeFilterManager: vi.fn(function () {
      return {
        storeSync: { on: vi.fn(() => () => {}), setUrlParams: vi.fn() },
        getTimeControlStore: () => writable({}),
        cleanup: vi.fn(),
      };
    }),
  }),
);

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
function viewerCanvasEntity(): CanvasEntity {
  const store = () => writable(undefined);
  const localExpressionFilters = {
    setParamForMetricsView: vi.fn(),
  };
  return {
    allowUnvalidatedSpec: false,
    client: {},
    specStore: store(),
    exactMeasures: store(),
    exportMode: writable(false),
    currentExactMeasures: () => [],
    dashboardProvider: { yamlConfigProvider: {} },
    timeFilterManager: {},
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
