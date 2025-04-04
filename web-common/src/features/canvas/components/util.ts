import type { QueryObserverResult } from "@tanstack/svelte-query";
import { KPIGridComponent } from "@rilldata/web-common/features/canvas/components/kpi-grid";
import type {
  ComponentInputParam,
  FilterInputParam,
  FilterInputTypes,
} from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasResponse } from "@rilldata/web-common/features/canvas/selector";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import type {
  RpcStatus,
  V1ComponentSpecRendererProperties,
} from "@rilldata/web-common/runtime-client";
import { ChartComponent } from "./charts";
import { ImageComponent } from "./image";
import { MarkdownCanvasComponent } from "./markdown";
import { PivotCanvasComponent } from "./pivot";
import { TableCanvasComponent } from "./table";
import type {
  CanvasComponentType,
  ComponentCommonProperties,
  ComponentFilterProperties,
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

// Component type to class mapping
const COMPONENT_CLASS_MAP = {
  markdown: MarkdownCanvasComponent,
  kpi_grid: KPIGridComponent,
  image: ImageComponent,
  table: TableCanvasComponent,
  pivot: PivotCanvasComponent,
} as const;

// Component display names mapping
const DISPLAY_MAP: Record<CanvasComponentType, string> = {
  kpi: "KPI",
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

export const getComponentObj = (
  fileArtifact: FileArtifact,
  path: (string | number)[],
  type: CanvasComponentType,
  params: Record<string, unknown>,
) => {
  const ComponentClass =
    COMPONENT_CLASS_MAP[type as keyof typeof COMPONENT_CLASS_MAP];
  if (ComponentClass) {
    return new ComponentClass(fileArtifact, path, params);
  }
  return new ChartComponent(fileArtifact, path, params);
};

export type CanvasComponentObj = ReturnType<typeof getComponentObj>;

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

export function getComponentFilterProperties(
  rendererProperties: V1ComponentSpecRendererProperties | undefined,
): ComponentFilterProperties {
  return {
    dimension_filters: rendererProperties?.dimension_filters as
      | string
      | undefined,
    time_filters: rendererProperties?.time_filters as string | undefined,
  };
}

export function getComponentRegistry(): Record<
  CanvasComponentType,
  CanvasComponentObj
> {
  return Object.fromEntries([
    ...Object.entries(COMPONENT_CLASS_MAP).map(([type, Class]) => [
      type,
      new Class(),
    ]),
    ...CHART_TYPES.map((type) => [type, new ChartComponent()]),
  ]) as Record<CanvasComponentType, CanvasComponentObj>;
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
