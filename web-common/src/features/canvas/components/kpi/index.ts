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
  show_sparkline?: boolean;
  comparison_range?: string;
}

export class KPIComponent extends BaseCanvasComponent<KPISpec> {
  minSize = { width: 2, height: 2 };
  defaultSize = { width: 4, height: 2 };

  constructor(
    fileArtifact: FileArtifact,
    path: (string | number)[],
    initialSpec: Partial<KPISpec> = {},
  ) {
    const defaultSpec: KPISpec = {
      metrics_view: "",
      measure: "",
      time_range: "",
      // show_sparkline: false,
      title: "",
      description: "",
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
      metrics_view: { type: "metrics", label: "Metric view" },
      measure: { type: "measure", label: "Measure" },
      time_range: { type: "rill_time", label: "Time Range" },
      comparison_range: { type: "rill_time", label: "Comparison Range" },
      show_sparkline: { type: "boolean", required: false, label: "Sparkline" },
      ...commonOptions,
    };
  }
}
