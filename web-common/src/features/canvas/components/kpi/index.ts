import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import {
  commonOptions,
  type ComponentCommonProperties,
} from "../component-types";

export { default as KPI } from "./KPI.svelte";

export interface KPISpec extends ComponentCommonProperties {
  metric_view_name: string;
  measure: string;
  time_range: string;
  show_sparkline: boolean;
  comparison_range?: string;
}

export class KPIComponent extends BaseCanvasComponent<KPISpec> {
  minSize = { width: 2, height: 2 };
  defaultSize = { width: 4, height: 2 };

  constructor(initialSpec: Partial<KPISpec> = {}) {
    const defaultSpec: KPISpec = {
      position: { x: 0, y: 0, width: 4, height: 2 },
      metric_view_name: "",
      measure: "",
      time_range: "",
      show_sparkline: false,
      title: "",
      description: "",
    };
    super(defaultSpec, initialSpec);
  }

  isValid(spec: KPISpec): boolean {
    return (
      typeof spec.measure === "string" && typeof spec.time_range === "string"
    );
  }

  inputParams(): Record<keyof KPISpec, ComponentInputParam> {
    return {
      ...commonOptions,
      metric_view_name: { type: "metrics_view", label: "Metric view" },
      measure: { type: "measure", label: "Measure" },
      time_range: { type: "rill_time", label: "Time Range" },
      comparison_range: { type: "rill_time", label: "Comparison Range" },
      show_sparkline: { type: "boolean", required: false, label: "Sparkline" },
    };
  }
}
