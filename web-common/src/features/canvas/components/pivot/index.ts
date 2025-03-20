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
  ComponentFilterProperties,
} from "../types";

export interface PivotSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  measures: string[];
  row_dimensions?: string[];
  col_dimensions?: string[];
}

export { default as Pivot } from "./CanvasPivotDisplay.svelte";

export class PivotCanvasComponent extends BaseCanvasComponent<PivotSpec> {
  minSize = { width: 2, height: 2 };
  defaultSize = { width: 4, height: 10 };
  resetParams = ["measures", "row_dimensions", "col_dimensions"];

  constructor(
    fileArtifact: FileArtifact | undefined = undefined,
    path: (string | number)[] = [],
    initialSpec: Partial<PivotSpec> = {},
  ) {
    const defaultSpec: PivotSpec = {
      metrics_view: "",
      measures: [],
      row_dimensions: [],
      col_dimensions: [],
    };
    super(fileArtifact, path, defaultSpec, initialSpec);
  }

  isValid(spec: PivotSpec): boolean {
    return typeof spec.metrics_view === "string";
  }

  inputParams(): InputParams<PivotSpec> {
    return {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        measures: {
          type: "multi_fields",
          meta: { allowedTypes: ["measure"] },
          label: "Measures",
        },
        col_dimensions: {
          type: "multi_fields",
          meta: { allowedTypes: ["time", "dimension"] },
          label: "Column dimensions",
        },
        row_dimensions: {
          type: "multi_fields",
          meta: { allowedTypes: ["time", "dimension"] },
          label: "Row dimensions",
        },
        ...commonOptions,
      },
      filter: getFilterOptions(true, false),
    };
  }

  newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): PivotSpec {
    const firstDimension = metricsViewSpec?.dimensions?.[0]?.name;
    const secondDimension = metricsViewSpec?.dimensions?.[1]?.name;

    return {
      metrics_view: metricsViewName,
      measures:
        metricsViewSpec?.measures?.slice(0, 3).map((m) => m.name as string) ??
        [],
      row_dimensions: firstDimension ? [firstDimension] : [],
      col_dimensions: secondDimension ? [secondDimension] : undefined,
    };
  }
}
