import { getResolvedMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  V1ExportFormat,
  V1TimeGrain,
  createQueryServiceExport,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { runtime } from "../../../runtime-client/runtime-store";
import type { StateManagers } from "../state-managers/state-managers";
import { PivotChipType } from "./types";

export default async function exportPivot({
  ctx,
  query,
  format,
  timeDimension,
}: {
  ctx: StateManagers;
  query: ReturnType<typeof createQueryServiceExport>;
  format: V1ExportFormat;
  timeDimension: string;
}) {
  const metricsView = get(ctx.metricsViewName);
  const dashboard = get(ctx.dashboardStore);
  const selectedTimeRange = get(
    ctx.selectors.timeRangeSelectors.selectedTimeRangeState,
  );

  // used for dimensions
  const rows = get(ctx.selectors.pivot.rows);
  // used for pivot on
  const columns = get(ctx.selectors.pivot.columns);

  const measureFilters = await getResolvedMeasureFilters(ctx);

  const result = await get(query).mutateAsync({
    instanceId: get(runtime).instanceId,
    data: {
      format,
      query: {
        metricsViewAggregationRequest: {
          // Q: should I use the `pivot.dimensions` selector directly?
          dimensions: rows.dimension.map((d) =>
            d.type === PivotChipType.Time
              ? {
                  name: timeDimension,
                  timeGrain: d.id as V1TimeGrain,
                  timeZone: dashboard.selectedTimezone,
                }
              : {
                  name: d.id,
                },
          ),
          where: sanitiseExpression(dashboard.whereFilter, measureFilters),
          instanceId: get(runtime).instanceId,
          limit: undefined, // the backend handles export limits
          // Q: should I use the `pivot.measures` selector directly?
          measures: columns.measure.map((m) => {
            return {
              name: m.id,
            };
          }),
          metricsView,
          offset: "0",
          pivotOn: columns.dimension.map((d) => d.id),
          sort: undefined, // future work
          timeRange: {
            start: selectedTimeRange?.start.toISOString(),
            end: selectedTimeRange?.end.toISOString(),
          },
        },
      },
    },
  });

  const downloadUrl = `${get(runtime).host}${result.downloadUrlPath}`;

  window.open(downloadUrl, "_self");
}
