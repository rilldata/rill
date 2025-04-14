import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type {
  CommonChartProperties,
  FieldConfig,
} from "@rilldata/web-common/features/canvas/components/charts/types";
import {
  commonOptions,
  getFilterOptions,
} from "@rilldata/web-common/features/canvas/components/util";
import type {
  AllKeys,
  ComponentInputParam,
  InputParams,
} from "@rilldata/web-common/features/canvas/inspector/types";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { get, writable, type Writable } from "svelte/store";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";
import type {
  ComponentCommonProperties,
  ComponentFilterProperties,
} from "../types";
import type { ChartType } from "./";
import Chart from "./Chart.svelte";

// Base interface for all chart configurations
export type BaseChartConfig = ComponentFilterProperties &
  ComponentCommonProperties &
  CommonChartProperties;

export abstract class BaseChart<
  TConfig extends BaseChartConfig,
> extends BaseCanvasComponent<TConfig> {
  minSize = { width: 4, height: 4 };
  defaultSize = { width: 6, height: 4 };
  resetParams = [];
  type: ChartType;
  chartType: Writable<ChartType>;
  component = Chart;

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const baseSpec: BaseChartConfig = {
      metrics_view: "",
      title: "",
      description: "",
    };
    super(resource, parent, path, baseSpec as TConfig);

    this.type = resource.component?.state?.validSpec?.renderer as ChartType;
    this.chartType = writable(this.type);
  }

  isValid(spec: TConfig): boolean {
    return typeof spec.metrics_view === "string";
  }

  inputParams(): InputParams<TConfig> {
    return {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        tooltip: { type: "tooltip", label: "Tooltip", showInUI: false },
        vl_config: { type: "config", showInUI: false },
        ...this.getChartSpecificOptions(),
        ...commonOptions,
      },
      filter: getFilterOptions(false),
    };
  }

  protected abstract getChartSpecificOptions(): Record<
    AllKeys<TConfig>,
    ComponentInputParam
  >;

  protected getDefaultFieldConfig(): Partial<FieldConfig> {
    return {
      showAxisTitle: true,
      zeroBasedOrigin: true,
      showNull: false,
    };
  }

  updateChartType(key: ChartType) {
    if (!this.parent.fileArtifact) return;
    const currentSpec = get(this.specStore);

    const parentPath = this.pathInYAML.slice(0, -1);

    this.chartType.set(key);

    const parseDocumentStore = this.parent.parsedContent;
    const parsedDocument = get(parseDocumentStore);

    const { updateEditorContent } = this.parent.fileArtifact;

    const width = parsedDocument.getIn([...parentPath, "width"]);

    parsedDocument.setIn(parentPath, { [key]: currentSpec, width });

    updateEditorContent(parsedDocument.toString(), false, true);
  }
}
