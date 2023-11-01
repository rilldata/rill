import {
  RpcStatus,
  V1MetricsViewComparisonResponse,
  V1MetricsViewSpec,
  createQueryServiceMetricsViewComparison,
} from "@rilldata/web-common/runtime-client";
import type { DashboardDataSources } from "./selectors/types";
import { activeMeasureName } from "./selectors/active-measure";
import {
  leaderboardSortedQueryBody,
  leaderboardSortedQueryOptions,
} from "./selectors/leaderboard-query";
import { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { derived, type Readable } from "svelte/store";
import type {
  CreateQueryResult,
  QueryKey,
  QueryObserverResult,
} from "@tanstack/svelte-query";
import { allDimensions } from "./selectors/dimensions";
import type { U } from "vitest/dist/types-dea83b3d";

const comparisonForDimension = (
  runtime: Runtime,
  dashData: DashboardDataSources
) => {
  const cache = new Map<string, Store>();
  return (dimensionName: string): Store => {
    return derived([ctx.metricsViewName], ([name], set) => {
      let store = cache.get(name);
      if (!store) {
        store = storeGetter(ctx);
        cache.set(name, store);
      }
      return store.subscribe(set);
    }) as Store;
  };
  let foo = createQueryServiceMetricsViewComparison(
    runtime.instanceId,
    activeMeasureName(dashData),
    leaderboardSortedQueryBody(dashData)(dimensionName)
  );

  let bar = foo.subscribe((x) => console.log(x));
};

// type queryMap = Map<string, Reada>;

type dimensionQueryLookup<Result> = (
  dimensionName: string
) => Readable<Result | undefined>;

export const createDimensionQueryLookup = (
  runtime: Runtime,
  // // dashData: DashboardDataSources
  // metricsSpecQueryResult: Readable<
  //   QueryObserverResult<V1MetricsViewSpec, RpcStatus>
  // >
  dashReadables: DashboardDataReadables
): dimensionQueryLookup<
  QueryObserverResult<V1MetricsViewComparisonResponse, RpcStatus>
> => {
  const cache = new Map<
    string,
    | (CreateQueryResult<RpcStatus, V1MetricsViewComparisonResponse> & {
        queryKey: QueryKey;
      })
    | undefined
  >();

  // return (dimensionName: string) => {
  //   let store = cache.get(dimensionName);
  //   if (store === undefined) {
  //     store = createQueryServiceMetricsViewComparison(
  //       runtime.instanceId,
  //       activeMeasureName(dashData),
  //       leaderboardSortedQueryBody(dashData)(dimensionName),
  //       leaderboardSortedQueryOptions(dashData)(dimensionName)
  //     );
  //     cache.set(dimensionName, store);
  //   }
  //   return store;
  // };

  // return (dimensionName: string) => {
  //   let derived = cache.get(dimensionName);
  //   if (derived === undefined) {
  //     const query = createQueryServiceMetricsViewComparison(
  //       runtime.instanceId,
  //       activeMeasureName(dashData),
  //       leaderboardSortedQueryBody(dashData)(dimensionName),
  //       leaderboardSortedQueryOptions(dashData)(dimensionName)
  //     );
  //   derived = derived([metricsSpecQueryResult], ([metricsSpecQueryResult], set) => )};

  return (dimensionName: string) => {
    metricsSpecQueryResult.subscribe((metricsSpecQueryResult) => {
      if (!metricsSpecQueryResult.data?.dimensions) {
        return () => undefined;
      }

      let store = cache.get(dimensionName);
      if (store === undefined) {
        store = createQueryServiceMetricsViewComparison(
          runtime.instanceId,
          activeMeasureName(dashData),
          leaderboardSortedQueryBody(dashData)(dimensionName),
          leaderboardSortedQueryOptions(dashData)(dimensionName)
        );
        cache.set(dimensionName, store);
      }
      return store;
    });
  };
};
// const cache = new Map<string, Readable>();
// return (dimensionName: string) => {
//   let store = cache.get(dimensionName);
//   if (!store) {
//     store = createQueryServiceMetricsViewComparison(
//       runtime.instanceId,
//       activeMeasureName(dashData),
//       leaderboardSortedQueryBody(dashData)(dimensionName)
//     );
//     cache.set(dimensionName, store);
//   }
//   return store;
// };
// };
