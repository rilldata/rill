import { getDimensionTableAggregationRequestForTime } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-export-utils";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import {
  V1ExportFormat,
  createQueryServiceExport,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { runtime } from "../../../runtime-client/runtime-store";

export default async function exportToplist({
  ctx,
  query,
  format,
  searchText,
}: {
  ctx: StateManagers;
  query: ReturnType<typeof createQueryServiceExport>;
  format: V1ExportFormat;
  searchText: string;
}) {
  const metricsViewName = get(ctx.metricsViewName);
  const dashboard = get(ctx.dashboardStore);
  const timeControlState = get(
    ctx.selectors.timeRangeSelectors.timeControlsState,
  );

  const result = await get(query).mutateAsync({
    instanceId: get(runtime).instanceId,
    data: {
      format,
      query: {
        metricsViewAggregationRequest:
          getDimensionTableAggregationRequestForTime(
            metricsViewName,
            dashboard,
            {
              start: timeControlState.timeStart,
              end: timeControlState.timeEnd,
            },
            {
              start: timeControlState.comparisonTimeStart,
              end: timeControlState.comparisonTimeEnd,
            },
            searchText,
          ),
      },
    },
  });

  const downloadUrl = `${get(runtime).host}${result.downloadUrlPath}`;

  window.open(downloadUrl, "_self");
}
