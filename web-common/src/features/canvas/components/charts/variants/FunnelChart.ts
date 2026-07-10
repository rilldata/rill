import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import {
  FunnelChartProvider,
  type FunnelBreakdownMode,
  type FunnelChartSpec as FunnelChartSpecBase,
} from "@rilldata/web-common/features/components/charts/funnel/FunnelChartProvider";
import {
  ChartSortType,
  type ChartFieldsMap,
} from "@rilldata/web-common/features/components/charts/types";
import { isMultiFieldConfig } from "@rilldata/web-common/features/components/charts/util";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  MetricsViewSpecDimensionType,
  type V1MetricsViewSpec,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { get, type Readable } from "svelte/store";
import type { ChartDataQuery } from "../../../../components/charts/types";
import type {
  CanvasEntity,
  ComponentPath,
} from "../../../stores/canvas-entity";
import { BaseChart, type BaseChartConfig } from "../BaseChart";

import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

const DEFAULT_STAGE_LIMIT = 15;
const DEFAULT_SORT = ChartSortType.Y_DESC;

export type FunnelCanvasChartSpec = BaseChartConfig & FunnelChartSpecBase;

export class FunnelChartComponent extends BaseChart<FunnelCanvasChartSpec> {
  private provider: FunnelChartProvider;

  // Static getter (not a static field) so the localized labels inside resolve
  // in the active locale at access time (render) rather than freezing to the
  // locale active when this class was defined at module load.
  static get chartInputParams(): Record<string, ComponentInputParam> {
    return {
      breakdownMode: {
        type: "switcher_tab",
        label: m.canvas_breakdown_by_label(),
        meta: {
          default: "dimension",
          options: [
            { label: m.canvas_dimension_label(), value: "dimension" },
            { label: m.canvas_measures_label(), value: "measures" },
          ],
        },
      },
      stage: {
        type: "positional",
        label: m.canvas_stage_label(),
        meta: {
          chartFieldInput: {
            type: "dimension",
            nullSelector: true,
            sortSelector: {
              enable: true,
              defaultSort: DEFAULT_SORT,
              options: [
                ChartSortType.Y_ASC,
                ChartSortType.Y_DESC,
                ChartSortType.CUSTOM,
              ],
            },
            limitSelector: { defaultLimit: DEFAULT_STAGE_LIMIT },
            hideTimeDimension: true,
          },
        },
      },
      measure: {
        type: "positional",
        label: m.canvas_measure_label(),
        meta: {
          chartFieldInput: {
            type: "measure",
          },
        },
      },
      mode: {
        type: "select",
        label: m.canvas_mode_label(),
        meta: {
          default: "width",
          options: [
            { label: m.canvas_width_option(), value: "width" },
            { label: m.canvas_order_option(), value: "order" },
          ],
        },
      },
      color: {
        type: "switcher_tab",
        label: m.canvas_color_label(),
        meta: {
          default: "stage",
          options: [
            { label: m.canvas_stage_label(), value: "stage" },
            { label: m.canvas_measure_label(), value: "measure" },
          ],
        },
      },
      percentMode: {
        type: "switcher_tab",
        label: m.canvas_percent_of_label(),
        meta: {
          default: "top",
          options: [
            { label: m.canvas_top_option(), value: "top" },
            { label: m.canvas_previous_option(), value: "previous" },
          ],
        },
      },
    };
  }

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    super(resource, parent, path);

    this.provider = new FunnelChartProvider(this.specStore, {
      stageLimit: DEFAULT_STAGE_LIMIT,
      sort: DEFAULT_SORT,
    });

    // Subscribe to provider's combinedWhere
    this.provider.combinedWhere.subscribe((where) => {
      this.componentFilters = where;
    });
  }

  getChartSpecificOptions(): Record<string, ComponentInputParam> {
    const inputParams = { ...FunnelChartComponent.chartInputParams };
    const config = get(this.specStore);
    const isMultiMeasure = config.breakdownMode === "measures";

    const sortSelector = inputParams.stage.meta?.chartFieldInput?.sortSelector;
    if (sortSelector && this.provider) {
      sortSelector.customSortItems = this.provider.customSortStageItems;
    }

    if (isMultiMeasure) {
      // In measures mode, hide stage field and update measure field for multi-selection
      inputParams.stage.showInUI = false;
      inputParams.measure.meta!.chartFieldInput = {
        type: "measure",
        multiFieldSelector: true,
      };

      // Update color field for measures mode: Name (discrete) and Value (continuous)
      inputParams.color.meta!.options = [
        { label: m.canvas_name_option(), value: "name" },
        { label: m.canvas_value_option(), value: "value" },
      ];
    } else {
      // In dimension mode, show stage field and single measure selection
      inputParams.stage.showInUI = true;
      inputParams.measure.meta!.chartFieldInput = {
        type: "measure",
      };

      // Update color field for dimension mode
      inputParams.color.meta!.options = [
        { label: m.canvas_stage_label(), value: "stage" },
        { label: m.canvas_measure_label(), value: "measure" },
      ];

      // Exclude the main measure field from multi-field selector
      if (inputParams.measure.meta?.chartFieldInput && config.measure?.field) {
        inputParams.measure.meta.chartFieldInput.excludedValues = [
          config.measure.field,
        ];
      }
    }

    return inputParams;
  }

  updateProperty(
    key: keyof FunnelCanvasChartSpec,
    value: FunnelCanvasChartSpec[keyof FunnelCanvasChartSpec],
  ) {
    const currentSpec = get(this.specStore);

    if (key === "breakdownMode") {
      const newBreakdownMode = value as FunnelBreakdownMode;
      const newSpec = { ...currentSpec, [key]: newBreakdownMode };

      if (newBreakdownMode === "measures") {
        if (currentSpec.measure?.field) {
          newSpec.measure = {
            type: "quantitative",
            field: currentSpec.measure.field,
          };
        }
        newSpec.stage = undefined;
        newSpec.color = "name";
      } else {
        if (isMultiFieldConfig(currentSpec.measure)) {
          const firstMeasure = currentSpec.measure.fields?.[0];
          if (firstMeasure) {
            newSpec.measure = {
              type: "quantitative",
              field: firstMeasure,
            };
          }
        }
        if (currentSpec.color === "name" || currentSpec.color === "value") {
          newSpec.color = "stage";
        }

        const dimensionsStore =
          this.parent.metricsView.getDimensionsForMetricView(
            currentSpec.metrics_view,
          );
        const dimensions = get(dimensionsStore);
        if (dimensions?.length) {
          newSpec.stage = {
            field: dimensions[0].name || (dimensions[0].column as string),
            type: "nominal",
          };
        }
      }

      this.setSpec(newSpec);
      return;
    }

    super.updateProperty(key, value);
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

  chartTitle(fields: ChartFieldsMap) {
    return this.provider.chartTitle(fields);
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): FunnelCanvasChartSpec {
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
      stage: {
        type: "nominal",
        field: randomDimension,
        limit: DEFAULT_STAGE_LIMIT,
      },
      measure: {
        type: "quantitative",
        field: randomMeasure,
      },
      mode: "width",
      color: "stage",
      breakdownMode: "dimension",
      percentMode: "top",
    };
  }
}
