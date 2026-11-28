import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  type V1Expression,
  type V1MetricsViewAggregationRequest,
  V1Operation,
  createQueryServiceMetricsViewAggregation,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { QueryClient } from "@tanstack/svelte-query";
import { get, readable, writable } from "svelte/store";
import { describe, expect, it, vi } from "vitest";
import { createPivotDataStore } from "./pivot-data-store";
import { buildExpandKey } from "./pivot-expand-keys";
import type {
  PivotDashboardContext,
  PivotDataState,
  PivotDataStoreConfig,
} from "./types";

vi.mock("@rilldata/web-common/runtime-client", async (importOriginal) => {
  const actual =
    await importOriginal<
      typeof import("@rilldata/web-common/runtime-client")
    >();
  return { ...actual, createQueryServiceMetricsViewAggregation: vi.fn() };
});

const ctx: PivotDashboardContext = {
  runtimeClient: {} as RuntimeClient,
  metricsViewName: readable("metrics_view"),
  queryClient: {} as QueryClient,
  enabled: readable(true),
};

// One publisher row by `domain` columns, with a single measure.
const config: PivotDataStoreConfig = {
  measureNames: ["impressions"],
  rowDimensionNames: ["publisher"],
  colDimensionNames: ["domain"],
  allMeasures: [],
  allDimensions: [],
  whereFilter: createAndExpression([]),
  pivot: {
    activeCell: null,
    columnPage: 1,
    columns: [],
    enableComparison: false,
    expanded: {},
    rowPage: 1,
    rows: [],
    sorting: [],
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

// The values of an `IN` filter on `domain` anywhere in the expression.
function domainsInFilter(expr: V1Expression | undefined): string[] | undefined {
  const [ident, ...values] = expr?.cond?.exprs ?? [];
  if (
    expr?.cond?.op === V1Operation.OPERATION_IN &&
    ident?.ident === "domain"
  ) {
    return values.map((value) => value.val as string);
  }
  for (const child of expr?.cond?.exprs ?? []) {
    const found = domainsInFilter(child);
    if (found) return found;
  }
  return undefined;
}

// Answers every aggregation the store makes for a table whose `domain`
// dimension has `domainCount` values, honouring each request's filter on
// `domain` and its limit.
function mockAggregations(domainCount: number) {
  const domains = Array.from({ length: domainCount }, (_, i) => `d${i}`);
  vi.mocked(createQueryServiceMetricsViewAggregation).mockImplementation(((
    _client: RuntimeClient,
    request: V1MetricsViewAggregationRequest,
  ) => {
    const dimensions = (request.dimensions ?? []).map((d) => d.name);
    const filteredDomains = domainsInFilter(request.where);
    const requestDomains = filteredDomains
      ? domains.filter((domain) => filteredDomains.includes(domain))
      : domains;
    let rows: Record<string, unknown>[];
    if (dimensions.length === 0) {
      rows = [{ impressions: domainCount }];
    } else if (dimensions.join() === "publisher") {
      rows = [{ publisher: "A", impressions: domainCount }];
    } else if (dimensions.join() === "domain") {
      rows = requestDomains.map((domain) => ({ domain, impressions: 1 }));
    } else {
      rows = requestDomains.map((domain) => ({
        domain,
        publisher: "A",
        impressions: 1,
      }));
    }
    const limit = Number(request.limit ?? rows.length);
    return readable({
      isFetching: false,
      isError: false,
      data: { data: rows.slice(0, limit) },
    });
  }) as unknown as typeof createQueryServiceMetricsViewAggregation);
}

describe("createPivotDataStore column truncation", () => {
  it("knows the total when a dimension has exactly the value limit", () => {
    // 100 domains fill the axis without passing its limit, so the total of
    // 100 columns is known even though only the first page is shown.
    mockAggregations(100);

    const state = get(createPivotDataStore(ctx, readable(config)));

    expect(state.columnTruncation).toEqual({ shown: 50, total: 100 });
  });

  it("leaves the total out when a dimension has more values than the limit", () => {
    mockAggregations(101);

    const state = get(createPivotDataStore(ctx, readable(config)));

    expect(state.columnTruncation).toEqual({ shown: 50, total: undefined });
  });
});

it("reloads an expanded child after its parent's rows arrive for a new measure", () => {
  const rowDimensions = ["country", "device", "campaign"];
  const parentRows = writable<{
    isFetching: boolean;
    data?: { data: Record<string, unknown>[] };
  }>({ isFetching: true });
  vi.mocked(createQueryServiceMetricsViewAggregation).mockImplementation(((
    _client: RuntimeClient,
    request: V1MetricsViewAggregationRequest,
  ) => {
    const dimension = request.dimensions?.[0]?.name;
    if (
      dimension === "device" &&
      request.measures?.some((measure) => measure.name === "clicks")
    ) {
      return parentRows;
    }
    const values: Record<string, string> = {
      country: "US",
      device: "Mobile",
      campaign: "campaign-1",
    };
    return readable({
      isFetching: false,
      data: {
        data: [
          {
            ...(dimension ? { [dimension]: values[dimension] } : {}),
            impressions: 10,
            clicks: 2,
          },
        ],
      },
    });
  }) as unknown as typeof createQueryServiceMetricsViewAggregation);

  const configStore = writable<PivotDataStoreConfig>({
    ...config,
    allMeasures: [{ name: "impressions" }, { name: "clicks" }],
    colDimensionNames: [],
    rowDimensionNames: rowDimensions,
    pivot: {
      ...config.pivot,
      expanded: {
        [buildExpandKey(rowDimensions, ["US"])]: true,
        [buildExpandKey(rowDimensions, ["US", "Mobile"])]: true,
      },
    },
  });
  let state: PivotDataState | undefined;
  const unsubscribe = createPivotDataStore(ctx, configStore).subscribe(
    (next) => (state = next),
  );
  const leaf = () => state?.data[0]?.subRows?.[0]?.subRows?.[0];
  const childQueries = () =>
    vi
      .mocked(createQueryServiceMetricsViewAggregation)
      .mock.calls.filter(
        ([, request]) =>
          request.dimensions?.[0]?.name === "campaign" &&
          request.measures?.some((measure) => measure.name === "clicks"),
      );

  try {
    expect(leaf()).toMatchObject({ country: "campaign-1", impressions: 10 });

    configStore.update((current) => ({
      ...current,
      measureNames: ["impressions", "clicks"],
    }));
    // The saved child key cannot resolve while its parent is still loading.
    expect(childQueries()).toHaveLength(0);

    parentRows.set({
      isFetching: false,
      data: { data: [{ device: "Mobile", impressions: 10, clicks: 2 }] },
    });

    expect(childQueries().length).toBeGreaterThan(0);
    expect(leaf()).toMatchObject({
      country: "campaign-1",
      impressions: 10,
      clicks: 2,
    });
  } finally {
    unsubscribe();
  }
});
