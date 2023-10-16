import type { TimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type {
  V1ExportFormat,
  createQueryServiceExport,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { runtime } from "../../../runtime-client/runtime-store";
import { useDashboardStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
import { getQuerySortType } from "../leaderboard/leaderboard-utils";
import { SortDirection } from "../proto-state/derived-types";

export default async function exportToplist({
  query,
  metricViewName,
  format,
  timeControlStore,
}: {
  query: ReturnType<typeof createQueryServiceExport>;
  metricViewName: string;
  format: V1ExportFormat;
  // we need this from argument since getContext is called to get the state managers
  // which cannot run outside of component initialisation
  timeControlStore: TimeControlStore;
}) {
  const dashboardStore = useDashboardStore(metricViewName);
  const timeControlState = get(timeControlStore);

  const dashboard = get(dashboardStore);
  const result = await get(query).mutateAsync({
    instanceId: get(runtime).instanceId,
    data: {
      format,
      query: {
        metricsViewComparisonToplistRequest: {
          instanceId: get(runtime).instanceId,
          metricsViewName: metricViewName,
          dimensionName: dashboard.selectedDimensionName,
          measureNames: dashboard.selectedMeasureNames,
          baseTimeRange: {
            start: timeControlState.timeStart,
            end: timeControlState.timeEnd,
          },
          comparisonTimeRange: {
            start: timeControlState.comparisonTimeStart,
            end: timeControlState.comparisonTimeEnd,
          },
          sort: [
            {
              measureName: dashboard.leaderboardMeasureName,
              ascending: dashboard.sortDirection === SortDirection.ASCENDING,
              type: getQuerySortType(dashboard.dashboardSortType),
            },
          ],
          filter: dashboard.filters,
          limit: undefined, // the backend handles export limits
          offset: "0",
        },
      },
    },
  });

  const downloadUrl = `${get(runtime).host}${result.downloadUrlPath}`;

  window.open(downloadUrl, "_self");
}
