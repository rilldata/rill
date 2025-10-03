<script lang="ts">
  import { ChartContainer } from "@rilldata/web-common/features/components/charts";
  import type { CartesianChartSpec } from "@rilldata/web-common/features/components/charts/cartesian/CartesianChartProvider";
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { readable, type Readable } from "svelte/store";

  const timeAndFilterStore: Readable<TimeAndFilterStore> = readable({
    timeRange: {
      start: "2024-11-03T20:00:00.000Z",
      end: "2024-11-04T20:00:00.000Z",
      timeZone: "UTC",
    },
    comparisonTimeRange: undefined,
    showTimeComparison: false,
    where: {
      cond: {
        op: "OPERATION_AND",
        exprs: [],
      },
    },
    timeGrain: "TIME_GRAIN_HOUR",
    timeRangeState: undefined,
    comparisonTimeRangeState: undefined,
    hasTimeSeries: true,
  });

  const spec: Readable<CartesianChartSpec> = readable({
    metrics_view: "traffic_metrics",
    color: {
      field: "Country",
      type: "nominal",
    },
    x: {
      field: "UpdateTimeUTC",
      limit: 20,
      showNull: true,
      type: "temporal",
    },
    y: {
      field: "total_jams_count_measure",
      type: "quantitative",
      zeroBasedOrigin: true,
    },
  });
</script>

<ChartContainer chartType="stacked_bar" {spec} {timeAndFilterStore} />
