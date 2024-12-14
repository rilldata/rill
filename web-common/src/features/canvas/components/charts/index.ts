import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { ChartConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
import { commonOptions } from "@rilldata/web-common/features/canvas/components/util";
import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import { type ComponentCommonProperties } from "../types";

export type ChartSpec = ComponentCommonProperties & ChartConfig;

export class ChartComponent extends BaseCanvasComponent<ChartSpec> {
  minSize = { width: 4, height: 4 };
  defaultSize = { width: 6, height: 6 };

  constructor(
    fileArtifact: FileArtifact,
    path: (string | number)[],
    initialSpec: Partial<ChartSpec> = {},
  ) {
    const defaultSpec: ChartSpec = {
      metrics_view: "",
      title: "",
      description: "",
    };
    super(fileArtifact, path, defaultSpec, initialSpec);
  }

  isValid(spec: ChartSpec): boolean {
    return typeof spec.metrics_view === "string" && Boolean(spec.x || spec.y);
  }

  inputParams(): Record<keyof ChartSpec, ComponentInputParam> {
    return {
      metrics_view: { type: "metrics_view", label: "Metric view" },
      x: { type: "positional", label: "X" },
      y: { type: "positional", label: "Y" },
      color: { type: "mark", label: "Color" },
      tooltip: { type: "tooltip", label: "Tooltip" },
      ...commonOptions,
    };
  }
}
