import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import {
  createTimeRangeSummary,
  useMetricsView,
} from "@rilldata/web-common/features/dashboards/selectors/index";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { getDefaultMetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { getUrlForPath } from "@rilldata/web-common/lib/url-utils";
import type { V1StructType } from "@rilldata/web-common/runtime-client";
import { Readable, derived, get } from "svelte/store";

export type DashboardUrlState = {
  isReady: boolean;
  proto?: string;
  defaultProto?: string;
  urlName?: string;
  urlProto?: string;
};
export type DashboardUrlStore = Readable<DashboardUrlState>;

/**
 * Creates a derived store that has the current proto and default proto of a dashboard along with the proto in the url.
 */
export function useDashboardUrlState(ctx: StateManagers): DashboardUrlStore {
  return derived(
    [
      derived(ctx.dashboardStore, (dashboard) => dashboard?.proto),
      useDashboardDefaultProto(ctx),
      page,
    ],
    ([proto, defaultProtoState, page]) => {
      if (defaultProtoState.isFetching)
        return {
          isReady: false,
        };

      const defaultProto = defaultProtoState.proto;

      let urlProto = page.url.searchParams.get("state");
      if (urlProto) urlProto = decodeURIComponent(urlProto);
      else urlProto = defaultProto;

      return {
        isReady: true,
        proto,
        defaultProto,
        urlName: getNameFromFile(page.url.pathname),
        urlProto,
      };
    },
  );
}

/**
 * Code that looks at dashboard state and url state and decides which one to sync with.
 * This depends on the fact that the same dashboard state results in the same proto and vice-versa.
 *
 * Case 1
 * 1. The dashboard state changes due to user interactions.
 * 2. Proto in the state is updated to match the new state.
 * 3. `lastKnownProto` is now different from the proto in state.
 * 4. This triggers a goto to the url with the correct proto.
 * 5. After navigation `urlProto` changes. Since this will be equal to `lastKnownProto` there will be no more operations.
 *
 * Case 2
 * 1. The url is changed by using the back button (or any other way).
 * 2. `urlProto` changes to reflect the one in the new url.
 * 3. `lastKnownProto` is now different to the `urlProto`.
 * 4. This triggers a sync of state in the url to the dashboard store.
 * 5. After updating the store proto in the state will be the same as `lastKnownProto`. No navigations happen.
 */
export function useDashboardUrlSync(ctx: StateManagers, schema: V1StructType) {
  const dashboardUrlState = useDashboardUrlState(ctx);
  const metricsView = useMetricsView(ctx);

  let lastKnownProto = get(dashboardUrlState)?.defaultProto;
  const unsub = dashboardUrlState.subscribe((state) => {
    const metricViewName = get(ctx.metricsViewName);
    if (state.urlName !== metricViewName) {
      // Edge case where the instance of sync doesnt match the active metrics view
      // TODO: We really need to rethink the new architecture that leads to this issue.
      //       Using a single constant store that changes based on name in ctx will lead to stale data.
      try {
        // Race condition when unsub is not yet initialised
        unsub();
      } catch (e) {
        // no-op
      }
      return;
    }

    if (!state.isReady || !state.proto) return;

    if (state.proto !== lastKnownProto) {
      // changed when filters etc are changed on the dashboard
      gotoNewDashboardUrl(get(page).url, state.proto, state.defaultProto);

      lastKnownProto = state.proto;
    } else if (state.urlProto !== lastKnownProto) {
      // changed when user updated the url manually
      metricsExplorerStore.syncFromUrl(
        metricViewName,
        state.urlProto,
        get(metricsView).data,
        schema,
      );
      lastKnownProto = state.urlProto;
    }
  });

  return unsub;
}

function gotoNewDashboardUrl(url: URL, newState: string, defaultState: string) {
  // this store the actual state. for default state it will be empty
  let newStateInUrl = "";
  // changed when filters etc are changed on the dashboard

  const newUrl = getUrlForPath(url.pathname, ["features", "theme"]);

  if (newState !== defaultState) {
    newStateInUrl = encodeURIComponent(newState);
    newUrl.searchParams.set("state", newStateInUrl);
  }

  const currentStateInUrl = url.searchParams.get("state") ?? "";

  if (newStateInUrl === currentStateInUrl) return;
  goto(newUrl.toString());
}

// NOTE: the data here can be stale when metricsViewName changes in ctx, along with the metricsView.
//       but the time range summary is yet to be triggered to change causing it to have data from previous active dashboard
// Above issue is currently fixed in useDashboardUrlSync and DashboardURLStateProvider by create a new instance when the url is changed.
// TODO: we need to update the architecture to perhaps recreate all derived stores when the metricsViewName changes
export function useDashboardDefaultProto(ctx: StateManagers) {
  return derived(
    [useMetricsView(ctx), createTimeRangeSummary(ctx)],
    ([metricsView, timeRangeSummary]) => {
      const hasTimeSeries = Boolean(metricsView.data?.timeDimension);
      if (!metricsView.data || (hasTimeSeries && !timeRangeSummary.data))
        return {
          isFetching: true,
          proto: "",
        };

      const metricsExplorer = getDefaultMetricsExplorerEntity(
        get(ctx.metricsViewName),
        metricsView.data,
        timeRangeSummary.data,
      );
      return {
        isFetching: false,
        proto: getProtoFromDashboardState(metricsExplorer),
      };
    },
  );
}
