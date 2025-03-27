import { page } from "$app/stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  convertExploreStateToURLSearchParams,
  convertExploreStateToURLSearchParamsNoCompression,
} from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import {
  type DashboardTimeControls,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

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

    // TODO: if all params are equal to dashboard defaults then should we skip it?
    return convertURLSearchParamsToExploreState(
      new URLSearchParams(rawUrlSearch),
      metricsView,
      explore,
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
    // TODO: find a way to avoid using any
    (newExploreState as any)[k] = exploreState[k];
  });

  // Custom handling for time range. We are retaining the previous range if the current range is a custom range.
  if (exploreState.selectedTimeRange?.name) {
    if (exploreState.selectedTimeRange.name === TimeRangePreset.CUSTOM) {
      newExploreState.selectedTimeRange =
        existingExploreState?.selectedTimeRange;
    } else {
      newExploreState.selectedTimeRange = {
        name: exploreState.selectedTimeRange.name,
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
    localStorage.setItem(
      getKeyForLocalStore(exploreName, prefix),
      urlSearchParams.toString(),
    );
  } catch {
    // no-op
  }
}
