import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type {
  AllKeys,
  InputParams,
} from "@rilldata/web-common/features/canvas/inspector/types";
import {
  type CanvasComponentType,
  type ComponentAlignment,
  type ComponentCommonProperties,
} from "../types";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";
import { get } from "svelte/store";
import Markdown from "./Markdown.svelte";
import { extractMetricsViewFromContent } from "./util";

export { default as Markdown } from "./Markdown.svelte";

export const defaultMarkdownAlignment: ComponentAlignment = {
  vertical: "middle",
  horizontal: "left",
};

export interface MarkdownSpec extends ComponentCommonProperties {
  content: string;
  metrics_view?: string;
  alignment?: ComponentAlignment;
  apply_formatting?: boolean;
}

export class MarkdownCanvasComponent extends BaseCanvasComponent<MarkdownSpec> {
  minSize = { width: 1, height: 1 };
  defaultSize = { width: 3, height: 2 };
  resetParams = [];
  type: CanvasComponentType = "markdown";
  component = Markdown;

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const defaultSpec: MarkdownSpec = {
      title: "",
      description: "",
      content: "Your text",
      alignment: defaultMarkdownAlignment,
      // Don't set apply_formatting in default - it defaults to false when missing
    };
    super(resource, parent, path, defaultSpec);
  }

  isValid(spec: MarkdownSpec): boolean {
    return typeof spec.content === "string" && spec.content.trim().length > 0;
  }

  updateProperty(
    key: AllKeys<MarkdownSpec>,
    value: MarkdownSpec[AllKeys<MarkdownSpec>],
  ) {
    super.updateProperty(key, value);

    // Auto-detect metrics_view when content changes
    if (key === "content" && typeof value === "string") {
      const currentSpec = get(this.specStore);
      if (!currentSpec.metrics_view) {
        const detected = extractMetricsViewFromContent(value);
        if (detected) {
          super.updateProperty(
            "metrics_view" as AllKeys<MarkdownSpec>,
            detected as MarkdownSpec[AllKeys<MarkdownSpec>],
          );
        }
      }
    }
  }

  inputParams(): InputParams<MarkdownSpec> {
    return {
      options: {
        metrics_view: {
          type: "metrics",
          label: "Metrics view",
          optional: true,
        },
        content: {
          type: "textArea",
          label: "Markdown",
          description: "Write text using the markdown syntax",
        },
        alignment: {
          type: "alignment",
          label: "Alignment",
          meta: {
            defaultAlignment: defaultMarkdownAlignment,
          },
        },
        apply_formatting: {
          type: "boolean",
          optional: true,
          showInUI: true,
          label: "Apply measure value formatting",
          description:
            "Format measure values according to their format settings",
        },
      },
      filter: {},
    };
  }

  static newComponentSpec(): MarkdownSpec {
    const defaultContent = `# H1 Markdown Text
## H2 Markdown text
### H3 Markdown text
#### H4 markdown text
Normal text paragraph with **bold** and _italics_

| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Data 1   | Data 2   | Data 3   |
| Data 4   | Data 5   | Data 6   |

Start writing in **Markdown** and see your text beautifully formatted! 🚀`;

    return {
      content: defaultContent,
      alignment: defaultMarkdownAlignment,
      // Don't set apply_formatting - it defaults to false when missing
    };
  }
}
