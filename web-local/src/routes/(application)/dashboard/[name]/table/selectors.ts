import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { createQueryServiceMetricsViewRows } from "@rilldata/web-common/runtime-client";
import { derived } from "svelte/store";

export function getTableData(managers: StateManagers) {
  // const metricViewName = get(managers.metricsViewName);
  // const runtime =
  return derived(
    [managers.dashboardStore, managers.metricsViewName, managers.runtime],
    ([dashboardStore, metricsViewName, runtime], set) => {
      const query = createQueryServiceMetricsViewRows(
        runtime.instanceId,
        metricsViewName,
        {
          limit: 100,
          //   filter: dashboardStore.filters,
          //   timeStart: hasTimeSeries ? timeStart : undefined,
          //   timeEnd: hasTimeSeries ? timeEnd : undefined,
        },
        {
          query: {
            enabled: true,
            queryClient: managers.queryClient,
            //   (hasTimeSeries ? !!timeStart && !!timeEnd : true) &&
            //   !!$dashboardStore?.filters,
          },
        }
      );

      return query.subscribe((q) => {
        if (q.data) set(q.data);
        else
          set({
            data: [],
            meta: [],
          });
      });
    }
  );
}
