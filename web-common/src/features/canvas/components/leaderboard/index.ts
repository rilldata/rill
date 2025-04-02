import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import {
  commonOptions,
  getFilterOptions,
} from "@rilldata/web-common/features/canvas/components/util";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import type {
  ComponentCommonProperties,
  ComponentComparisonOptions,
  ComponentFilterProperties,
} from "../types";

export { default as Leaderboard } from "./CanvasLeaderboardDisplay.svelte";

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

  constructor(
    fileArtifact: FileArtifact | undefined = undefined,
    path: (string | number)[] = [],
    initialSpec: Partial<LeaderboardSpec> = {},
  ) {
    const defaultSpec: LeaderboardSpec = {
      metrics_view: "",
      measures: [],
      dimensions: [],
      num_rows: 7,
    };
    super(fileArtifact, path, defaultSpec, initialSpec);
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

  newComponentSpec(
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
