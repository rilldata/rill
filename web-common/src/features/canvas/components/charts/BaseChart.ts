import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import { CHART_CONFIG } from "@rilldata/web-common/features/canvas/components/charts";
import {
  commonOptions,
  createComponent,
  getFilterOptions,
} from "@rilldata/web-common/features/canvas/components/util";
import { getPivotStateFromChartSpec } from "@rilldata/web-common/features/canvas/explore-link/canvas-explore-transformer";
import type {
  AllKeys,
  ComponentInputParam,
  InputParams,
} from "@rilldata/web-common/features/canvas/inspector/types";
import type { CanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  V1Expression,
  V1MetricsViewSpec,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import { get, writable, type Readable, type Writable } from "svelte/store";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";
import type {
  ComponentCommonProperties,
  ComponentFilterProperties,
} from "../types";
import Chart from "./Chart.svelte";
import type {
  ChartDataQuery,
  ChartFieldsMap,
  ChartType,
  CommonChartProperties,
  FieldConfig,
} from "./types";

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
  combinedWhere: V1Expression | undefined;
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

  abstract getChartSpecificOptions(): Record<
    AllKeys<TConfig>,
    ComponentInputParam
  >;

  abstract createChartDataQuery(
    ctx: CanvasStore,
    timeAndFilterStore: Readable<TimeAndFilterStore>,
  ): ChartDataQuery;

  abstract chartTitle(fields: ChartFieldsMap): string;

  protected getDefaultFieldConfig(): Partial<FieldConfig> {
    return {
      showAxisTitle: true,
      zeroBasedOrigin: true,
      showNull: false,
    };
  }

  getExploreTransformerProperties(): Partial<ExploreState> {
    const spec = get(this.specStore);
    const { dimensionFilters, dimensionThresholdFilters } = splitWhereFilter(
      this.combinedWhere,
    );

    const timeGrain = get(this.timeAndFilterStore)?.timeGrain;

    return {
      whereFilter: dimensionFilters,
      dimensionThresholdFilters,
      showTimeComparison: false,
      activePage: DashboardState_ActivePage.PIVOT,
      pivot: getPivotStateFromChartSpec(spec, timeGrain),
    };
  }

  updateChartType(
    key: ChartType,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ) {
    if (!this.parent.fileArtifact) return;

    const currentSpec = get(this.specStore);
    const parentPath = this.pathInYAML.slice(0, -1);

    const parseDocumentStore = this.parent.parsedContent;
    const parsedDocument = get(parseDocumentStore);
    const { updateEditorContent } = this.parent.fileArtifact;

    const newSpecForKey = CHART_CONFIG[key].component.newComponentSpec(
      currentSpec.metrics_view,
      metricsViewSpec,
    );

    const commonProps = this.extractCommonProperties(
      currentSpec,
      this.type,
      key,
    );
    const mergedSpec = {
      ...newSpecForKey,
      ...commonProps,
    };

    const newResource = this.parent.createOptimisticResource({
      type: key,
      row: this.pathInYAML[1],
      column: this.pathInYAML[3],
      metricsViewName: currentSpec.metrics_view,
      metricsViewSpec,
      spec: mergedSpec,
    });

    const newComponent = createComponent(
      newResource,
      this.parent,
      this.pathInYAML,
    );

    this.parent.components.set(newComponent.id, newComponent);
    this.parent.selectedComponent.set(newComponent.id);
    this.parent._rows.refresh();

    // Preserve the width from the current chart
    const width = parsedDocument.getIn([...parentPath, "width"]);

    parsedDocument.setIn(parentPath, { [key]: mergedSpec, width });

    updateEditorContent(parsedDocument.toString(), false, true);

    this.chartType.set(key);
  }

  private extractCommonProperties(
    spec: TConfig,
    sourceType: ChartType,
    targetType: ChartType,
  ): Partial<BaseChartConfig> {
    const {
      metrics_view,
      title,
      description,
      vl_config,
      time_filters,
      dimension_filters,
    } = spec;

    const sourceChartParams =
      CHART_CONFIG[sourceType].component.chartInputParams || {};
    const targetChartParams =
      CHART_CONFIG[targetType].component.chartInputParams || {};

    // Check for common keys and type match first
    const commonProps = Object.keys(sourceChartParams).filter((key) => {
      const isKeyAndTypeMatch =
        targetChartParams?.[key]?.type === sourceChartParams[key]?.type;
      const isFieldTypeMatch =
        targetChartParams?.[key]?.meta?.chartFieldInput?.type ===
        sourceChartParams[key]?.meta?.chartFieldInput?.type;
      return isKeyAndTypeMatch && isFieldTypeMatch;
    });

    const commonPropsObject = commonProps.reduce(
      (acc, key) => {
        acc[key] = spec[key];
        return acc;
      },
      {} as Record<string, unknown>,
    );

    return {
      metrics_view,
      title,
      description,
      vl_config,
      time_filters,
      dimension_filters,
      ...commonPropsObject,
    };
  }
}
