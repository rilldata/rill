import { getDashboardFromEmbedRoute } from "@rilldata/web-admin/features/embeds/embed-route-utils.ts";
import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import {
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { derived } from "svelte/store";

/**
 * Returns the explore that alerts can be created on for an embed route, or an empty string if alerts can't be created there.
 * Alerts are created on an explore, so they can only be created from explore pages.
 */
export function getEmbedAlertExploreName(
  routeId: string | null,
  params: Record<string, string> | undefined,
): string {
  const dashboard = getDashboardFromEmbedRoute(routeId, params);
  return dashboard?.kind === ResourceKind.Explore ? dashboard.name : "";
}

/**
 * Returns whether an alert can be created on the explore, which requires a time dimension.
 * The embed header is outside of the explore's state managers, so the explore's spec and state are read by name.
 */
export function getEmbedCreateAlertState(
  runtimeClient: RuntimeClient,
  exploreName: string,
) {
  return derived(
    [
      useExploreValidSpec(runtimeClient, exploreName),
      useExploreState(exploreName),
    ],
    ([exploreSpec, exploreState]) => ({
      canCreate:
        !!exploreName &&
        !!exploreSpec.data?.metricsView?.timeDimension &&
        !!exploreState,
      // Same as the guard of the dashboard's create alert button (see CreateAlert).
      isCustomTimeRange:
        exploreState?.selectedTimeRange?.name === TimeRangePreset.CUSTOM ||
        exploreState?.selectedComparisonTimeRange?.name ===
          TimeComparisonOption.CUSTOM,
    }),
  );
}
