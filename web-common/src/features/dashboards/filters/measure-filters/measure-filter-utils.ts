import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import {
  createAndExpression,
  createInExpression,
  getAllIdentifiers,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import {
  createQueryServiceMetricsViewToplist,
  V1Expression,
  type V1MetricsViewSort,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { derived, get, Readable } from "svelte/store";

export type ResolvedMeasureFilter = {
  ready: boolean;
  filter: V1Expression | undefined;
};

/**
 * We run the measure filters on the frontend for now.
 * The resulting dimension values for those are then converted to an `in` query.
 */
export function prepareMeasureFilterResolutions(
  dashboard: MetricsExplorerEntity,
  timeControls: TimeControlState,
  queryClient: QueryClient,
): Readable<ResolvedMeasureFilter> {
  return derived(
    dashboard.dimensionThresholdFilters.map((dtf) => {
      const measureNames = getAllIdentifiers(dtf.filter);
      const sort: V1MetricsViewSort[] = measureNames.map((m) => ({
        name: m,
        ascending: false,
      }));
      sort.forEach((s) => {
        if (s.name === dashboard.leaderboardMeasureName) {
          // retain the sort order for selected measure
          s.ascending = dashboard.sortDirection === SortDirection.ASCENDING;
        }
      });
      return createQueryServiceMetricsViewToplist(
        get(runtime).instanceId,
        dashboard.name,
        {
          dimensionName: dtf.name,
          measureNames,
          having: dtf.filter,
          timeStart: timeControls.timeStart,
          timeEnd: timeControls.timeEnd,
          limit: "50",
          offset: "0",
          sort,
        },
        {
          query: {
            enabled: !!dtf.filter.cond?.exprs?.length,
            queryClient,
          },
        },
      );
    }),
    (toplists) => {
      if (toplists.some((t) => t.isFetching) || toplists.some((t) => !t.data)) {
        return {
          ready: false,
          filter: undefined,
        };
      }

      if (toplists.length === 0) {
        return {
          ready: true,
          filter: undefined,
        };
      }

      if (toplists.some((t) => !t.data?.data?.length)) {
        // if there is some toplist data not returning any result
        // then add a `false` in the condition to not match any rows
        return {
          ready: true,
          filter: {
            val: false,
          },
        };
      }

      // This makes sure dashboard.dimensionThresholdFilters and toplist is in sync
      if (
        toplists.length !== dashboard.dimensionThresholdFilters.length ||
        toplists.some(
          (t, i) =>
            (t.data?.meta?.findIndex(
              (c) => c.name === dashboard.dimensionThresholdFilters[i].name,
            ) ?? -1) === -1,
        )
      ) {
        return {
          ready: false,
          filter: undefined,
        };
      }

      const inFilters = toplists.map((t, i) =>
        // create an in expression for each dimension in the filters
        createInExpression(
          dashboard.dimensionThresholdFilters[i].name,
          // add the values from the toplist response within the "in expression"
          t.data?.data?.map(
            (d) => d[dashboard.dimensionThresholdFilters[i].name],
          ) ?? [],
        ),
      );
      if (inFilters.length === 0) {
        return {
          ready: true,
          filter: undefined,
        };
      }

      return {
        ready: true,
        filter: createAndExpression(inFilters),
      };
    },
  );
}

export function measureFilterResolutionsStore(
  ctx: StateManagers,
): Readable<ResolvedMeasureFilter> {
  return derived(
    [ctx.dashboardStore, useTimeControlStore(ctx)],
    ([dashboard, timeControlState], set) => {
      prepareMeasureFilterResolutions(
        dashboard,
        timeControlState,
        ctx.queryClient,
      ).subscribe(set);
    },
  );
}

export async function getResolvedMeasureFilters(ctx: StateManagers) {
  const measureFiltersStore = measureFilterResolutionsStore(ctx);
  await waitUntil(() => get(measureFiltersStore).ready);
  return get(measureFiltersStore).filter;
}
