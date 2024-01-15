import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import {
  createAndExpression,
  createInExpression,
  getAllIdentifiers,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  createQueryServiceMetricsViewToplist,
  V1Expression,
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
    dashboard.dimensionThresholdFilters.map((dtf) =>
      createQueryServiceMetricsViewToplist(
        get(runtime).instanceId,
        dashboard.name,
        {
          dimensionName: dtf.name,
          measureNames: getAllIdentifiers(dtf.filter),
          having: dtf.filter,
          timeStart: timeControls.timeStart,
          timeEnd: timeControls.timeEnd,
          limit: "50",
          offset: "0",
          sort: [],
        },
        {
          query: {
            enabled: !!dtf.filter.cond?.exprs?.length,
            queryClient,
          },
        },
      ),
    ),
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

      const filter = createAndExpression(
        toplists.map((t, i) =>
          // create an in expression for each dimension in the filters
          createInExpression(
            dashboard.dimensionThresholdFilters[i].name,
            // add the values from the toplist response within the "in expression"
            t.data?.data?.map(
              (d) => d[dashboard.dimensionThresholdFilters[i].name],
            ) ?? [],
          ),
        ),
      );

      return {
        ready: true,
        filter,
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
