import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { ComponentWithMetricsView } from "@rilldata/web-common/features/canvas/components/types";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { mapObjectToExploreState } from "@rilldata/web-common/features/explore-mappers/map-to-explore";
import type { ComponentTransformerProperties } from "@rilldata/web-common/features/explore-mappers/types";

export interface CanvasLinkContext {
  instanceId: string;
  organization?: string;
  project?: string;
  metricsViewName: string;
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
  const exploreName = context.exploreName;
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

    const transformerProperties: Partial<ComponentTransformerProperties> = {
      ...gTP,
      ...(cTP ?? {}),
      metricsViewName: context.metricsViewName,
    };

    const executionTime = new Date().toISOString();

    return mapObjectToExploreState(
      exploreName,
      "ComponentTransformer",
      transformerProperties,
      executionTime,
      {},
    );
  }
}
