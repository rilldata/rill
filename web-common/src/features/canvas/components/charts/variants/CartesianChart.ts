import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  CartesianChartProvider,
  type CartesianChartSpec as CartesianChartSpecBase,
} from "@rilldata/web-common/features/components/charts/cartesian/CartesianChartProvider";
import {
  type ChartDataQuery,
  type ChartFieldsMap,
  type FieldConfig,
} from "@rilldata/web-common/features/components/charts/types";
import { isMultiFieldConfig } from "@rilldata/web-common/features/components/charts/util";
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

export type CartesianCanvasChartSpec = BaseChartConfig & CartesianChartSpecBase;

const DEFAULT_NOMINAL_LIMIT = 20;
const DEFAULT_SPLIT_LIMIT = 10;
const DEFAULT_SORT = "-y";

export class CartesianChartComponent extends BaseChart<CartesianCanvasChartSpec> {
  private provider: CartesianChartProvider;

  static chartInputParams: Record<string, ComponentInputParam> = {
    x: {
      type: "positional",
      label: "X-axis",
      meta: {
        chartFieldInput: {
          type: "dimension",
          axisTitleSelector: true,
          sortSelector: {
            enable: true,
            defaultSort: DEFAULT_SORT,
            options: ["x", "-x", "y", "-y", "custom"],
          },
          limitSelector: { defaultLimit: DEFAULT_NOMINAL_LIMIT },
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
          type: "measure",
          axisTitleSelector: true,
          originSelector: true,
          axisRangeSelector: true,
          colorMappingSelector: { enable: false },
          multiFieldSelector: true,
        },
      },
    },
    // TODO: Refactor to use simpler primitives
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

    this.provider = new CartesianChartProvider(this.specStore, {
      nominalLimit: DEFAULT_NOMINAL_LIMIT,
      splitLimit: DEFAULT_SPLIT_LIMIT,
      sort: DEFAULT_SORT,
    });

    // Subscribe to provider's combinedWhere
    this.provider.combinedWhere.subscribe((where) => {
      this.componentFilters = where;
    });
  }

  getMeasureLabels(): string[] | undefined {
    const config = get(this.specStore);
    const metricsViewName = config.metrics_view;
    const measuresStore =
      this.parent.metricsView.getMeasuresForMetricView(metricsViewName);
    const measures = get(measuresStore);
    return this.provider.getMeasureLabels(measures);
  }

  getChartSpecificOptions(): Record<string, ComponentInputParam> {
    const inputParams = { ...CartesianChartComponent.chartInputParams };
    const config = get(this.specStore);
    const isMultiMeasure = isMultiFieldConfig(config.y);

    const sortSelector = inputParams.x.meta?.chartFieldInput?.sortSelector;
    if (sortSelector) {
      sortSelector.customSortItems = this.provider.customSortXItems;
    }

    if (isMultiMeasure) {
      inputParams.color.meta!.chartFieldInput = {
        type: "value",
        colorMappingSelector: {
          enable: true,
          values: this.getMeasureLabels(),
        },
        defaultLegendOrientation: "top",
      };

      inputParams.y.meta!.chartFieldInput!.excludedValues = [];
    } else {
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

      // Exclude the main y field from multi-field selector
      if (inputParams.y.meta?.chartFieldInput && config.y?.field) {
        inputParams.y.meta.chartFieldInput.excludedValues = [config.y.field];
      }
    }

    return inputParams;
  }

  updateProperty(
    key: keyof CartesianCanvasChartSpec,
    value: CartesianCanvasChartSpec[keyof CartesianCanvasChartSpec],
  ) {
    const currentSpec = get(this.specStore);

    if (key === "y") {
      const updatedYField = value as FieldConfig;
      const isMultiMeasure = isMultiFieldConfig(updatedYField);

      if (isMultiMeasure) {
        const newSpec = { ...currentSpec, [key]: updatedYField };
        if (typeof currentSpec.color === "string" || !currentSpec.color) {
          newSpec.color = {
            type: "value",
            field: "rill_measures", // dummy field for multi-measure mode
            legendOrientation: "top",
          };
        }

        this.setSpec(newSpec);
        return;
      } else if (!isMultiMeasure) {
        const newSpec = { ...currentSpec, [key]: updatedYField };

        if (
          typeof currentSpec.color === "object" &&
          currentSpec.color?.field === "rill_measures"
        ) {
          newSpec.color = "primary";
        }

        this.setSpec(newSpec);
        return;
      }
    }

    super.updateProperty(key, value);
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
  ): CartesianCanvasChartSpec {
    // Randomly select a measure and dimension if available
    const measures = metricsViewSpec?.measures || [];
    const timeDimension = metricsViewSpec?.timeDimension;
    const dimensions = [...(metricsViewSpec?.dimensions || [])].filter(
      (d) => d.type === MetricsViewSpecDimensionType.DIMENSION_TYPE_CATEGORICAL,
    );

    const randomMeasure = measures[Math.floor(Math.random() * measures.length)]
      ?.name as string;

    let randomDimension = "";
    if (!timeDimension) {
      randomDimension = dimensions[
        Math.floor(Math.random() * dimensions.length)
      ]?.name as string;
    }

    return {
      metrics_view: metricsViewName,
      color: "primary",
      x: {
        type: timeDimension ? "temporal" : "nominal",
        field: timeDimension || randomDimension,
        sort: DEFAULT_SORT,
        limit: DEFAULT_NOMINAL_LIMIT,
      },
      y: {
        type: "quantitative",
        field: randomMeasure,
        zeroBasedOrigin: true,
      },
    };
  }

  chartTitle(fields: ChartFieldsMap) {
    return this.provider.chartTitle(fields);
  }

  getChartDomainValues() {
    const config = get(this.specStore);
    const metricsViewName = config.metrics_view;
    const measuresStore =
      this.parent.metricsView.getMeasuresForMetricView(metricsViewName);
    const measures = get(measuresStore);
    return this.provider.getChartDomainValues(measures);
  }
}
