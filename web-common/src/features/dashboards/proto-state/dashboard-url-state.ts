import { goto } from "$app/navigation";
import { page } from "$app/stores";
import {
  metricsExplorerStore,
  useDashboardStore,
} from "@rilldata/web-common/features/dashboards/dashboard-stores";
import type { V1MetricsView } from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived, get, Readable } from "svelte/store";

export type DashboardUrlState = {
  proto: string;
  defaultProto: string;
  urlProto: string;
};
export type DashboardUrlStore = Readable<DashboardUrlState>;

/**
 * Creates a derived store that has the current proto and default proto of a dashboard along with the proto in the url.
 */
export function useDashboardUrlState(
  metricViewName: string
): DashboardUrlStore {
  return derived(
    [
      useDashboardProto(metricViewName),
      useDashboardDefaultProto(metricViewName),
      page,
    ],
    ([proto, defaultProto, page]) => {
      let urlProto = page.url.searchParams.get("state");
      if (urlProto) urlProto = decodeURIComponent(urlProto);
      else urlProto = defaultProto;

      return {
        proto,
        defaultProto,
        urlProto,
      };
    }
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
export function useDashboardUrlSync(
  metricViewName: string,
  metaQuery: CreateQueryResult<V1MetricsView>
) {
  const dashboardUrlState = useDashboardUrlState(metricViewName);
  let lastKnownProto: string;
  return dashboardUrlState.subscribe((state) => {
    if (state.proto !== lastKnownProto) {
      // changed when filters etc are changed on the dashboard
      gotoNewDashboardUrl(get(page).url, state.proto, state.defaultProto);

      lastKnownProto = state.proto;
    } else if (state.urlProto !== lastKnownProto) {
      // changed when user updated the url manually
      metricsExplorerStore.syncFromUrl(
        metricViewName,
        state.urlProto,
        get(metaQuery).data
      );
      lastKnownProto = state.urlProto;
    }
  });
}

function gotoNewDashboardUrl(url: URL, newState: string, defaultState: string) {
  // this store the actual state. for default state it will be empty
  let newStateInUrl = "";
  // changed when filters etc are changed on the dashboard
  let newPath = url.pathname;
  if (newState !== defaultState) {
    newStateInUrl = encodeURIComponent(newState);
    newPath = `${newPath}?state=${newStateInUrl}`;
  }

  const currentStateInUrl = url.searchParams.get("state") ?? "";

  if (newStateInUrl === currentStateInUrl) return;
  goto(newPath);
}

// TODO: if these are necessary anywhere else move them to a separate file
// memoization of dashboard proto
export function useDashboardProto(name: string) {
  return derived(useDashboardStore(name), (dashboard) => dashboard?.proto);
}
export function useDashboardDefaultProto(name: string) {
  return derived(
    useDashboardStore(name),
    (dashboard) => dashboard?.defaultProto
  );
}
