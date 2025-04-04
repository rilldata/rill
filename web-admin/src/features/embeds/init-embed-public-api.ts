import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { derived, get, type Readable } from "svelte/store";

import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { convertExploreStateToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";

import {
  getTimeControlState,
  type TimeControlState,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  emitNotification,
  registerRPCMethod,
} from "@rilldata/web-common/lib/rpc";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export default function initEmbedPublicAPI(instanceId: string): () => void {
  const {
    metricsViewName,
    validSpecStore,
    dashboardStore,
    timeRangeSummaryStore,
  } = getStateManagers();

  const metricsViewNameValue = get(metricsViewName);
  const metricsViewTimeRange = useMetricsViewTimeRange(
    instanceId,
    metricsViewNameValue,
    {},
    queryClient,
  );

  const derivedState: Readable<string> = derived(
    [
      validSpecStore,
      dashboardStore,
      timeRangeSummaryStore,
      metricsViewTimeRange,
    ],
    ([
      $validSpecStore,
      $dashboardStore,
      $timeRangeSummaryStore,
      $metricsViewTimeRange,
    ]) => {
      const exploreSpec = $validSpecStore.data?.explore ?? {};
      const metricsViewSpec = $validSpecStore.data?.metricsView ?? {};

      const defaultExplorePreset = getDefaultExplorePreset(
        exploreSpec,
        metricsViewSpec,
        $metricsViewTimeRange?.data,
      );

      let timeControlsState: TimeControlState | undefined = undefined;
      if (metricsViewSpec && exploreSpec && $dashboardStore) {
        timeControlsState = getTimeControlState(
          metricsViewSpec,
          exploreSpec,
          $timeRangeSummaryStore.data?.timeRangeSummary,
          $dashboardStore,
        );
      }

      return decodeURIComponent(
        convertExploreStateToURLSearchParams(
          $dashboardStore,
          exploreSpec,
          timeControlsState,
          defaultExplorePreset,
        ).toString(),
      );
    },
  );

  const unsubscribe = derivedState.subscribe((stateString) => {
    emitNotification("stateChange", { state: stateString });
  });

  registerRPCMethod("getState", () => {
    const validSpec = get(validSpecStore);
    const dashboard = get(dashboardStore);
    const timeSummary = get(timeRangeSummaryStore).data;
    const metricsTime = get(metricsViewTimeRange);

    const exploreSpec = validSpec.data?.explore ?? {};
    const metricsViewSpec = validSpec.data?.metricsView ?? {};

    const defaultExplorePreset = getDefaultExplorePreset(
      exploreSpec,
      metricsViewSpec,
      metricsTime?.data,
    );

    let timeControlsState: TimeControlState | undefined = undefined;
    if (metricsViewSpec && exploreSpec && dashboard) {
      timeControlsState = getTimeControlState(
        metricsViewSpec,
        exploreSpec,
        timeSummary?.timeRangeSummary,
        dashboard,
      );
    }
    const stateString = decodeURIComponent(
      convertExploreStateToURLSearchParams(
        dashboard,
        exploreSpec,
        timeControlsState,
        defaultExplorePreset,
      ).toString(),
    );
    return { state: stateString };
  });

  registerRPCMethod("setState", (state: string) => {
    if (typeof state !== "string") {
      return new Error("Expected state to be a string");
    }
    const currentUrl = new URL(get(page).url);
    currentUrl.search = state;
    void goto(currentUrl, { replaceState: true });
    return true;
  });

  emitNotification("ready");

  return unsubscribe;
}
