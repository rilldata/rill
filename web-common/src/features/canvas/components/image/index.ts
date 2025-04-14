import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import {
  type CanvasComponentType,
  type ComponentAlignment,
  type ComponentCommonProperties,
} from "@rilldata/web-common/features/canvas/components/types";
import { commonOptions } from "@rilldata/web-common/features/canvas/components/util";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";
import Image from "./Image.svelte";

export { default as Image } from "./Image.svelte";

export const defaultImageAlignment: ComponentAlignment = {
  horizontal: "center",
  vertical: "middle",
};

export interface ImageSpec extends ComponentCommonProperties {
  url: string;
  alignment?: ComponentAlignment;
}

export class ImageComponent extends BaseCanvasComponent<ImageSpec> {
  minSize = { width: 1, height: 1 };
  defaultSize = { width: 2, height: 2 };
  resetParams = [];
  type: CanvasComponentType = "image";
  component = Image;

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const defaultSpec: ImageSpec = {
      url: "",
    };
    super(resource, parent, path, defaultSpec);
  }

  isValid(spec: ImageSpec): boolean {
    return typeof spec.url === "string" && spec.url.trim().length > 0;
  }

  inputParams(): InputParams<ImageSpec> {
    return {
      options: {
        url: { type: "text", label: "URL" },
        alignment: {
          type: "alignment",
          label: "Alignment",
          meta: { defaultAlignment: defaultImageAlignment },
        },
        ...commonOptions,
      },
      filter: {},
    };
  }

  static newComponentSpec(): ImageSpec {
    return {
      url: "https://cdn.prod.website-files.com/659ddac460dbacbdc813b204/660b0f85094eb576187342cf_rill_logo_sq_gradient.svg",
    };
  }
}
