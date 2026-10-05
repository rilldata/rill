import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  CartesianChartProvider,
  type CartesianChartSpec as CartesianChartSpecBase,
} from "@rilldata/web-common/features/components/charts/cartesian/CartesianChartProvider";
import {
  chartOrientation,
  dimensionChannel,
  isHorizontal,
  measureChannel,
  supportsOrientation,
  swapAxes,
  swapSortChannel,
  toVerticalSpec,
  type ChartOrientation,
} from "@rilldata/web-common/features/components/charts/cartesian/orientation";
import {
  ChartSortType,
  type ChartDataQuery,
  type ChartFieldsMap,
  type ChartType,
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

import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

export type CartesianCanvasChartSpec = BaseChartConfig & CartesianChartSpecBase;

const DEFAULT_NOMINAL_LIMIT = 20;
const DEFAULT_SPLIT_LIMIT = 10;
const DEFAULT_SORT = ChartSortType.Y_DESC;

// Inspector-only param: it is never written to the YAML (see `updateProperty`).
const ORIENTATION_PARAM = "orientation";

// The dimension picker's sort options, written for the dimension on x.
const DIMENSION_SORT_OPTIONS = [
  ChartSortType.X_ASC,
  ChartSortType.X_DESC,
  ChartSortType.Y_ASC,
  ChartSortType.Y_DESC,
  ChartSortType.Y_DELTA_ASC,
  ChartSortType.Y_DELTA_DESC,
  ChartSortType.CUSTOM,
];

export class CartesianChartComponent extends BaseChart<CartesianCanvasChartSpec> {
  private provider: CartesianChartProvider;

  // Static getter (not a static field) so the localized labels inside resolve
  // in the active locale at access time (render) rather than freezing to the
  // locale active when this class was defined at module load.
  //
  // The `x` and `y` params describe the vertical layout (dimension on x,
  // measure on y). `getChartSpecificOptions` swaps their contents for a
  // horizontal bar chart so that each picker keeps its role and the YAML keys
  // keep naming the axis a field is drawn on.
  static get chartInputParams(): Record<string, ComponentInputParam> {
    return {
      x: {
        type: "positional",
        label: m.canvas_dimension_label(),
        meta: {
          axisLabel: m.canvas_x_axis_label(),
          chartFieldInput: {
            type: "dimension",
            axisTitleSelector: true,
            sortSelector: {
              enable: true,
              defaultSort: DEFAULT_SORT,
              options: DIMENSION_SORT_OPTIONS,
            },
            limitSelector: { defaultLimit: DEFAULT_NOMINAL_LIMIT },
            nullSelector: true,
            labelAngleSelector: true,
          },
        },
      },
      y: {
        type: "positional",
        label: m.canvas_measure_label(),
        meta: {
          axisLabel: m.canvas_y_axis_label(),
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
        label: m.canvas_color_label(),
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
      // Not a YAML property: toggling it swaps the `x` and `y` fields.
      [ORIENTATION_PARAM]: {
        type: "switcher_tab",
        label: m.canvas_orientation_label(),
        meta: {
          default: "vertical",
          options: [
            { label: m.canvas_vertical_option(), value: "vertical" },
            { label: m.canvas_horizontal_option(), value: "horizontal" },
          ],
        },
      },
    };
  }

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

  protected supportsComparison(): boolean {
    return true;
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
    const staticParams = CartesianChartComponent.chartInputParams;
    const config = get(this.specStore);
    const horizontal = isHorizontal(config);
    const dimensionKey = dimensionChannel(config);
    const measureKey = measureChannel(config);

    const dimensionParam = staticParams.x;
    const measureParam = staticParams.y;
    if (horizontal) {
      // The pickers keep their role; only the axis they drive changes, and
      // channel-relative sort options follow the dimension to the y channel.
      dimensionParam.meta!.axisLabel = m.canvas_y_axis_label();
      measureParam.meta!.axisLabel = m.canvas_x_axis_label();
      const sortSelector = dimensionParam.meta!.chartFieldInput!.sortSelector!;
      sortSelector.defaultSort = swapSortChannel(DEFAULT_SORT);
      sortSelector.options = DIMENSION_SORT_OPTIONS.map(swapSortChannel);
    }

    // Dimension first, then measure, whichever channel each one is on.
    const inputParams: Record<string, ComponentInputParam> = {
      [dimensionKey]: dimensionParam,
      [measureKey]: measureParam,
      color: staticParams.color,
    };

    // Only bar charts can be drawn horizontally.
    if (supportsOrientation(this.type)) {
      inputParams[ORIENTATION_PARAM] = {
        ...staticParams[ORIENTATION_PARAM],
        meta: {
          ...staticParams[ORIENTATION_PARAM].meta,
          value: chartOrientation(config),
        },
      };
    }

    const sortSelector = dimensionParam.meta?.chartFieldInput?.sortSelector;
    if (sortSelector) {
      sortSelector.customSortItems = this.provider.customSortXItems;
    }

    const measure = config[measureKey];
    if (isMultiFieldConfig(measure)) {
      inputParams.color.meta!.chartFieldInput = {
        type: "value",
        colorMappingSelector: {
          enable: true,
          values: this.getMeasureLabels(),
        },
        defaultLegendOrientation: "top",
      };

      measureParam.meta!.chartFieldInput!.excludedValues = [];
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

      // Exclude the main measure field from the multi-field selector
      if (measureParam.meta?.chartFieldInput && measure?.field) {
        measureParam.meta.chartFieldInput.excludedValues = [measure.field];
      }
    }

    return inputParams;
  }

  updateProperty = (
    key: keyof CartesianCanvasChartSpec,
    value: CartesianCanvasChartSpec[keyof CartesianCanvasChartSpec],
  ) => {
    const currentSpec = get(this.specStore);

    if ((key as string) === ORIENTATION_PARAM) {
      // Orientation lives in which axis holds the measure, so changing it
      // swaps the `x` and `y` fields instead of writing a property. A stray
      // `orientation` key in the YAML is dropped so it cannot shadow the toggle.
      const changed =
        (value as ChartOrientation) !== chartOrientation(currentSpec);
      const spec = { ...currentSpec };
      delete (spec as Record<string, unknown>)[ORIENTATION_PARAM];
      this.setSpec(
        changed ? (swapAxes(spec) as CartesianCanvasChartSpec) : spec,
      );
      return;
    }

    if (key === measureChannel(currentSpec)) {
      const updatedMeasureField = value as FieldConfig;
      const isMultiMeasure = isMultiFieldConfig(updatedMeasureField);

      if (isMultiMeasure) {
        const newSpec = { ...currentSpec, [key]: updatedMeasureField };
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
        const newSpec = { ...currentSpec, [key]: updatedMeasureField };

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
  };

  protected specForChartTypeSwitch(
    spec: CartesianCanvasChartSpec,
    targetType: ChartType,
  ): CartesianCanvasChartSpec {
    // Only bar charts read a measure on x; every other target expects the
    // vertical layout.
    return supportsOrientation(targetType)
      ? spec
      : (toVerticalSpec(spec) as CartesianCanvasChartSpec);
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
