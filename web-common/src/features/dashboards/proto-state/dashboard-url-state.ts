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
 */
export function useDashboardUrlSync(
  metricViewName: string,
  metaQuery: CreateQueryResult<V1MetricsView>
) {
  const dashboardUrlState = useDashboardUrlState(metricViewName);
  let desiredProto: string;
  return dashboardUrlState.subscribe((state) => {
    if (state.proto !== desiredProto) {
      // changed when filters etc are changed on the dashboard
      const pathName = get(page).url.pathname;
      if (state.proto === state.defaultProto) {
        goto(`${pathName}`);
      } else {
        goto(`${pathName}?state=${encodeURIComponent(state.proto)}`);
      }
      // set the desired proto to the one from state
      desiredProto = state.proto;
    } else if (state.urlProto !== desiredProto) {
      // changed when user updated the url manually
      metricsExplorerStore.syncFromUrl(
        metricViewName,
        state.urlProto,
        get(metaQuery).data
      );
      // set desired proto to the one from url
      desiredProto = state.urlProto;
    }
  });
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
