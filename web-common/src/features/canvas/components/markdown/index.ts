import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import { type ComponentCommonProperties } from "../types";

export { default as Markdown } from "./Markdown.svelte";

export interface MarkdownSpec extends ComponentCommonProperties {
  content: string;
}

export class MarkdownCanvasComponent extends BaseCanvasComponent<MarkdownSpec> {
  minSize = { width: 1, height: 1 };
  defaultSize = { width: 6, height: 2 };

  constructor(
    fileArtifact: FileArtifact | undefined = undefined,
    path: (string | number)[] = [],
    initialSpec: Partial<MarkdownSpec> = {},
  ) {
    const defaultSpec: MarkdownSpec = {
      title: "",
      description: "",
      content: "Your text",
    };
    super(fileArtifact, path, defaultSpec, initialSpec);
  }

  isValid(spec: MarkdownSpec): boolean {
    return typeof spec.content === "string" && spec.content.trim().length > 0;
  }

  inputParams(): InputParams<MarkdownSpec> {
    return {
      options: {
        content: { type: "textArea" },
      },
      filter: {},
    };
  }

  newComponentSpec(): MarkdownSpec {
    return {
      content: "Markdown Text",
    };
  }
}
