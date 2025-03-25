import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import {
  type ComponentAlignment,
  type ComponentCommonProperties,
} from "@rilldata/web-common/features/canvas/components/types";
import { commonOptions } from "@rilldata/web-common/features/canvas/components/util";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";

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

  constructor(
    fileArtifact: FileArtifact | undefined = undefined,
    path: (string | number)[] = [],
    initialSpec: Partial<ImageSpec> = {},
  ) {
    const defaultSpec: ImageSpec = {
      url: "",
    };
    super(fileArtifact, path, defaultSpec, initialSpec);
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

  newComponentSpec(): ImageSpec {
    return {
      url: "https://cdn.prod.website-files.com/659ddac460dbacbdc813b204/660b0f85094eb576187342cf_rill_logo_sq_gradient.svg",
    };
  }
}
