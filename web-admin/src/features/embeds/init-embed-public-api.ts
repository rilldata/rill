import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import { derived, get, type Readable } from "svelte/store";
import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import {
  getTimeControlState,
  type TimeControlState,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  emitNotification,
  registerRPCMethod,
} from "@rilldata/web-common/lib/rpc";

export default function initEmbedPublicAPI(): () => void {
  const { validSpecStore, dashboardStore, timeRangeSummaryStore } =
    getStateManagers();

  const derivedState: Readable<string> = derived(
    [validSpecStore, dashboardStore, timeRangeSummaryStore],
    ([$validSpecStore, $dashboardStore, $timeRangeSummaryStore]) => {
      const exploreSpec = $validSpecStore.data?.explore ?? {};
      const metricsViewSpec = $validSpecStore.data?.metricsView ?? {};

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
        convertPartialExploreStateToUrlParams(
          exploreSpec,
          $dashboardStore,
          timeControlsState,
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

    const exploreSpec = validSpec.data?.explore ?? {};
    const metricsViewSpec = validSpec.data?.metricsView ?? {};

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
      convertPartialExploreStateToUrlParams(
        exploreSpec,
        dashboard,
        timeControlsState,
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
