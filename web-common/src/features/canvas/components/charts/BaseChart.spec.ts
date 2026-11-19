import { describe, expect, it, vi } from "vitest";
import { get, writable } from "svelte/store";
import { isMap, parseDocument } from "yaml";
import type {
  V1ComponentSpecRendererProperties,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";
import type { ChartType } from "../../../components/charts/types";
import { BaseChart, type BaseChartConfig } from "./BaseChart";

vi.mock("@rilldata/web-common/features/canvas/components/charts", () => ({
  CANVAS_CHART_CONFIG: Object.fromEntries(
    ["bar_chart", "heatmap", "donut_chart"].map((type) => [
      type,
      {
        component: {
          newComponentSpec: (metrics_view: string) => ({ metrics_view }),
          chartInputParams: {},
        },
      },
    ]),
  ),
}));
vi.mock("@rilldata/web-common/features/canvas/components/util", () => ({
  // Construct the next instance from the path supplied by the real updateChartType.
  // Query providers and time/filter managers are unrelated to YAML renderer replacement.
  createComponent: (
    resource: V1Resource,
    parent: CanvasEntity,
    path: ComponentPath,
  ) => chart(resource, parent, path),
  getCommonOptions: vi.fn(),
  getFilterOptions: vi.fn(),
}));

function chart(
  resource: V1Resource,
  parent: CanvasEntity,
  path: ComponentPath,
): BaseChart<BaseChartConfig> {
  const spec = resource.component!.spec!;
  return Object.assign(
    Object.create(BaseChart.prototype) as BaseChart<BaseChartConfig>,
    {
      id: "component",
      parent,
      pathInYAML: path,
      type: spec.renderer as ChartType,
      resource: writable(resource),
      specStore: writable(
        spec.rendererProperties as unknown as BaseChartConfig,
      ),
      chartType: writable(spec.renderer as ChartType),
    },
  );
}
function resource(type: string, spec: BaseChartConfig): V1Resource {
  return {
    component: {
      spec: {
        renderer: type,
        rendererProperties:
          spec as unknown as V1ComponentSpecRendererProperties,
      },
    },
  };
}

const item = { bar_chart: { metrics_view: "ads" }, width: 8, if: "true" };
const cases: {
  name: string;
  yaml: object;
  path: ComponentPath;
  serverResponse?: boolean;
  targets: ChartType[];
}[] = [
  {
    name: "top-level item before reconciliation",
    yaml: { rows: [{ items: [item] }] },
    path: ["rows", 0, "items", 0, "bar_chart"],
    targets: ["heatmap", "donut_chart"],
  },
  {
    name: "item nested in a tab group",
    yaml: { rows: [{ tabs: [{ rows: [{ items: [item] }] }] }] },
    path: ["rows", 0, "tabs", 0, "rows", 0, "items", 0, "bar_chart"],
    targets: ["heatmap", "donut_chart"],
  },
  {
    name: "server response between switches",
    yaml: { rows: [{ items: [item] }] },
    path: ["rows", 0, "items", 0, "bar_chart"],
    serverResponse: true,
    targets: ["heatmap", "donut_chart"],
  },
  {
    name: "three switches before reconciliation",
    yaml: { rows: [{ items: [item] }] },
    path: ["rows", 0, "items", 0, "bar_chart"],
    targets: ["heatmap", "donut_chart", "bar_chart"],
  },
];

describe("optimistic chart type switching", () => {
  it.each(cases)(
    "keeps exactly one renderer for $name",
    ({ yaml, path, targets, serverResponse }) => {
      const parsedContent = writable(parseDocument(JSON.stringify(yaml)));
      const components = new Map<string, BaseChart<BaseChartConfig>>();
      const parent = {
        allowUnvalidatedSpec: true,
        parsedContent,
        fileArtifact: {
          updateEditorContent: (content: string) =>
            parsedContent.set(parseDocument(content)),
        },
        componentsStore: {
          set: (id: string, component: BaseChart<BaseChartConfig>) =>
            components.set(id, component),
        },
        setSelectedComponent: vi.fn(),
        _rows: { refresh: vi.fn() },
        createOptimisticResource: ({
          type,
          spec,
        }: {
          type: string;
          spec: BaseChartConfig;
        }) => resource(type, spec),
      } as unknown as CanvasEntity;
      let current = chart(
        resource("bar_chart", { metrics_view: "ads" }),
        parent,
        path,
      );
      const itemPath = path.slice(0, -1);

      targets.forEach((target) => {
        current.updateChartType(target, undefined);
        current = components.get("component")!;
        const expectedPath = [...itemPath, target];
        expect(current.pathInYAML).toEqual(expectedPath);
        const document = get(parsedContent);
        const actual = document.getIn(itemPath);
        expect(isMap(actual) ? actual.toJSON() : actual).toEqual({
          [target]: { metrics_view: "ads" },
          width: 8,
          if: "true",
        });
        if (serverResponse)
          current.update(
            resource(target, { metrics_view: "ads" }),
            expectedPath as ComponentPath,
          );
      });
    },
  );
});
