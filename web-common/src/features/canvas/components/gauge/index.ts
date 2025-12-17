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
  ComponentFilterProperties,
} from "../types";
import CanvasGauge from "./CanvasGauge.svelte";

export { default as Gauge } from "./Gauge.svelte";

export interface GaugeSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  measure: string;
}

export class GaugeComponent extends BaseCanvasComponent<GaugeSpec> {
  minSize = { width: 3, height: 3 };
  defaultSize = { width: 4, height: 4 };
  resetParams = ["measure"];
  type: CanvasComponentType = "gauge";
  component = CanvasGauge;

  constructor(
    resource: V1Resource,
    parent: CanvasEntity,
    path: ComponentPath,
  ) {
    const defaultSpec: GaugeSpec = {
      metrics_view: "",
      measure: "",
    };
    super(resource, parent, path, defaultSpec);
  }

  isValid(spec: GaugeSpec): boolean {
    return (
      typeof spec.metrics_view === "string" && typeof spec.measure === "string"
    );
  }

  inputParams(): InputParams<GaugeSpec> {
    const inputParams: InputParams<GaugeSpec> = {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        measure: {
          type: "field",
          meta: { allowedTypes: ["measure"] },
          label: "Measure",
        },
        ...commonOptions,
      },
      filter: getFilterOptions(),
    };

    return inputParams;
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): GaugeSpec {
    const measures = metricsViewSpec?.measures || [];
    const measureName = measures[0]?.name as string | undefined;
    return {
      metrics_view: metricsViewName,
      measure: measureName || "",
    };
  }
}

