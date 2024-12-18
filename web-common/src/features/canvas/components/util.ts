import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
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
  title: { type: "text", required: false, showInUI: true, label: "Title" },
  description: {
    type: "text",
    required: false,
    showInUI: true,
    label: "Description",
  },
  // position: { type: "text", showInUI: false },
};

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
