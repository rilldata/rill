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
  timeDimension: string | undefined;
}) {
  const instanceId = get(runtime).instanceId;
  const metricsView = get(ctx.metricsViewName);
  const dashboard = get(ctx.dashboardStore);
  const selectedTimeRange = get(
    ctx.selectors.timeRangeSelectors.selectedTimeRangeState,
  );
  const rows = get(ctx.selectors.pivot.rows);
  const columns = get(ctx.selectors.pivot.columns);

  const timeRange = {
    start: selectedTimeRange?.start.toISOString(),
    end: selectedTimeRange?.end.toISOString(),
  };

  const measures = columns.measure.map((m) => {
    return {
      name: m.id,
    };
  });

  const allDimensions = [...rows.dimension, ...columns.dimension].map((d) =>
    d.type === PivotChipType.Time
      ? {
          name: timeDimension,
          timeGrain: d.id as V1TimeGrain,
          timeZone: dashboard.selectedTimezone,
          alias: `Time ${d.title}`,
        }
      : {
          name: d.id,
        },
  );

  const measureFilters = await getResolvedMeasureFilters(ctx);

  const pivotOn = columns.dimension.map((d) =>
    d.type === PivotChipType.Time ? (timeDimension as string) : d.id,
  );

  const rowDimensions = [...rows.dimension].map((d) =>
    d.type === PivotChipType.Time
      ? {
          name: timeDimension,
          timeGrain: d.id as V1TimeGrain,
          timeZone: dashboard.selectedTimezone,
          alias: `Time ${d.title}`,
        }
      : {
          name: d.id,
        },
  );

  // Sort by the dimensions in the pivot's rows
  const sort = rowDimensions.map((d) => {
    if (d.alias) {
      return  {
        name: d.alias,
        desc: true,
      }
    }
    return {
      name: d.name,
      desc: true,
    };
  });

  const result = await get(query).mutateAsync({
    instanceId,
    data: {
      format,
      query: {
        metricsViewAggregationRequest: {
          instanceId,
          metricsView,
          timeRange,
          measures,
          dimensions: allDimensions,
          where: sanitiseExpression(dashboard.whereFilter, measureFilters),
          pivotOn,
          sort,
          offset: "0",
          limit: undefined, // the backend handles export limits
        },
      },
    },
  });

  const downloadUrl = `${get(runtime).host}${result.downloadUrlPath}`;

  window.open(downloadUrl, "_self");
}
