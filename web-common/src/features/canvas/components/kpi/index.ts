import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import { commonOptions } from "@rilldata/web-common/features/canvas/components/util";
import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import { type ComponentCommonProperties } from "../types";

export { default as KPI } from "./KPI.svelte";

export interface KPISpec extends ComponentCommonProperties {
  metrics_view: string;
  measure: string;
  time_range: string;
  sparkline?: boolean;
  comparison_range?: string;
}

export class KPIComponent extends BaseCanvasComponent<KPISpec> {
  minSize = { width: 2, height: 2 };
  defaultSize = { width: 6, height: 4 };

  constructor(
    fileArtifact: FileArtifact,
    path: (string | number)[] = [],
    initialSpec: Partial<KPISpec> = {},
  ) {
    const defaultSpec: KPISpec = {
      metrics_view: "",
      measure: "",
      time_range: "PT24H",
      sparkline: true,
    };
    super(fileArtifact, path, defaultSpec, initialSpec);
  }

  isValid(spec: KPISpec): boolean {
    return (
      typeof spec.measure === "string" && typeof spec.time_range === "string"
    );
  }

  inputParams(): Record<keyof KPISpec, ComponentInputParam> {
    return {
      metrics_view: { type: "metrics", label: "Metrics view" },
      measure: { type: "measure", label: "Measure" },
      sparkline: { type: "boolean", optional: true, label: "Sparkline" },
      time_range: { type: "rill_time", label: "Time Range" },
      comparison_range: {
        type: "rill_time",
        label: "Comparison Range",
        optional: true,
      },
      ...commonOptions,
    };
  }

  newComponentSpec(metrics_view: string, measure: string): KPISpec {
    return {
      metrics_view,
      measure,
      time_range: "PT24H",
      sparkline: true,
    };
  }
}
