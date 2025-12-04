import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import {
  commonOptions,
  getFilterOptions,
} from "@rilldata/web-common/features/canvas/components/util";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";

import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  V1MetricsViewSpec,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";
import type {
  CanvasComponentType,
  ComponentCommonProperties,
  ComponentComparisonOptions,
  ComponentFilterProperties,
} from "../types";
import KPIGrid from "./KPIGrid.svelte";

export { default as KPIGrid } from "./KPIGrid.svelte";

export const defaultComparisonOptions: ComponentComparisonOptions[] = [
  "delta",
  "percent_change",
];

export interface KPIGridSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  measures: string[];
  // Defaults to "bottom"
  sparkline?: "none" | "bottom" | "right";
  // Defaults to false if undefined;
  hide_time_range?: boolean;
  // Defaults to "delta" and "percent_change"
  comparison?: ComponentComparisonOptions[];
}

export class KPIGridComponent extends BaseCanvasComponent<KPIGridSpec> {
  minSize = { width: 2, height: 2 };
  defaultSize = { width: 6, height: 4 };
  resetParams = ["measures"];
  type: CanvasComponentType = "kpi_grid";
  component = KPIGrid;

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const defaultSpec: KPIGridSpec = {
      metrics_view: "",
      measures: [],
      comparison: defaultComparisonOptions,
    };
    super(resource, parent, path, defaultSpec);
  }

  isValid(spec: KPIGridSpec): boolean {
    return typeof spec.metrics_view === "string";
  }

  getExploreTransformerProperties(): Partial<ExploreState> {
    const spec = get(this.specStore);
    return {
      visibleMeasures: spec.measures,
      activePage: DashboardState_ActivePage.DEFAULT,
      allMeasuresVisible: false,
      leaderboardSortByMeasureName: spec.measures[0],
    };
  }

  inputParams(): InputParams<KPIGridSpec> {
    return {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        measures: {
          type: "multi_fields",
          meta: { allowedTypes: ["measure"] },
          label: "Measures",
        },
        sparkline: { type: "sparkline", optional: true, label: "Sparkline" },
        hide_time_range: {
          type: "boolean",
          optional: true,
          label: "Time range display",
          meta: { invertBoolean: true },
        },
        comparison: { type: "comparison_options", label: "Comparison values" },
        ...commonOptions,
      },
      filter: getFilterOptions(),
    };
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): KPIGridSpec {
    const measures = metricsViewSpec?.measures || [];
    const measureNames = measures.map((m) => m.name as string).slice(0, 4);
    return {
      metrics_view: metricsViewName,
      measures: measureNames,
      comparison: defaultComparisonOptions,
    };
  }
}
