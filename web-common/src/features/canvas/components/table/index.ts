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
import type {
  CanvasComponentType,
  ComponentCommonProperties,
  ComponentFilterProperties,
} from "../types";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";

export interface TableSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  columns: string[];
}

export { default as Table } from "./CanvasTableDisplay.svelte";

export class TableCanvasComponent extends BaseCanvasComponent<TableSpec> {
  minSize = { width: 2, height: 2 };
  defaultSize = { width: 4, height: 10 };
  resetParams = ["measures", "row_dimensions", "col_dimensions"];
  type: CanvasComponentType = "table";

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const defaultSpec: TableSpec = {
      metrics_view: "",
      columns: [],
    };
    super(resource, parent, path, defaultSpec);
  }

  isValid(spec: TableSpec): boolean {
    return typeof spec.metrics_view === "string";
  }

  inputParams(): InputParams<TableSpec> {
    return {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        columns: {
          type: "multi_fields",
          label: "Columns",
          meta: { allowedTypes: ["time", "dimension", "measure"] },
        },
        ...commonOptions,
      },
      filter: getFilterOptions(true, false),
    };
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): TableSpec {
    const measures =
      metricsViewSpec?.measures?.slice(0, 3).map((m) => m.name as string) ?? [];

    const dimensions =
      metricsViewSpec?.dimensions
        ?.slice(0, 3)
        .map((d) => d.name || (d.column as string)) ?? [];

    return {
      metrics_view: metricsViewName,
      columns: [...dimensions, ...measures],
    };
  }
}
