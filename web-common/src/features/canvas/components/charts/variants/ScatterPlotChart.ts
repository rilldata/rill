import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  ScatterPlotChartProvider,
  type ScatterPlotChartSpec as ScatterPlotChartSpecBase,
} from "@rilldata/web-common/features/components/charts/scatter/ScatterPlotChartProvider";
import {
  type ChartDataQuery,
  type ChartFieldsMap,
} from "@rilldata/web-common/features/components/charts/types";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  MetricsViewSpecDimensionType,
  type V1MetricsViewSpec,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { get, type Readable } from "svelte/store";
import type {
  CanvasEntity,
  ComponentPath,
} from "../../../stores/canvas-entity";
import { BaseChart, type BaseChartConfig } from "../BaseChart";

export type ScatterPlotCanvasChartSpec = BaseChartConfig &
  ScatterPlotChartSpecBase;

const DEFAULT_SPLIT_LIMIT = 10;

export class ScatterPlotChartComponent extends BaseChart<ScatterPlotCanvasChartSpec> {
  private provider: ScatterPlotChartProvider;

  static chartInputParams: Record<string, ComponentInputParam> = {
    x: {
      type: "positional",
      label: "X-axis",
      meta: {
        chartFieldInput: {
          type: "measure",
          axisTitleSelector: true,
          axisRangeSelector: true,
        },
      },
    },
    y: {
      type: "positional",
      label: "Y-axis",
      meta: {
        chartFieldInput: {
          type: "measure",
          axisTitleSelector: true,
          axisRangeSelector: true,
        },
      },
    },
    dimension: {
      type: "positional",
      label: "Dimension",
      meta: {
        chartFieldInput: {
          type: "dimension",
          nullSelector: true,
        },
      },
    },
    size: {
      type: "positional",
      label: "Size",
      meta: {
        chartFieldInput: {
          type: "measure",
        },
      },
    },
    color: {
      type: "mark",
      label: "Color",
      showInUI: true,
      meta: {
        type: "color",
        chartFieldInput: {
          type: "dimension",
          defaultLegendOrientation: "top",
          limitSelector: { defaultLimit: DEFAULT_SPLIT_LIMIT },
          colorMappingSelector: { enable: true },
          nullSelector: true,
        },
      },
    },
  };

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    super(resource, parent, path);

    this.provider = new ScatterPlotChartProvider(this.specStore);

    this.provider.combinedWhere.subscribe((where) => {
      this.componentFilters = where;
    });
  }

  getChartSpecificOptions(): Record<string, ComponentInputParam> {
    const inputParams = { ...ScatterPlotChartComponent.chartInputParams };

    inputParams.color.meta!.chartFieldInput = {
      type: "dimension",
      defaultLegendOrientation: "top",
      limitSelector: { defaultLimit: DEFAULT_SPLIT_LIMIT },
      colorMappingSelector: {
        enable: true,
        values: this.provider.customColorValues,
      },
      nullSelector: true,
    };

    return inputParams;
  }

  createChartDataQuery(
    ctx: CanvasStore,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    return this.provider.createChartDataQuery(ctx.runtime, timeAndFilterStore);
  }

  chartTitle(fields: ChartFieldsMap): string {
    const config = get(this.specStore);
    const xField = fields[config.x?.field || ""];
    const yField = fields[config.y?.field || ""];
    const xTitle = xField?.displayName || config.x?.field || "X";
    const yTitle = yField?.displayName || config.y?.field || "Y";
    return `${xTitle} vs ${yTitle}`;
  }

  getChartDomainValues() {
    return this.provider.getChartDomainValues();
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): ScatterPlotCanvasChartSpec {
    const measures = metricsViewSpec?.measures || [];
    const dimensions = [...(metricsViewSpec?.dimensions || [])].filter(
      (d) => d.type === MetricsViewSpecDimensionType.DIMENSION_TYPE_CATEGORICAL,
    );

    const randomMeasure1 = measures[Math.floor(Math.random() * measures.length)]
      ?.name as string;
    const randomMeasure2 = measures[Math.floor(Math.random() * measures.length)]
      ?.name as string;
    const randomDimension = dimensions[
      Math.floor(Math.random() * dimensions.length)
    ]?.name as string;

    return {
      metrics_view: metricsViewName,
      color: "primary",
      x: {
        type: "quantitative",
        field: randomMeasure1,
        zeroBasedOrigin: false,
      },
      y: {
        type: "quantitative",
        field: randomMeasure2,
        zeroBasedOrigin: false,
      },
      dimension: randomDimension
        ? {
            type: "nominal",
            field: randomDimension,
            limit: DEFAULT_SPLIT_LIMIT,
          }
        : undefined,
    };
  }
}
