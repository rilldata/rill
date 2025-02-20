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
  defaultSize = { width: 6, height: 4 };

  constructor(
    fileArtifact: FileArtifact | undefined = undefined,
    path: (string | number)[] = [],
    initialSpec: Partial<ChartSpec> = {},
  ) {
    const defaultSpec: ChartSpec = {
      metrics_view: "",
      title: "",
      description: "",
    };
    super(fileArtifact, path, defaultSpec, initialSpec);
  }

  isValid(spec: ChartSpec): boolean {
    return typeof spec.metrics_view === "string" && Boolean(spec.x || spec.y);
  }

  inputParams(): InputParams<ChartSpec> {
    return {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        x: { type: "positional", label: "X-axis" },
        y: { type: "positional", label: "Y-axis" },
        color: { type: "mark", label: "Color", meta: { type: "color" } },
        tooltip: { type: "tooltip", label: "Tooltip", showInUI: false },
        vl_config: { type: "config", showInUI: false },
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
    };
  }
}
