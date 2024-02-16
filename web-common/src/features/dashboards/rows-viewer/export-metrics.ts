import { getResolvedMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { get } from "svelte/store";
import { runtime } from "../../../runtime-client/runtime-store";
import type {
  V1ExportFormat,
  createQueryServiceExport,
} from "@rilldata/web-common/runtime-client";

export default async function exportMetrics({
  ctx,
  query,
  format,
}: {
  ctx: StateManagers;
  query: ReturnType<typeof createQueryServiceExport>;
  format: V1ExportFormat;
}) {
  const metricsViewName = get(ctx.metricsViewName);
  const dashboard = get(ctx.dashboardStore);
  const timeControlState = get(
    ctx.selectors.timeRangeSelectors.timeControlsState,
  );
  const measureFilters = await getResolvedMeasureFilters(ctx);

  const result = await get(query).mutateAsync({
    instanceId: get(runtime).instanceId,
    data: {
      format,
      query: {
        metricsViewRowsRequest: {
          instanceId: get(runtime).instanceId,
          metricsViewName,
          where: sanitiseExpression(dashboard.whereFilter, measureFilters),
          timeStart: timeControlState.timeStart,
          timeEnd: timeControlState.timeEnd,
        },
      },
    },
  });
  const downloadUrl = `${get(runtime).host}${result.downloadUrlPath}`;

  window.open(downloadUrl, "_self");
}
