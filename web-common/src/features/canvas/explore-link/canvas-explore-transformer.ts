import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { ComponentWithMetricsView } from "@rilldata/web-common/features/canvas/components/types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { transformTimeAndFiltersToExploreState } from "@rilldata/web-common/features/explores/explore-link/explore-state-transformer";
import type { ExpressionState } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";

/**
 *  Orchestrator function that transforms canvas component to explore state
 */
export function useTransformCanvasToExploreState(
  component: BaseCanvasComponent<ComponentWithMetricsView>,
  expressionState: ExpressionState,
  timeControlState: TimeControlState,
) {
  // if (!validateUserPermissions()) {
  //   throw createLinkError(
  //     "PERMISSION_ERROR",
  //     "You do not have permission to access this explore dashboard",
  //   );
  // }

  // Get component-specific transformer properties
  const cTP = component.getExploreTransformerProperties?.();

  // Get global transformer properties from time and filter store
  const gTP = transformTimeAndFiltersToExploreState(
    expressionState,
    timeControlState,
  );

  const partialExploreState: Partial<ExploreState> = {
    ...gTP,
    ...(cTP ?? {}),
  };

  return partialExploreState;
}
