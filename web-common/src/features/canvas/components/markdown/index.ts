import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import {
  type CanvasComponentType,
  type ComponentAlignment,
  type ComponentCommonProperties,
  type ComponentFilterProperties,
} from "../types";
import { getFilterOptions } from "../util";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";
import MarkdownCanvas from "./MarkdownCanvas.svelte";

export { default as Markdown } from "./Markdown.svelte";

export const defaultMarkdownAlignment: ComponentAlignment = {
  vertical: "middle",
  horizontal: "left",
};

export interface MarkdownSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  content: string;
  alignment?: ComponentAlignment;
}

export class MarkdownCanvasComponent extends BaseCanvasComponent<MarkdownSpec> {
  minSize = { width: 1, height: 1 };
  defaultSize = { width: 3, height: 2 };
  resetParams = [];
  type: CanvasComponentType = "markdown";
  component = MarkdownCanvas;

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const defaultSpec: MarkdownSpec = {
      title: "",
      description: "",
      content: "Your text",
      alignment: defaultMarkdownAlignment,
    };
    super(resource, parent, path, defaultSpec);
  }

  isValid(spec: MarkdownSpec): boolean {
    return typeof spec.content === "string" && spec.content.trim().length > 0;
  }

  inputParams(): InputParams<MarkdownSpec> {
    return {
      options: {
        content: {
          type: "textArea",
          label: "Markdown",
          description:
            'Write markdown with Go templates. Use {{ metrics_sql "select..." }} to query metrics. Supports Sprig functions for data transformation.',
        },
        alignment: {
          type: "alignment",
          label: "Alignment",
          meta: {
            defaultAlignment: defaultMarkdownAlignment,
          },
        },
      },
      filter: getFilterOptions(false, false),
    };
  }

  static newComponentSpec(): MarkdownSpec {
    const defaultContent = String.raw`# 📊 My Dashboard

Write **markdown** with _formatting_ and access to data from metrics.

Use Go templates with the \`metrics_sql\` function to query your data.`;

    return {
      content: defaultContent,
      alignment: defaultMarkdownAlignment,
    };
  }
}
