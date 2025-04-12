import { getChartComponent } from "@rilldata/web-common/features/canvas/components/charts";
import { CartesianChartComponent } from "@rilldata/web-common/features/canvas/components/charts/cartesian-charts/CartesianChart";
import { KPIGridComponent } from "@rilldata/web-common/features/canvas/components/kpi-grid";
import type {
  ComponentInputParam,
  FilterInputParam,
  FilterInputTypes,
} from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasResponse } from "@rilldata/web-common/features/canvas/selector";
import type {
  RpcStatus,
  V1MetricsViewSpec,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import type { CanvasEntity, ComponentPath } from "../stores/canvas-entity";
import type { BaseCanvasComponent } from "./BaseCanvasComponent";
import { ImageComponent } from "./image";
import { MarkdownCanvasComponent } from "./markdown";
import { PivotCanvasComponent } from "./pivot";
import type {
  CanvasComponentType,
  ComponentCommonProperties,
  ComponentSpec,
} from "./types";

export const commonOptions: Record<
  keyof ComponentCommonProperties,
  ComponentInputParam
> = {
  title: {
    type: "text",
    optional: true,
    showInUI: true,
    label: "Title",
    meta: { placeholder: "Add a title to describe this component" },
  },
  description: {
    type: "text",
    optional: true,
    showInUI: true,
    label: "Description",
    meta: {
      placeholder: "Add additional context for this component",
    },
  },
};

export function getFilterOptions(
  hasComparison = true,
  hasGrain = true,
): Partial<Record<FilterInputTypes, FilterInputParam>> {
  return {
    time_filters: { type: "time_filters", meta: { hasComparison, hasGrain } },
    dimension_filters: {
      type: "dimension_filters",
    },
  };
}

const TABLE_TYPES = ["table", "pivot"] as const;
const CHART_TYPES = [
  "line_chart",
  "bar_chart",
  "stacked_bar",
  "stacked_bar_normalized",
  "area_chart",
] as const;
const NON_CHART_TYPES = [
  "markdown",
  "kpi",
  "kpi_grid",
  "image",
  "table",
  "pivot",
] as const;
const ALL_COMPONENT_TYPES = [...CHART_TYPES, ...NON_CHART_TYPES] as const;

type ChartType = (typeof CHART_TYPES)[number];
type TableType = (typeof TABLE_TYPES)[number];

interface BaseCanvasComponentConstructor<
  T extends ComponentSpec = ComponentSpec,
> {
  new (
    resource: V1Resource,
    parent: CanvasEntity,
    path: ComponentPath,
  ): BaseCanvasComponent<T>;

  newComponentSpec(
    metricsViewName: string,
    metricsViewSpec?: V1MetricsViewSpec,
  ): T;
}

// Component type to class mapping
export const COMPONENT_CLASS_MAP: Record<
  CanvasComponentType,
  BaseCanvasComponentConstructor
> = {
  markdown: MarkdownCanvasComponent,
  kpi_grid: KPIGridComponent,
  image: ImageComponent,
  table: PivotCanvasComponent,
  pivot: PivotCanvasComponent,
  bar_chart: getChartComponent("bar_chart"),
  line_chart: getChartComponent("line_chart"),
  stacked_bar: getChartComponent("stacked_bar"),
  stacked_bar_normalized: getChartComponent("stacked_bar_normalized"),
  area_chart: getChartComponent("area_chart"),
} as const;

// Component display names mapping
const DISPLAY_MAP: Record<CanvasComponentType, string> = {
  kpi_grid: "KPI Grid",
  markdown: "Markdown",
  table: "Table",
  pivot: "Pivot",
  image: "Image",
  bar_chart: "Chart",
  line_chart: "Chart",
  stacked_bar: "Chart",
  stacked_bar_normalized: "Chart",
  area_chart: "Chart",
} as const;

export function createComponent(
  resource: V1Resource,
  parent: CanvasEntity,
  path: ComponentPath,
) {
  const type = resource.component?.spec?.renderer as CanvasComponentType;
  const ComponentClass =
    COMPONENT_CLASS_MAP[type as keyof typeof COMPONENT_CLASS_MAP];
  if (ComponentClass) {
    return new ComponentClass(resource, parent, path);
  }
  return new CartesianChartComponent(resource, parent, path);
}

export function isCanvasComponentType(
  value: string | undefined,
): value is CanvasComponentType {
  if (!value) return false;
  return ALL_COMPONENT_TYPES.includes(value as CanvasComponentType);
}

export function isChartComponentType(
  value: string | undefined,
): value is ChartType {
  if (!value) return false;
  return CHART_TYPES.includes(value as ChartType);
}

export function isTableComponentType(
  value: string | undefined,
): value is TableType {
  if (!value) return false;
  return TABLE_TYPES.includes(value as TableType);
}

export function getHeaderForComponent(
  componentType: CanvasComponentType | null,
) {
  if (!componentType) return "Component";
  return DISPLAY_MAP[componentType] || "Component";
}

export function getComponentMetricsViewFromSpec(
  componentName: string | undefined,
  spec: QueryObserverResult<CanvasResponse, RpcStatus>,
): string | undefined {
  if (!componentName) return undefined;
  const resource = spec.data?.components?.[componentName]?.component;

  if (resource) {
    return resource?.state?.validSpec?.rendererProperties?.metrics_view as
      | string
      | undefined;
  }
  return undefined;
}
