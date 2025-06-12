import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import {
  type CanvasComponentType,
  type ComponentAlignment,
  type ComponentCommonProperties,
} from "../types";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";
import Markdown from "./Markdown.svelte";

export { default as Markdown } from "./Markdown.svelte";

export const defaultMarkdownAlignment: ComponentAlignment = {
  vertical: "top",
  horizontal: "left",
};

export interface MarkdownSpec extends ComponentCommonProperties {
  content: string;
  alignment?: ComponentAlignment;
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
      content: "",
      alignment: defaultMarkdownAlignment,
    };
    super(resource, parent, path, defaultSpec);
  }

  isValid(spec: MarkdownSpec): boolean {
    return !!spec;
  }

  inputParams(): InputParams<MarkdownSpec> {
    return {
      options: {
        alignment: {
          type: "alignment",
          label: "Alignment",
          meta: {
            defaultAlignment: defaultMarkdownAlignment,
          },
        },
      },
      filter: {},
    };
  }

  static newComponentSpec(): MarkdownSpec {
    return {
      content: "",
      alignment: defaultMarkdownAlignment,
    };
  }
}
