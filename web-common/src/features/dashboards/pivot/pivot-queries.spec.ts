import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  type V1MetricsViewAggregationRequest,
  createQueryServiceMetricsViewAggregation,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { QueryClient } from "@tanstack/svelte-query";
import { get, readable } from "svelte/store";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  getAxisForDimensions,
  getAxisQueryForMeasureTotals,
} from "./pivot-queries";
import type { PivotDashboardContext, PivotDataStoreConfig } from "./types";

vi.mock("@rilldata/web-common/runtime-client", async (importOriginal) => {
  const actual =
    await importOriginal<
      typeof import("@rilldata/web-common/runtime-client")
    >();
  return {
    ...actual,
    createQueryServiceMetricsViewAggregation: vi.fn(() =>
      readable({ isFetching: false, data: { data: [] } }),
    ),
  };
});

const ctx: PivotDashboardContext = {
  runtimeClient: {} as RuntimeClient,
  metricsViewName: readable("metrics_view"),
  queryClient: {} as QueryClient,
  enabled: readable(true),
};

// `country` renders as a link, `measure` is the only measure.
const config: PivotDataStoreConfig = {
  measureNames: ["measure"],
  rowDimensionNames: ["country", "city"],
  colDimensionNames: [],
  allMeasures: [],
  allDimensions: [{ name: "country", uri: "https://example.com/{{ .value }}" }],
  whereFilter: createAndExpression([]),
  pivot: {
    activeCell: null,
    columnPage: 1,
    columns: [],
    enableComparison: false,
    expanded: {},
    rowPage: 1,
    rows: [],
    sorting: [{ id: "measure", desc: true }],
    tableMode: "nest",
    showTotalsColumn: true,
    showTotalsRow: true,
  },
  time: {
    timeStart: undefined,
    timeEnd: undefined,
    timeZone: "UTC",
    timeDimension: "",
  },
  comparisonTime: undefined,
  enableComparison: false,
  searchText: undefined,
  isFlat: false,
};

const uriMeasure = { name: "country__rill_uri", uri: { dimension: "country" } };

function requestedMeasures() {
  return vi
    .mocked(createQueryServiceMetricsViewAggregation)
    .mock.calls.map(
      (call) => (call[1] as V1MetricsViewAggregationRequest).measures,
    );
}

describe("getAxisForDimensions", () => {
  beforeEach(() => {
    vi.mocked(createQueryServiceMetricsViewAggregation).mockClear();
  });

  it("appends the URI measure of a row dimension that renders as a link", () => {
    get(
      getAxisForDimensions(
        ctx,
        config,
        ["country"],
        [{ name: "measure" }],
        createAndExpression([]),
        [],
        undefined,
        "100",
        "0",
        true,
      ),
    );

    expect(requestedMeasures()).toEqual([[{ name: "measure" }, uriMeasure]]);
  });

  it("keeps a sort-only query to its single filtered measure", () => {
    // Expanding a row sorted by a measure filters that measure by the parent
    // row's value. The runtime rejects a measure-level filter alongside any
    // other measure, so the URI measure must not ride along.
    const sortOnlyMeasures = [
      { name: "measure", filter: createInExpression("city", ["Paris"]) },
    ];

    get(
      getAxisForDimensions(
        ctx,
        config,
        ["country"],
        sortOnlyMeasures,
        createAndExpression([]),
        [{ name: "measure", desc: true }],
        undefined,
        "101",
        "0",
        true,
      ),
    );

    expect(requestedMeasures()).toEqual([sortOnlyMeasures]);
  });
});

describe("getAxisQueryForMeasureTotals", () => {
  beforeEach(() => {
    vi.mocked(createQueryServiceMetricsViewAggregation).mockClear();
  });

  it("requests the URI measure with the totals shown for measure-sorted rows", () => {
    get(
      getAxisQueryForMeasureTotals(
        ctx,
        config,
        true,
        "measure",
        "country",
        ["France", "Spain"],
        { start: undefined, end: undefined },
      ),
    );

    expect(requestedMeasures()).toEqual([[{ name: "measure" }, uriMeasure]]);
  });
});
