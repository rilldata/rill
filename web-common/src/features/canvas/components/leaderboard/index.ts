import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import {
  commonOptions,
  getFilterOptions,
} from "@rilldata/web-common/features/canvas/components/util";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import type {
  V1MetricsViewSpec,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";
import type {
  CanvasComponentType,
  ComponentCommonProperties,
  ComponentComparisonOptions,
  ComponentFilterProperties,
} from "../types";
import Leaderboard from "./LeaderboardDisplay.svelte";

export { default as Leaderboard } from "./LeaderboardDisplay.svelte";

export const defaultComparisonOptions: ComponentComparisonOptions[] = [
  "delta",
  "percent_change",
];

export interface LeaderboardSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  measures: string[];
  dimensions: string[];
  num_rows: number;
}

export class LeaderboardComponent extends BaseCanvasComponent<LeaderboardSpec> {
  minSize = { width: 3, height: 3 };
  defaultSize = { width: 6, height: 3 };
  resetParams = ["measures", "dimensions"];
  type: CanvasComponentType = "leaderboard";
  component = Leaderboard;

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const defaultSpec: LeaderboardSpec = {
      metrics_view: "",
      measures: [],
      dimensions: [],
      num_rows: 7,
    };
    super(resource, parent, path, defaultSpec);
  }

  isValid(spec: LeaderboardSpec): boolean {
    return typeof spec.metrics_view === "string";
  }

  inputParams(): InputParams<LeaderboardSpec> {
    return {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        measures: {
          type: "multi_fields",
          meta: { allowedTypes: ["measure"] },
          label: "Measures",
        },
        dimensions: {
          type: "multi_fields",
          meta: { allowedTypes: ["dimension"] },
          label: "Dimensions",
        },
        num_rows: { type: "number", label: "Number of rows" },
        ...commonOptions,
      },
      filter: getFilterOptions(),
    };
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): LeaderboardSpec {
    const measures =
      metricsViewSpec?.measures?.slice(0, 1).map((m) => m.name as string) ?? []; // TODO: change to 3

    const dimensions =
      metricsViewSpec?.dimensions
        ?.slice(0, 3)
        .map((d) => d.name || (d.column as string)) ?? [];

    return {
      metrics_view: metricsViewName,
      measures,
      dimensions,
      num_rows: 7,
    };
  }
}
