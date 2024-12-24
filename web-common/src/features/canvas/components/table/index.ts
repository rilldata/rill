import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import { commonOptions } from "@rilldata/web-common/features/canvas/components/util";
import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import { type ComponentCommonProperties } from "../types";

export { default as Table } from "./TableTemplate.svelte";

export interface TableSpec extends ComponentCommonProperties {
  metrics_view: string;
  time_range: string;
  measures: string[];
  comparison_range?: string;
  row_dimensions?: string[];
  col_dimensions?: string[];
}

export class TableCanvasComponent extends BaseCanvasComponent<TableSpec> {
  minSize = { width: 2, height: 2 };
  defaultSize = { width: 12, height: 4 };

  constructor(
    fileArtifact: FileArtifact,
    path: (string | number)[] = [],
    initialSpec: Partial<TableSpec> = {},
  ) {
    const defaultSpec: TableSpec = {
      metrics_view: "",
      measures: [],
      time_range: "",
      comparison_range: "",
      row_dimensions: [],
      col_dimensions: [],
    };
    super(fileArtifact, path, defaultSpec, initialSpec);
  }

  isValid(spec: TableSpec): boolean {
    return (
      typeof spec.time_range === "string" &&
      ((Array.isArray(spec.measures) && spec.measures.length > 0) ||
        (Array.isArray(spec.row_dimensions) &&
          spec.row_dimensions.length > 0) ||
        (Array.isArray(spec.col_dimensions) && spec.col_dimensions.length > 0))
    );
  }

  inputParams(): Record<keyof TableSpec, ComponentInputParam> {
    return {
      metrics_view: { type: "metrics", label: "Metrics view" },
      measures: { type: "multi_measures", label: "Measures" },
      col_dimensions: { type: "multi_dimensions", label: "Column dimensions" },
      row_dimensions: { type: "multi_dimensions", label: "Row dimensions" },
      time_range: { type: "rill_time", label: "Time range" },
      comparison_range: {
        type: "rill_time",
        label: "Comparison range",
        optional: true,
      },
      ...commonOptions,
    };
  }

  newComponentSpec(
    metrics_view: string,
    measure: string,
    dimension: string,
  ): TableSpec {
    return {
      metrics_view,
      measures: [measure],
      row_dimensions: [dimension],
      time_range: "PT24H",
    };
  }
}
