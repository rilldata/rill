import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { ComponentWithMetricsView } from "@rilldata/web-common/features/canvas/components/types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { transformTimeAndFiltersToExploreState } from "@rilldata/web-common/features/explores/explore-link/explore-state-transformer";

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

  // Get global transformer properties from time and filter store
  const gTP = transformTimeAndFiltersToExploreState(timeAndFilterStore);

  const partialExploreState: Partial<ExploreState> = {
    ...gTP,
    ...(cTP ?? {}),
  };

  return partialExploreState;
}
