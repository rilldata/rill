import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  CircularChartProvider,
  type CircularChartSpec as CircularChartSpecBase,
} from "@rilldata/web-common/features/components/charts/circular/CircularChartProvider";
import { DEFAULT_LABELS_THRESHOLD } from "@rilldata/web-common/features/components/charts/circular/constants";
import {
  ChartSortType,
  type ChartFieldsMap,
} from "@rilldata/web-common/features/components/charts/types";
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
const DEFAULT_SORT = ChartSortType.MEASURE_DESC;

export type CircularCanvasChartSpec = BaseChartConfig & CircularChartSpecBase;

export class CircularChartComponent extends BaseChart<CircularCanvasChartSpec> {
  private provider: CircularChartProvider;
  private isTruncated = false;

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
    show_other: {
      type: "boolean",
      label: 'Show "Other" bucket',
    },
    labels: {
      type: "labels",
      label: "Data labels",
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
            options: [
              ChartSortType.COLOR_ASC,
              ChartSortType.COLOR_DESC,
              ChartSortType.MEASURE_ASC,
              ChartSortType.MEASURE_DESC,
              ChartSortType.CUSTOM,
            ],
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

    this.provider.isTruncated.subscribe((value) => {
      this.isTruncated = value;
    });
  }

  getChartSpecificOptions(): Record<string, ComponentInputParam> {
    const inputParams = CircularChartComponent.chartInputParams;
    const colorMappingSelector =
      inputParams.color.meta?.chartFieldInput?.colorMappingSelector;
    if (colorMappingSelector) {
      colorMappingSelector.values = this.provider.customColorValues;
    }

    inputParams.show_other.showInUI = this.isTruncated;

    return inputParams;
  }

  createChartDataQuery(
    ctx: CanvasStore,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
    visible: Readable<boolean>,
  ): ChartDataQuery {
    return this.provider.createChartDataQuery(
      ctx.runtimeClient,
      timeAndFilterStore,
      visible,
    );
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
    const measures = metricsViewSpec?.measures || [];
    const dimensions = [...(metricsViewSpec?.dimensions || [])].filter(
      (d) => d.type === MetricsViewSpecDimensionType.DIMENSION_TYPE_CATEGORICAL,
    );

    // Prefer summable measures since percentage-style measures don't make
    // sense as slices of a whole.
    const summableMeasures = measures.filter((m) => m.validPercentOfTotal);
    const measurePool = summableMeasures.length ? summableMeasures : measures;
    const randomMeasure = measurePool[
      Math.floor(Math.random() * measurePool.length)
    ]?.name as string;

    const randomDimension = dimensions[
      Math.floor(Math.random() * dimensions.length)
    ]?.name as string;

    return {
      metrics_view: metricsViewName,
      innerRadius: 50,
      show_other: true,
      labels: {
        show: true,
        format: "percent",
        threshold: DEFAULT_LABELS_THRESHOLD,
      },
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
