import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";

export function getUrlFromMetricsExplorer(
  metrics: MetricsExplorerEntity,
  searchParams: URLSearchParams,
  metricsView: V1MetricsViewSpec,
) {
  if (!metrics) return;

  // TODO: filter

  if (
    metrics.selectedTimeRange?.name &&
    metrics.selectedTimeRange?.name !== metricsView.defaultTimeRange
  ) {
    searchParams.set("tr", metrics.selectedTimeRange?.name);
  }
  // TODO: rest of time range

  toOverviewUrl(metrics, searchParams, metricsView);

  toTimeDimensionUrlParams(metrics, searchParams);

  toPivotUrlParams(metrics, searchParams);
}

function toOverviewUrl(
  metrics: MetricsExplorerEntity,
  searchParams: URLSearchParams,
  metricsView: V1MetricsViewSpec,
) {
  if (!metrics.allMeasuresVisible) {
    searchParams.set("e.m", [...metrics.visibleMeasureKeys].join(","));
  }
  if (!metrics.allDimensionsVisible) {
    searchParams.set("e.m", [...metrics.visibleMeasureKeys].join(","));
  }
  if (metrics.leaderboardMeasureName !== metricsView.measures?.[0]?.name) {
    searchParams.set("e.sb", metrics.leaderboardMeasureName);
  }
  if (metrics.sortDirection !== SortDirection.DESCENDING) {
    searchParams.set("e.sd", "ASC");
  }
  if (metrics.selectedDimensionName) {
    searchParams.set("e.ed", metrics.selectedDimensionName);
  }
}

function toTimeDimensionUrlParams(
  metrics: MetricsExplorerEntity,
  searchParams: URLSearchParams,
) {
  if (metrics.tdd.expandedMeasureName) {
    searchParams.set("tdd.m", metrics.tdd.expandedMeasureName);
  }
  if (metrics.tdd.pinIndex !== -1) {
    searchParams.set("tdd.p", metrics.tdd.pinIndex + "");
  }
  if (metrics.tdd.chartType !== TDDChart.DEFAULT) {
    searchParams.set("tdd.p", metrics.tdd.chartType);
  }
}

function toPivotUrlParams(
  metrics: MetricsExplorerEntity,
  searchParams: URLSearchParams,
) {
  if (!metrics.pivot.active) return;

  searchParams.set(
    "p.r",
    metrics.pivot.rows.dimension.map((d) => d.id).join(","),
  );
  searchParams.set(
    "p.c",
    [
      ...metrics.pivot.columns.dimension.map((d) => d.id),
      ...metrics.pivot.columns.measure.map((m) => m.id),
    ].join(","),
  );

  // TODO: other fields
}
