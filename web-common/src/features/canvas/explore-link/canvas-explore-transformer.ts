import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { ChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
import { isFieldConfig } from "@rilldata/web-common/features/canvas/components/charts/util";
import type { ComponentWithMetricsView } from "@rilldata/web-common/features/canvas/components/types";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type {
  PivotChipData,
  PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

export interface CanvasLinkContext {
  organization?: string;
  project?: string;
  timeAndFilterStore: TimeAndFilterStore;
  exploreName: string;
}

/**
 *  Orchestrator function that transforms canvas component to explore state
 */
export function useTransformCanvasToExploreState(
  component: BaseCanvasComponent<ComponentWithMetricsView>,
  context: CanvasLinkContext,
) {
  const timeAndFilterStore = context.timeAndFilterStore;

  // if (!validateUserPermissions()) {
  //   throw createLinkError(
  //     "PERMISSION_ERROR",
  //     "You do not have permission to access this explore dashboard",
  //   );
  // }

  // Get component-specific transformer properties
  const cTP = component.getExploreTransformerProperties?.();

  // Get global transformer properties
  const gTP: Partial<ExploreState> = {};

  if (timeAndFilterStore.where) {
    const { dimensionFilters, dimensionThresholdFilters } = splitWhereFilter(
      timeAndFilterStore.where,
    );
    gTP.whereFilter = dimensionFilters;
    gTP.dimensionThresholdFilters = dimensionThresholdFilters;
  }

  if (timeAndFilterStore.timeRangeState) {
    gTP.selectedTimeRange = timeAndFilterStore.timeRangeState.selectedTimeRange;
    gTP.selectedTimezone = timeAndFilterStore?.timeRange?.timeZone || "UTC";

    if (timeAndFilterStore.showTimeComparison) {
      gTP.showTimeComparison = true;
      gTP.selectedComparisonTimeRange =
        timeAndFilterStore.comparisonTimeRangeState?.selectedComparisonTimeRange;
    } else {
      gTP.showTimeComparison = false;
      gTP.selectedComparisonTimeRange = undefined;
    }
  }

  const partialExploreState: Partial<ExploreState> = {
    ...gTP,
    ...(cTP ?? {}),
  };

  return partialExploreState;
}

export function getPivotStateFromChartSpec(
  spec: ChartSpec,
  timeGrain: V1TimeGrain | undefined,
): PivotState {
  const columns: PivotChipData[] = [];
  const rows: PivotChipData[] = [];

  // Iterate over all properties in the spec
  for (const [key, value] of Object.entries(spec)) {
    // Skip non-field properties
    if (key === "metrics_view" || key === "title" || key === "description") {
      continue;
    }

    // Check if this property is a field config object
    if (isFieldConfig(value)) {
      const fieldConfig = value;

      let chipType: PivotChipType;
      let id: string;
      if (fieldConfig.type === "quantitative") {
        id = fieldConfig.field;
        chipType = PivotChipType.Measure;
      } else if (fieldConfig.type === "temporal") {
        id = timeGrain || V1TimeGrain.TIME_GRAIN_DAY;
        chipType = PivotChipType.Time;
      } else {
        id = fieldConfig.field;
        chipType = PivotChipType.Dimension;
      }

      if (key === "x" || chipType === PivotChipType.Measure) {
        columns.push({
          id,
          title: fieldConfig.field,
          type: chipType,
        });
      } else {
        rows.push({
          id,
          title: fieldConfig.field,
          type: chipType,
        });
      }
    }
  }

  return {
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
}
