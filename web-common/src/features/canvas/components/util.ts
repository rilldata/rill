import { ImageComponent } from "@rilldata/web-common/features/canvas/components/image";
import { KPIComponent } from "@rilldata/web-common/features/canvas/components/kpi";
import { MarkdownCanvasComponent } from "@rilldata/web-common/features/canvas/components/markdown";
import { TableCanvasComponent } from "@rilldata/web-common/features/canvas/components/table";
import type {
  CanvasComponentType,
  ComponentCommonProperties,
} from "@rilldata/web-common/features/canvas/components/types";
import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";

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
      throw new Error(`Unknown component type: ${type}`);
  }
};
