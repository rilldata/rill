import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { ChartConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
import {
  commonOptions,
  getFilterOptions,
} from "@rilldata/web-common/features/canvas/components/util";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import type {
  ComponentCommonProperties,
  ComponentFilterProperties,
} from "../types";

export { default as Chart } from "./Chart.svelte";

export type ChartSpec = ComponentFilterProperties &
  ComponentCommonProperties &
  ChartConfig;

export class ChartComponent extends BaseCanvasComponent<ChartSpec> {
  minSize = { width: 4, height: 4 };
  defaultSize = { width: 12, height: 8 };

  constructor(
    fileArtifact: FileArtifact,
    path: (string | number)[] = [],
    initialSpec: Partial<ChartSpec> = {},
  ) {
    const defaultSpec: ChartSpec = {
      metrics_view: "",
      title: "",
      description: "",
      time_range: "P1D",
    };
    super(fileArtifact, path, defaultSpec, initialSpec);
  }

  isValid(spec: ChartSpec): boolean {
    return (
      typeof spec.metrics_view === "string" &&
      Boolean(spec.x || spec.y) &&
      typeof spec.time_range === "string"
    );
  }

  inputParams(): InputParams<ChartSpec> {
    return {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        x: { type: "positional", label: "X-axis" },
        y: { type: "positional", label: "Y-axis" },
        color: { type: "mark", label: "Color", meta: { type: "color" } },
        tooltip: { type: "tooltip", label: "Tooltip", showInUI: false },
        ...commonOptions,
      },
      filter: getFilterOptions(false),
    };
  }

  newComponentSpec(
    metrics_view: string,
    measure: string,
    dimension: string,
  ): ChartSpec {
    return {
      metrics_view,
      x: {
        type: "nominal",
        field: dimension,
      },
      y: {
        type: "quantitative",
        field: measure,
      },
      time_range: "PT24H",
    };
  }
}
