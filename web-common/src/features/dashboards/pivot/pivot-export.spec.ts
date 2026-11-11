import { getPivotAggregationRequest } from "@rilldata/web-common/features/dashboards/pivot/pivot-export";
import {
  type PivotChipData,
  PivotChipType,
  type PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { SortingState } from "tanstack-table-8-svelte-5";
import { describe, expect, it } from "vitest";

const chip = (type: PivotChipType, id: string): PivotChipData => ({
  id,
  title: id,
  type,
});

describe("getPivotAggregationRequest", () => {
  function flatRequestSort(
    sorting: SortingState,
    enableComparison = false,
    columns: PivotState["columns"] = [
      chip(PivotChipType.Dimension, "domain"),
      chip(PivotChipType.Time, V1TimeGrain.TIME_GRAIN_DAY),
      chip(PivotChipType.Measure, "impressions"),
    ],
  ) {
    const pivotState: PivotState = {
      rows: [],
      columns: [],
      sorting,
      expanded: {},
      columnPage: 1,
      rowPage: 1,
      enableComparison: true,
      tableMode: "flat",
      activeCell: null,
      showTotalsColumn: true,
      showTotalsRow: true,
    };
    return getPivotAggregationRequest({
      instanceId: "default",
      metricsViewName: "AdBids_metrics",
      timeDimension: "timestamp",
      exploreState: {
        selectedTimezone: "UTC",
        whereFilter: createAndExpression([]),
      } as ExploreState,
      timeRange: { start: "2024-01-01T00:00:00Z", end: "2024-02-01T00:00:00Z" },
      rows: [],
      columns: {
        dimension: columns.filter((c) => c.type !== PivotChipType.Measure),
        measure: columns.filter((c) => c.type === PivotChipType.Measure),
      },
      enableComparison,
      comparisonTime: enableComparison
        ? { start: "2023-12-01T00:00:00Z", end: "2024-01-01T00:00:00Z" }
        : undefined,
      isFlat: true,
      pivotState,
    })?.sort;
  }

  it("sorts a flat export by a dimension or measure column", () => {
    expect(flatRequestSort([{ id: "domain", desc: false }])).toEqual([
      { name: "domain", desc: false },
    ]);
    expect(
      flatRequestSort([{ id: "timestamp_rill_TIME_GRAIN_DAY", desc: true }]),
    ).toEqual([{ name: "timestamp_rill_TIME_GRAIN_DAY", desc: true }]);
    expect(flatRequestSort([{ id: "impressions", desc: false }])).toEqual([
      { name: "impressions", desc: false },
    ]);
  });

  it("sorts by a comparison column only while comparing", () => {
    expect(
      flatRequestSort([{ id: "impressions__delta_abs", desc: true }], true),
    ).toEqual([{ name: "impressions__delta_abs", desc: true }]);
    // Without a comparison the request carries no comparison measures, so the
    // sort falls back to the default instead of naming a field the runtime
    // cannot resolve.
    expect(
      flatRequestSort([{ id: "impressions__delta_abs", desc: true }], false),
    ).toEqual([{ name: "impressions", desc: true }]);
  });

  it("falls back to the default sort for a field that is not in the request", () => {
    expect(flatRequestSort([{ id: "bid_price", desc: true }])).toEqual([
      { name: "impressions", desc: true },
    ]);
    expect(flatRequestSort([])).toEqual([{ name: "impressions", desc: true }]);
  });

  it("falls back to the time column's request alias when there are no measures", () => {
    const timeOnly = [chip(PivotChipType.Time, V1TimeGrain.TIME_GRAIN_DAY)];
    expect(
      flatRequestSort([{ id: "bid_price", desc: true }], false, timeOnly),
    ).toEqual([{ name: "timestamp_rill_TIME_GRAIN_DAY", desc: false }]);
  });
});
