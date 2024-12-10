import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import {
  commonOptions,
  type ComponentCommonProperties,
} from "../component-types";

export { default as Table } from "./TableTemplate.svelte";

export interface TableSpec extends ComponentCommonProperties {
  metric_view_name: string;
  time_range: string;
  measures: string[];
  comparison_range?: string;
  row_dimensions?: string[];
  col_dimensions?: string[];
}

export class TableComponent extends BaseCanvasComponent<TableSpec> {
  minSize = { width: 2, height: 2 };
  defaultSize = { width: 16, height: 10 };

  constructor(initialSpec: Partial<TableSpec> = {}) {
    const defaultSpec: TableSpec = {
      position: { x: 0, y: 0, width: 4, height: 2 },
      title: "",
      description: "",
      metric_view_name: "",
      measures: [],
      time_range: "",
      comparison_range: "",
      row_dimensions: [],
      col_dimensions: [],
    };
    super(defaultSpec, initialSpec);
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
      ...commonOptions,
      metric_view_name: { type: "metrics_view", label: "Metric view" },
      measures: { type: "multi_measures", label: "Measures" },
      time_range: { type: "rill_time" },
      comparison_range: { type: "rill_time" },
      col_dimensions: { type: "multi_dimensions" },
      row_dimensions: { type: "multi_dimensions" },
    };
  }
}
