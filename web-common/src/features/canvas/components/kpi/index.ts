import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
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

export { default as KPI } from "./KPI.svelte";

export interface KPISpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  measure: string;
  sparkline?: "none" | "bottom" | "right"; // defaults to "bottom"
}

export class KPIComponent extends BaseCanvasComponent<KPISpec> {
  minSize = { width: 2, height: 2 };
  defaultSize = { width: 6, height: 4 };

  constructor(
    fileArtifact: FileArtifact | undefined = undefined,
    path: (string | number)[] = [],
    initialSpec: Partial<KPISpec> = {},
  ) {
    const defaultSpec: KPISpec = {
      metrics_view: "",
      measure: "",
    };
    super(fileArtifact, path, defaultSpec, initialSpec);
  }

  isValid(spec: KPISpec): boolean {
    return (
      typeof spec.measure === "string" && typeof spec.metrics_view === "string"
    );
  }

  inputParams(): InputParams<KPISpec> {
    return {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        measure: { type: "measure", label: "Measure" },
        sparkline: { type: "sparkline", optional: true, label: "Sparkline" },
        ...commonOptions,
      },
      filter: getFilterOptions(),
    };
  }

  newComponentSpec(metrics_view: string, measure: string): KPISpec {
    return {
      metrics_view,
      measure,
    };
  }
}
