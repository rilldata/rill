import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import {
  commonOptions,
  type ComponentCommonProperties,
} from "../component-types";

export { default as Markdown } from "./Markdown.svelte";

export interface MarkdownSpec extends ComponentCommonProperties {
  content: string;
}

export class MarkdownCanvasComponent extends BaseCanvasComponent<MarkdownSpec> {
  minSize = { width: 1, height: 1 };
  defaultSize = { width: 6, height: 2 };

  constructor(initialSpec: Partial<MarkdownSpec> = {}) {
    const defaultSpec: MarkdownSpec = {
      position: { x: 0, y: 0, width: 4, height: 2 },
      title: "",
      description: "",
      content: "Your text",
    };
    super(defaultSpec, initialSpec);
  }

  isValid(spec: MarkdownSpec): boolean {
    return typeof spec.content === "string" && spec.content.trim().length > 0;
  }

  inputParams(): Record<keyof MarkdownSpec, ComponentInputParam> {
    return {
      ...commonOptions,
      content: { type: "textArea", required: true },
    };
  }
}
