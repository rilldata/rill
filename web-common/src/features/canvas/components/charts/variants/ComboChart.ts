import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  ComboChartProvider,
  type ComboChartSpec as ComboChartSpecBase,
} from "@rilldata/web-common/features/components/charts/combo/ComboChartProvider";
import {
  ChartSortType,
  type ChartFieldsMap,
  type FieldConfig,
} from "@rilldata/web-common/features/components/charts/types";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import {
  PivotChipType,
  type PivotChipData,
  type PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  MetricsViewSpecDimensionType,
  type V1MetricsViewSpec,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { get, type Readable } from "svelte/store";
import type { ChartDataQuery } from "../../../../components/charts/types";
import type {
  CanvasEntity,
  ComponentPath,
} from "../../../stores/canvas-entity";
import { BaseChart, type BaseChartConfig } from "../BaseChart";

export type ComboCanvasChartSpec = BaseChartConfig & ComboChartSpecBase;

const DEFAULT_NOMINAL_LIMIT = 20;
const DEFAULT_SORT = ChartSortType.Y_DESC;

export class ComboChartComponent extends BaseChart<ComboCanvasChartSpec> {
  private provider: ComboChartProvider;

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
            options: [
              ChartSortType.X_ASC,
              ChartSortType.X_DESC,
              ChartSortType.Y_ASC,
              ChartSortType.Y_DESC,
              ChartSortType.CUSTOM,
            ],
          },
          limitSelector: { defaultLimit: DEFAULT_NOMINAL_LIMIT },
          nullSelector: true,
          labelAngleSelector: true,
        },
      },
    },
    y1: {
      type: "positional",
      label: "Left Y-Axis",
      meta: {
        chartFieldInput: {
          type: "measure",
          axisTitleSelector: true,
          originSelector: true,
          axisRangeSelector: true,
          markTypeSelector: true,
        },
      },
    },

    y2: {
      type: "positional",
      label: "Right Y-Axis",
      meta: {
        chartFieldInput: {
          type: "measure",
          axisTitleSelector: true,
          originSelector: true,
          axisRangeSelector: true,
          markTypeSelector: true,
        },
      },
    },

    color: {
      type: "mark",
      label: "Color",
      meta: {
        type: "color",
        chartFieldInput: {
          type: "value",
          defaultLegendOrientation: "top",
          colorMappingSelector: { enable: true },
        },
      },
    },
  };

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    super(resource, parent, path);

    this.provider = new ComboChartProvider(this.specStore, {
      nominalLimit: DEFAULT_NOMINAL_LIMIT,
      sort: DEFAULT_SORT,
    });

    // Subscribe to provider's combinedWhere
    this.provider.combinedWhere.subscribe((where) => {
      this.componentFilters = where;
    });
  }

  updateProperty(
    key: keyof ComboCanvasChartSpec,
    value: ComboCanvasChartSpec[keyof ComboCanvasChartSpec],
  ) {
    const currentSpec = get(this.specStore);

    // Handle mark type mutual exclusivity
    if (key === "y1" || key === "y2") {
      const updatedField = value as FieldConfig;

      if (updatedField?.mark) {
        const otherKey = key === "y1" ? "y2" : "y1";
        const otherField = currentSpec[otherKey];

        // If the other field exists and has the same mark type, switch it
        if (otherField?.mark === updatedField.mark) {
          const oppositeMarkType = updatedField.mark === "bar" ? "line" : "bar";
          const updatedOtherField = { ...otherField, mark: oppositeMarkType };

          const newSpec = {
            ...currentSpec,
            [key]: updatedField,
            [otherKey]: updatedOtherField,
          };

          this.setSpec(newSpec);
          return;
        }
      }
    }
    super.updateProperty(key, value);
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
    const inputParams = { ...ComboChartComponent.chartInputParams };
    const config = get(this.specStore);

    const sortSelector = inputParams.x.meta?.chartFieldInput?.sortSelector;
    if (sortSelector && this.provider) {
      sortSelector.customSortItems = this.provider.customSortXItems;
    }

    const colorMappingSelector =
      inputParams.color.meta?.chartFieldInput?.colorMappingSelector;
    if (colorMappingSelector) {
      colorMappingSelector.values = this.getMeasureLabels();
    }

    if (inputParams.y1.meta?.chartFieldInput && config.y2?.field) {
      inputParams.y1.meta.chartFieldInput.excludedValues = [config.y2.field];
    }

    if (inputParams.y2.meta?.chartFieldInput && config.y1?.field) {
      inputParams.y2.meta.chartFieldInput.excludedValues = [config.y1.field];
    }

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

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): ComboCanvasChartSpec {
    // Randomly select measures and dimension if available
    const measures = metricsViewSpec?.measures || [];
    const timeDimension = metricsViewSpec?.timeDimension;
    const dimensions = [...(metricsViewSpec?.dimensions || [])].filter(
      (d) => d.type === MetricsViewSpecDimensionType.DIMENSION_TYPE_CATEGORICAL,
    );
    const randomMeasure1 = measures[Math.floor(Math.random() * measures.length)]
      ?.name as string;

    // Ensure randomMeasure2 is different from randomMeasure1
    let randomMeasure2: string;
    if (measures.length > 1) {
      do {
        randomMeasure2 = measures[Math.floor(Math.random() * measures.length)]
          ?.name as string;
      } while (randomMeasure2 === randomMeasure1);
    } else {
      randomMeasure2 = "Other_measure";
    }

    let randomDimension = "";
    if (!timeDimension) {
      randomDimension = dimensions[
        Math.floor(Math.random() * dimensions.length)
      ]?.name as string;
    }

    return {
      metrics_view: metricsViewName,
      x: {
        type: timeDimension ? "temporal" : "nominal",
        field: timeDimension || randomDimension,
        sort: DEFAULT_SORT,
        limit: DEFAULT_NOMINAL_LIMIT,
      },
      y1: {
        type: "quantitative",
        field: randomMeasure1,
        zeroBasedOrigin: true,
        mark: "bar",
      },
      y2: {
        type: "quantitative",
        field: randomMeasure2,
        zeroBasedOrigin: true,
        mark: "line",
      },
      color: {
        type: "value",
        field: "measures",
        legendOrientation: "top",
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

  override getExploreTransformerProperties(): Partial<ExploreState> {
    const spec = get(this.specStore);
    const { dimensionFilters, dimensionThresholdFilters } = splitWhereFilter(
      this.componentFilters,
    );
    const timeGrain = get(this.timeAndFilterStore)?.timeGrain;

    const columns: PivotChipData[] = [];
    const rows: PivotChipData[] = [];

    if (spec.x?.field) {
      if (spec.x.type === "temporal") {
        rows.push({
          id: timeGrain || V1TimeGrain.TIME_GRAIN_DAY,
          title: spec.x.field,
          type: PivotChipType.Time,
        });
      } else {
        rows.push({
          id: spec.x.field,
          title: spec.x.field,
          type: PivotChipType.Dimension,
        });
      }
    }

    if (spec.y1?.field && spec.y1.type === "quantitative") {
      columns.push({
        id: spec.y1.field,
        title: spec.y1.field,
        type: PivotChipType.Measure,
      });
    }

    if (spec.y2?.field && spec.y2.type === "quantitative") {
      columns.push({
        id: spec.y2.field,
        title: spec.y2.field,
        type: PivotChipType.Measure,
      });
    }

    const hasDimensionRows = rows.some(
      (r) => r.type === PivotChipType.Dimension,
    );
    if (hasDimensionRows) {
      const timeChips = rows.filter((r) => r.type === PivotChipType.Time);
      const nonTimeRows = rows.filter((r) => r.type !== PivotChipType.Time);
      columns.push(...timeChips);
      rows.length = 0;
      rows.push(...nonTimeRows);
    }

    const pivot: PivotState = {
      columns,
      rows,
      expanded: {},
      sorting: [],
      columnPage: 0,
      rowPage: 0,
      enableComparison: false,
      tableMode: "nest",
      activeCell: null,
    };

    return {
      whereFilter: dimensionFilters,
      dimensionThresholdFilters,
      showTimeComparison: false,
      activePage: DashboardState_ActivePage.PIVOT,
      pivot,
    };
  }
}
