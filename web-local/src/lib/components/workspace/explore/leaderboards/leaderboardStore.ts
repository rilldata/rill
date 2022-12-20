import { get, derived, writable } from "svelte/store";
import { useRuntimeServiceMetricsViewToplist } from "@rilldata/web-common/runtime-client";
import {
  MetricsExplorerEntity,
  metricsExplorerStore,
} from "../../../../application-state-stores/explorer-stores";
import { getFilterForDimension } from "@rilldata/web-local/lib/svelte-query/dashboards";

// Store with leaderboard values in the format
// {name: dimensionName, values: [<list of values>]}
export function getLeaderboardStore(
  instanceId,
  metricViewName,
  measureName,
  dimensionColumns
) {
  const metricsExplorer: MetricsExplorerEntity =
    get(metricsExplorerStore).entities[metricViewName];

  if (!measureName && !dimensionColumns && !dimensionColumns?.length) return;
  return derived(
    dimensionColumns.map((column) => {
      const filterForDimension = getFilterForDimension(
        metricsExplorer?.filters,
        column.name
      );
      return derived(
        [
          writable(column),
          useRuntimeServiceMetricsViewToplist(instanceId, metricViewName, {
            dimensionName: column.name,
            measureNames: [measureName],
            limit: "7",
            offset: "0",
            sort: [
              {
                name: measureName,
                ascending: false,
              },
            ],
            timeStart: metricsExplorer.selectedTimeRange?.start,
            timeEnd: metricsExplorer.selectedTimeRange?.end,
            filter: filterForDimension,
          }),
        ],
        ([col, topListResult]) => {
          return {
            name: col.name,
            values: topListResult?.data?.data?.map((v) => v[measureName]),
          };
        }
      );
    }),

    (combos) => {
      return combos;
    }
  );
}

// All leaderboard values flattened into a single list
export function getAllLeaderboardValues(leaderboardStore) {
  return derived(leaderboardStore, (store) => {
    return store?.map((dimension) => dimension.values).flat();
  });
}
