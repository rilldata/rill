import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import {
  commonOptions,
  type ComponentCommonProperties,
} from "@rilldata/web-common/features/canvas/components/component-types";
import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";

export { default as Image } from "./Image.svelte";

export interface ImageSpec extends ComponentCommonProperties {
  url: string;
}

export class MarkdownCanvasComponent extends BaseCanvasComponent<ImageSpec> {
  minSize = { width: 1, height: 1 };
  defaultSize = { width: 5, height: 5 };

  constructor(initialSpec: Partial<ImageSpec> = {}) {
    const defaultSpec: ImageSpec = {
      position: { x: 0, y: 0, width: 4, height: 2 },
      url: "",
      title: "",
      description: "",
    };
    super(defaultSpec, initialSpec);
  }

  isValid(spec: ImageSpec): boolean {
    return typeof spec.url === "string" && spec.url.trim().length > 0;
  }

  inputParams(): Record<keyof ImageSpec, ComponentInputParam> {
    return {
      ...commonOptions,
      url: { type: "string", label: "URL" },
    };
  }
}
