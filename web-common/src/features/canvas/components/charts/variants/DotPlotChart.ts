import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  DotPlotChartProvider,
  type DotPlotChartSpec as DotPlotChartSpecBase,
} from "@rilldata/web-common/features/components/charts/dot-plot/DotPlotChartProvider";
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
import { type Readable } from "svelte/store";
import type {
  CanvasEntity,
  ComponentPath,
} from "../../../stores/canvas-entity";
import { BaseChart, type BaseChartConfig } from "../BaseChart";

export type DotPlotCanvasChartSpec = BaseChartConfig & DotPlotChartSpecBase;

const DEFAULT_NOMINAL_LIMIT = 20;
const DEFAULT_SPLIT_LIMIT = 10;
const DEFAULT_SORT = "-x";

export class DotPlotChartComponent extends BaseChart<DotPlotCanvasChartSpec> {
  private provider: DotPlotChartProvider;

  static chartInputParams: Record<string, ComponentInputParam> = {
    y: {
      type: "positional",
      label: "Y-axis (Dimension)",
      meta: {
        chartFieldInput: {
          type: "dimension",
          axisTitleSelector: true,
          sortSelector: {
            enable: true,
            defaultSort: DEFAULT_SORT,
            options: ["y", "-y", "x", "-x", "custom"],
          },
          limitSelector: { defaultLimit: DEFAULT_NOMINAL_LIMIT },
          nullSelector: true,
          labelAngleSelector: true,
        },
      },
    },
    x: {
      type: "positional",
      label: "X-axis (Measure)",
      meta: {
        chartFieldInput: {
          type: "measure",
          axisTitleSelector: true,
          axisRangeSelector: true,
        },
      },
    },
    jitter: {
      type: "boolean",
      label: "Jitter points",
      meta: {
        invertBoolean: false,
      },
    },
    detail: {
      type: "positional",
      label: "Detail (Dimension for dots)",
      meta: {
        chartFieldInput: {
          type: "dimension",
          nullSelector: true,
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

    this.provider = new DotPlotChartProvider(this.specStore, {
      nominalLimit: DEFAULT_NOMINAL_LIMIT,
      splitLimit: DEFAULT_SPLIT_LIMIT,
      sort: DEFAULT_SORT,
    });

    this.provider.combinedWhere.subscribe((where) => {
      this.componentFilters = where;
    });
  }

  getChartSpecificOptions(): Record<string, ComponentInputParam> {
    return { ...DotPlotChartComponent.chartInputParams };
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
  ): DotPlotCanvasChartSpec {
    const measures = metricsViewSpec?.measures || [];
    const dimensions = [...(metricsViewSpec?.dimensions || [])].filter(
      (d) => d.type === MetricsViewSpecDimensionType.DIMENSION_TYPE_CATEGORICAL,
    );

    const randomMeasure = measures.length > 0
      ? measures[Math.floor(Math.random() * measures.length)]?.name
      : undefined;

    const randomYDimension = dimensions.length > 0
      ? dimensions[Math.floor(Math.random() * dimensions.length)]?.name
      : undefined;

    const remainingDimensions = dimensions.filter(
      (d) => d.name !== randomYDimension,
    );
    const randomDetailDimension = remainingDimensions.length > 0
      ? remainingDimensions[Math.floor(Math.random() * remainingDimensions.length)]?.name
      : undefined;

    return {
      metrics_view: metricsViewName,
      color: "primary",
      jitter: false,
      ...(randomYDimension && {
        y: {
          type: "nominal",
          field: randomYDimension,
          sort: DEFAULT_SORT,
          limit: DEFAULT_NOMINAL_LIMIT,
        },
      }),
      ...(randomMeasure && {
        x: {
          type: "quantitative",
          field: randomMeasure,
        },
      }),
      ...(randomDetailDimension && {
        detail: {
          type: "nominal",
          field: randomDetailDimension,
        },
      }),
    };
  }

  chartTitle(fields: ChartFieldsMap) {
    return this.provider.chartTitle(fields);
  }

  getChartDomainValues() {
    return this.provider.getChartDomainValues();
  }
}

