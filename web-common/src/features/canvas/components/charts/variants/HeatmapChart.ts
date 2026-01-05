import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  HeatmapChartProvider,
  type HeatmapChartSpec as HeatmapChartSpecBase,
} from "@rilldata/web-common/features/components/charts/heatmap/HeatmapChartProvider";
import type { ChartFieldsMap } from "@rilldata/web-common/features/components/charts/types";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  MetricsViewSpecDimensionType,
  type V1MetricsViewSpec,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { type Readable } from "svelte/store";
import type { ChartDataQuery } from "../../../../components/charts/types";
import type {
  CanvasEntity,
  ComponentPath,
} from "../../../stores/canvas-entity";
import { BaseChart, type BaseChartConfig } from "../BaseChart";

const DEFAULT_NOMINAL_LIMIT = 40;
const DEFAULT_SORT = "-color";

export type HeatmapCanvasChartSpec = BaseChartConfig & HeatmapChartSpecBase;

export class HeatmapChartComponent extends BaseChart<HeatmapCanvasChartSpec> {
  private provider: HeatmapChartProvider;

  static chartInputParams: Record<string, ComponentInputParam> = {
    x: {
      type: "positional",
      label: "X-axis",
      meta: {
        chartFieldInput: {
          type: "dimension",
          limitSelector: { defaultLimit: DEFAULT_NOMINAL_LIMIT },
          sortSelector: {
            enable: true,
            defaultSort: DEFAULT_SORT,
            options: ["x", "-x", "color", "-color", "custom"],
          },
          axisTitleSelector: true,
          nullSelector: true,
          labelAngleSelector: true,
        },
      },
    },
    y: {
      type: "positional",
      label: "Y-axis",
      meta: {
        chartFieldInput: {
          type: "dimension",
          limitSelector: { defaultLimit: DEFAULT_NOMINAL_LIMIT },
          sortSelector: {
            enable: true,
            defaultSort: DEFAULT_SORT,
            options: ["y", "-y", "color", "-color", "custom"],
          },
          axisTitleSelector: true,
          nullSelector: true,
        },
      },
    },
    color: {
      type: "positional",
      label: "Color",
      meta: {
        chartFieldInput: {
          type: "measure",
          defaultLegendOrientation: "right",
          colorRangeSelector: {
            enable: true,
          },
        },
      },
    },
    show_data_labels: {
      type: "boolean",
      label: "Data labels",
    },
  };

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    super(resource, parent, path);

    this.provider = new HeatmapChartProvider(this.specStore, {
      nominalLimit: DEFAULT_NOMINAL_LIMIT,
      sort: DEFAULT_SORT,
    });

    // Subscribe to provider's combinedWhere
    this.provider.combinedWhere.subscribe((where) => {
      this.componentFilters = where;
    });
  }

  getChartSpecificOptions(): Record<string, ComponentInputParam> {
    const inputParams = HeatmapChartComponent.chartInputParams;
    const xSortSelector = inputParams.x.meta?.chartFieldInput?.sortSelector;
    if (xSortSelector && this.provider) {
      xSortSelector.customSortItems = this.provider.customSortXItems;
    }
    const ySortSelector = inputParams.y.meta?.chartFieldInput?.sortSelector;
    if (ySortSelector && this.provider) {
      ySortSelector.customSortItems = this.provider.customSortYItems;
    }
    return inputParams;
  }

  createChartDataQuery(
    ctx: CanvasStore,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    return this.provider.createChartDataQuery(ctx.runtime, timeAndFilterStore);
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): HeatmapCanvasChartSpec {
    // Select two dimensions and one measure if available
    const measures = metricsViewSpec?.measures || [];
    const dimensions = [...(metricsViewSpec?.dimensions || [])].filter(
      (d) => d.type === MetricsViewSpecDimensionType.DIMENSION_TYPE_CATEGORICAL,
    );
    const randomMeasure = measures[Math.floor(Math.random() * measures.length)]
      ?.name as string;

    // Get two random dimensions
    const availableDimensions = [...dimensions];
    const randomDimension1 = availableDimensions.splice(
      Math.floor(Math.random() * availableDimensions.length),
      1,
    )[0]?.name as string;
    const randomDimension2 = availableDimensions[
      Math.floor(Math.random() * availableDimensions.length)
    ]?.name as string;

    return {
      metrics_view: metricsViewName,
      x: {
        type: "nominal",
        field: randomDimension1,
        limit: DEFAULT_NOMINAL_LIMIT,
      },
      y: {
        type: "nominal",
        field: randomDimension2,
        limit: DEFAULT_NOMINAL_LIMIT,
      },
      color: {
        type: "quantitative",
        field: randomMeasure,
        colorRange: {
          mode: "scheme",
          scheme: "sequential", // Use sequential palette for ordered data
        },
      },
    };
  }

  chartTitle(fields: ChartFieldsMap) {
    return this.provider.chartTitle(fields);
  }

  getChartDomainValues() {
    return this.provider.getChartDomainValues();
  }
}
