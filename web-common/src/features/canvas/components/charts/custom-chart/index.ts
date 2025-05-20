import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import {
  type CanvasComponentType,
  type ComponentCommonProperties,
  type ComponentFilterProperties,
} from "@rilldata/web-common/features/canvas/components/types";
import { commonOptions } from "@rilldata/web-common/features/canvas/components/util";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import type {
  CanvasEntity,
  ComponentPath,
} from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import CanvasCustomChart from "./CanvasCustomChart.svelte";

export interface CustomChart
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_sql: string[];
  vega_spec: string;
}

export class CustomChartComponent extends BaseCanvasComponent<CustomChart> {
  minSize = { width: 4, height: 4 };
  defaultSize = { width: 6, height: 4 };
  resetParams = [];
  type: CanvasComponentType = "custom_chart";
  component = CanvasCustomChart;

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const defaultSpec: CustomChart = {
      metrics_sql: [""],
      vega_spec: "",
    };
    super(resource, parent, path, defaultSpec);
  }

  isValid(spec: CustomChart): boolean {
    return (
      Array.isArray(spec.metrics_sql) &&
      spec.metrics_sql.length > 0 &&
      spec.metrics_sql.every((q) => q.trim().length > 0) &&
      spec.vega_spec.trim().length > 0
    );
  }

  inputParams(): InputParams<CustomChart> {
    return {
      options: {
        metrics_sql: { type: "metrics_sql", label: "Metrics SQL" },
        vega_spec: { type: "textarea", label: "Spec" },
        ...commonOptions,
      },
      filter: {},
    };
  }

  static newComponentSpec(): CustomChart {
    return {
      metrics_sql: [""],
      vega_spec: "",
    };
  }
}
