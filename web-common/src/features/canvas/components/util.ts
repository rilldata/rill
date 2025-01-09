import type {
  ComponentInputParam,
  FilterInputParam,
  FilterInputTypes,
} from "@rilldata/web-common/features/canvas/inspector/types";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import { ChartComponent } from "./charts";
import { ImageComponent } from "./image";
import { KPIComponent } from "./kpi";
import { MarkdownCanvasComponent } from "./markdown";
import { TableCanvasComponent } from "./table";
import type { CanvasComponentType, ComponentCommonProperties } from "./types";

export const commonOptions: Record<
  keyof ComponentCommonProperties,
  ComponentInputParam
> = {
  title: { type: "text", optional: true, showInUI: true, label: "Title" },
  description: {
    type: "text",
    optional: true,
    showInUI: true,
    label: "Description",
  },
  // position: { type: "text", showInUI: false },
};

export function getFilterOptions(
  includeComparisonRange = true,
): Partial<Record<FilterInputTypes, FilterInputParam>> {
  return {
    time_range: { type: "time_range", label: "Time Range" },
    ...(includeComparisonRange
      ? {
          comparison_range: {
            type: "comparison_range",
            label: "Comparison Range",
          },
        }
      : {}),
    dimension_filters: {
      type: "dimension_filters",
      label: "Dimension Filters",
    },
  };
}

export const getComponentObj = (
  fileArtifact: FileArtifact,
  path: (string | number)[],
  type: CanvasComponentType,
  params: Record<string, unknown>,
) => {
  switch (type) {
    case "markdown":
      return new MarkdownCanvasComponent(fileArtifact, path, params);
    case "kpi":
      return new KPIComponent(fileArtifact, path, params);
    case "image":
      return new ImageComponent(fileArtifact, path, params);
    case "table":
      return new TableCanvasComponent(fileArtifact, path, params);
    default:
      return new ChartComponent(fileArtifact, path, params);
  }
};

export type CanvasComponentObj = ReturnType<typeof getComponentObj>;

// TODO: Apply DRY
export function isCanvasComponentType(
  value: string | undefined,
): value is CanvasComponentType {
  if (!value) return false;
  return [
    "line_chart",
    "bar_chart",
    "stacked_bar",
    "markdown",
    "kpi",
    "image",
    "table",
  ].includes(value as CanvasComponentType);
}

export function isChartComponentType(
  value: string | undefined,
): value is CanvasComponentType {
  if (!value) return false;
  return ["line_chart", "bar_chart", "stacked_bar"].includes(
    value as CanvasComponentType,
  );
}

export function getComponentRegistry(
  fileArtifact: FileArtifact,
): Record<CanvasComponentType, CanvasComponentObj> {
  return {
    kpi: new KPIComponent(fileArtifact),
    markdown: new MarkdownCanvasComponent(fileArtifact),
    table: new TableCanvasComponent(fileArtifact),
    image: new ImageComponent(fileArtifact),
    bar_chart: new ChartComponent(fileArtifact),
    line_chart: new ChartComponent(fileArtifact),
    stacked_bar: new ChartComponent(fileArtifact),
  };
}

// TODO: Move to config
const displayMap: Record<CanvasComponentType, string> = {
  kpi: "KPI",
  markdown: "Markdown",
  table: "Table",
  image: "Image",
  bar_chart: "Chart",
  line_chart: "Chart",
  stacked_bar: "Chart",
};

export function getHeaderForComponent(componentType: string | undefined) {
  if (!componentType) return "Component";
  if (!displayMap[componentType as CanvasComponentType]) {
    return "Component";
  }
  return displayMap[componentType as CanvasComponentType];
}
