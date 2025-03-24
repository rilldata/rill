import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import { defaultComparisonOptions } from "@rilldata/web-common/features/canvas/components/kpi";
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

export { default as KPIGrid } from "./KPIGrid.svelte";

export interface KPIGridSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  measures: string[];
  // Defaults to "bottom"
  sparkline?: "none" | "bottom" | "right";
  // Defaults to "delta" and "percent_change"
  comparison?: ComponentComparisonOptions[];
}

export class KPIGridComponent extends BaseCanvasComponent<KPIGridSpec> {
  minSize = { width: 2, height: 2 };
  defaultSize = { width: 6, height: 4 };
  resetParams = ["measures"];

  constructor(
    fileArtifact: FileArtifact | undefined = undefined,
    path: (string | number)[] = [],
    initialSpec: Partial<KPIGridSpec> = {},
  ) {
    const defaultSpec: KPIGridSpec = {
      metrics_view: "",
      measures: [],
      comparison: defaultComparisonOptions,
    };
    super(fileArtifact, path, defaultSpec, initialSpec);
  }

  isValid(spec: KPIGridSpec): boolean {
    return typeof spec.metrics_view === "string";
  }

  inputParams(): InputParams<KPIGridSpec> {
    return {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        measures: {
          type: "multi_fields",
          meta: { allowedTypes: ["measure"] },
          label: "Measures",
        },
        sparkline: { type: "sparkline", optional: true, label: "Sparkline" },
        comparison: { type: "comparison_options", label: "Comparison values" },
        ...commonOptions,
      },
      filter: getFilterOptions(),
    };
  }

  newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): KPIGridSpec {
    const measures = metricsViewSpec?.measures || [];
    const measureNames = measures.map((m) => m.name as string).slice(0, 4);
    return {
      metrics_view: metricsViewName,
      measures: measureNames,
      comparison: defaultComparisonOptions,
    };
  }
}
