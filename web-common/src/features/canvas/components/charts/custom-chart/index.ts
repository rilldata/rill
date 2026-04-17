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
import { get } from "svelte/store";
import CanvasCustomChart from "./CanvasCustomChart.svelte";

export interface CustomChart
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view?: string;
  prompt?: string;
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
    this.inferMetricsViewName();
  }

  override update(resource: V1Resource, path: ComponentPath) {
    super.update(resource, path);
    this.inferMetricsViewName();
  }

  /**
   * If metrics_view isn't explicitly set in the spec, extract it from the
   * first metrics_sql query. Metrics SQL always has a `FROM <metrics_view>` clause.
   * This is needed so the canvas filter system can look up filters and
   * dimensions/measures for this component's metrics view.
   */
  private inferMetricsViewName() {
    const spec = get(this.specStore);
    if (spec.metrics_view) {
      this.metricsViewName = spec.metrics_view;
      return;
    }

    const sql = spec.metrics_sql?.[0];
    if (!sql) return;

    const match = sql.match(/\bFROM\s+(\w+)/i);
    if (!match) return;

    this.metricsViewName = match[1];
    this.localFilters = this.parent.filterManager.createLocalFilterStore(
      this.metricsViewName,
    );
    // Update specStore in-memory so the filter system has a metrics_view name
    // to work with. This does NOT persist to YAML — only setSpec() and
    // updateProperty() write to YAML via updateYAML().
    this.specStore.update((s) => ({
      ...s,
      metrics_view: this.metricsViewName,
    }));
  }

  isValid(spec: CustomChart): boolean {
    return (
      Array.isArray(spec.metrics_sql) &&
      spec.metrics_sql.length > 0 &&
      spec.metrics_sql.every((q) => q.trim().length > 0) &&
      spec.vega_spec.trim().length > 0
    );
  }

  hasContent(spec: CustomChart): boolean {
    return (
      (Array.isArray(spec.metrics_sql) &&
        spec.metrics_sql.some((q) => q.trim().length > 0)) ||
      spec.vega_spec.trim().length > 0
    );
  }

  inputParams(): InputParams<CustomChart> {
    return {
      options: {
        prompt: { type: "ai_generate", label: "Edit with AI" },
        metrics_sql: { type: "metrics_sql", label: "Metrics SQL" },
        vega_spec: { type: "vega_spec", label: "Vega Lite Spec" },
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
