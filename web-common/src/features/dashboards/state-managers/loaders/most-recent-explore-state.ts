import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertExploreStateToURLSearchParamsNoCompression } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import {
  type DashboardTimeControls,
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

// Keys that do not need any special handling and can be directly copied over
const DirectCopyExploreStateKeys: (keyof MetricsExplorerEntity)[] = [
  "activePage",
  "showTimeComparison",
  "allMeasuresVisible",
  "visibleMeasureKeys",
  "allDimensionsVisible",
  "visibleDimensionKeys",
  "leaderboardMeasureName",
  "sortDirection",
  "leaderboardContextColumn",
];

function getKeyForLocalStore(exploreName: string, prefix: string | undefined) {
  return `rill:app:explore:${prefix ?? ""}${exploreName}`.toLowerCase();
}

export function getMostRecentExploreState(
  exploreName: string,
  prefix: string | undefined,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
) {
  try {
    const rawUrlSearch = localStorage.getItem(
      getKeyForLocalStore(exploreName, prefix),
    );
    if (!rawUrlSearch) return { partialExploreState: undefined, errors: [] };

    console.log("LOCAL:LOAD", rawUrlSearch);
    return convertURLSearchParamsToExploreState(
      new URLSearchParams(rawUrlSearch),
      metricsView,
      explore,
      // Send empty preset so that fields are always stored.
      {},
    );
  } catch {
    // no-op
  }
  return { partialExploreState: undefined, errors: [] };
}

export function saveMostRecentExploreState(
  exploreName: string,
  prefix: string | undefined,
  metricsView: V1MetricsViewSpec,
  explore: V1ExploreSpec,
  timeControlsState: TimeControlState | undefined,
  exploreState: MetricsExplorerEntity,
) {
  // TODO: save relevant fields on non-default pages
  if (exploreState.activePage !== DashboardState_ActivePage.DEFAULT) return;

  const { partialExploreState: existingExploreState } =
    getMostRecentExploreState(exploreName, prefix, metricsView, explore) ?? {};
  const newExploreState: Partial<MetricsExplorerEntity> = {};

  DirectCopyExploreStateKeys.forEach((k) => {
    (newExploreState as any)[k] = exploreState[k];
  });

  // Since we are storing a few settings in timeControlsState, url params is populated using it.
  // Hopefully we will store everything in a single place in the future and we can update timeControlsState directly.
  if (timeControlsState) {
    // Custom handling for time range. We are retaining the previous range if the current range is a custom range.
    if (exploreState.selectedTimeRange?.name) {
      if (exploreState.selectedTimeRange.name === TimeRangePreset.CUSTOM) {
        timeControlsState.selectedTimeRange =
          existingExploreState?.selectedTimeRange;
      } else {
        timeControlsState.selectedTimeRange = {
          name: exploreState.selectedTimeRange.name,
        } as DashboardTimeControls;
      }
    }

    // Reset the comparison time range to default. We are only saving whether it is enabled and not the actual range.
    if (timeControlsState.selectedComparisonTimeRange) {
      timeControlsState.selectedComparisonTimeRange = {
        name: TimeComparisonOption.CONTIGUOUS,
      } as DashboardTimeControls;
    }
  }

  const urlSearchParams = convertExploreStateToURLSearchParamsNoCompression(
    newExploreState as MetricsExplorerEntity,
    explore,
    timeControlsState,
    {},
  );

  try {
    console.log("LOCAL:SAVE", urlSearchParams.toString());
    localStorage.setItem(
      getKeyForLocalStore(exploreName, prefix),
      urlSearchParams.toString(),
    );
  } catch {
    // no-op
  }
}

export function clearMostRecentExploreState(
  exploreName: string,
  prefix: string | undefined,
) {
  const key = getKeyForLocalStore(exploreName, prefix);
  localStorage.removeItem(key);
}
