import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { ChartConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
import {
  commonOptions,
  getFilterOptions,
} from "@rilldata/web-common/features/canvas/components/util";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import { defaultPrimaryColors } from "@rilldata/web-common/features/themes/color-config";
import type {
  V1MetricsViewSpec,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type {
  CanvasComponentType,
  ComponentCommonProperties,
  ComponentFilterProperties,
} from "../types";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";

export { default as Chart } from "./Chart.svelte";

export type ChartSpec = ComponentFilterProperties &
  ComponentCommonProperties &
  ChartConfig;

export class ChartComponent extends BaseCanvasComponent<ChartSpec> {
  minSize = { width: 4, height: 4 };
  defaultSize = { width: 6, height: 4 };
  resetParams = [];
  type: CanvasComponentType;

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const defaultSpec: ChartSpec = {
      metrics_view: "",
      title: "",
      description: "",
    };

    super(resource, parent, path, defaultSpec);

    this.type = resource.component?.state?.validSpec
      ?.renderer as CanvasComponentType;
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

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): ChartSpec {
    // Randomly select a measure and dimension if available
    const measures = metricsViewSpec?.measures || [];

    const timeDimension = metricsViewSpec?.timeDimension;
    const dimensions = metricsViewSpec?.dimensions || [];

    const randomMeasure = measures[Math.floor(Math.random() * measures.length)]
      ?.name as string;

    let randomDimension = "";
    if (!timeDimension) {
      randomDimension = dimensions[
        Math.floor(Math.random() * dimensions.length)
      ]?.name as string;
    }

    const spec: ChartSpec = {
      metrics_view: metricsViewName,
      x: {
        type: timeDimension ? "temporal" : "nominal",
        field: timeDimension || randomDimension,
        sort: "-y",
        limit: 20,
      },
      y: {
        type: "quantitative",
        field: randomMeasure,
        zeroBasedOrigin: true,
      },
      color: `hsl(${defaultPrimaryColors[500].split(" ").join(",")})`,
    };

    return spec;
  }
}
