import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  CircularChartProvider,
  type CircularChartSpec as CircularChartSpecBase,
} from "@rilldata/web-common/features/components/charts/circular/CircularChartProvider";
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

const DEFAULT_COLOR_LIMIT = 20;
const DEFAULT_SORT = "-measure";

export type CircularCanvasChartSpec = BaseChartConfig & CircularChartSpecBase;

export class CircularChartComponent extends BaseChart<CircularCanvasChartSpec> {
  private provider: CircularChartProvider;

  static chartInputParams: Record<string, ComponentInputParam> = {
    measure: {
      type: "positional",
      label: "Measure",
      meta: {
        chartFieldInput: {
          type: "measure",
          totalSelector: true,
        },
      },
    },
    innerRadius: {
      type: "number",
      label: "Inner Radius (%)",
    },
    color: {
      type: "positional",
      label: "Color",
      meta: {
        chartFieldInput: {
          type: "dimension",
          nullSelector: true,
          limitSelector: { defaultLimit: DEFAULT_COLOR_LIMIT },
          hideTimeDimension: true,
          defaultLegendOrientation: "right",
          sortSelector: {
            enable: true,
            defaultSort: DEFAULT_SORT,
            options: ["color", "-color", "measure", "-measure", "custom"],
          },
          colorMappingSelector: { enable: true },
        },
      },
    },
  };

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    super(resource, parent, path);

    this.provider = new CircularChartProvider(this.specStore, {
      colorLimit: DEFAULT_COLOR_LIMIT,
      colorSort: DEFAULT_SORT,
    });

    // Subscribe to provider's combinedWhere
    this.provider.combinedWhere.subscribe((where) => {
      this.componentFilters = where;
    });
  }

  getChartSpecificOptions(): Record<string, ComponentInputParam> {
    const inputParams = CircularChartComponent.chartInputParams;
    const colorMappingSelector =
      inputParams.color.meta?.chartFieldInput?.colorMappingSelector;
    if (colorMappingSelector) {
      colorMappingSelector.values = this.provider.customColorValues;
    }
    return inputParams;
  }

  createChartDataQuery(
    ctx: CanvasStore,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery {
    return this.provider.createChartDataQuery(ctx.runtime, timeAndFilterStore);
  }

  getChartDomainValues() {
    return this.provider.getChartDomainValues();
  }

  chartTitle(fields: ChartFieldsMap) {
    return this.provider.chartTitle(fields);
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): CircularCanvasChartSpec {
    // Randomly select a measure and dimension if available
    const measures = metricsViewSpec?.measures || [];
    const dimensions = [...(metricsViewSpec?.dimensions || [])].filter(
      (d) => d.type === MetricsViewSpecDimensionType.DIMENSION_TYPE_CATEGORICAL,
    );
    const randomMeasure = measures[Math.floor(Math.random() * measures.length)]
      ?.name as string;

    const randomDimension = dimensions[
      Math.floor(Math.random() * dimensions.length)
    ]?.name as string;

    return {
      metrics_view: metricsViewName,
      innerRadius: 50,
      color: {
        type: "nominal",
        field: randomDimension,
        limit: DEFAULT_COLOR_LIMIT,
        sort: DEFAULT_SORT,
      },
      measure: {
        type: "quantitative",
        field: randomMeasure,
        showTotal: true,
      },
    };
  }
}
