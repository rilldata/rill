import {
  createAndExpression,
  createInExpression,
  forEachExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  V1Operation,
  type V1Expression,
  type V1MetricsViewAggregationRequest,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  get,
  readable,
  writable,
  type Readable,
  type Writable,
} from "svelte/store";
import { describe, expect, it, vi } from "vitest";
import {
  CircularChartProvider,
  type CircularChartSpec,
} from "./circular/CircularChartProvider";
import {
  FunnelChartProvider,
  type FunnelChartSpec,
} from "./funnel/FunnelChartProvider";

type QueryOptions = { queryKey: unknown[]; enabled?: boolean };
type QueryResult = {
  data: unknown;
  error: null;
  isFetching: boolean;
  isPending: boolean;
  isPlaceholderData: boolean;
};
type CreatedQuery = {
  options: Readable<QueryOptions>;
  result: Writable<QueryResult>;
};

// The providers call `createQuery` without a query client, which needs a Svelte
// component context. It is replaced with a store the test fills in, and the
// queries are listed in the order the provider creates them.
const { createdQueries } = vi.hoisted(() => ({
  createdQueries: [] as CreatedQuery[],
}));

vi.mock("@tanstack/svelte-query", async (importOriginal) => {
  const actual =
    await importOriginal<typeof import("@tanstack/svelte-query")>();
  const { writable } = await import("svelte/store");
  return {
    ...actual,
    createQuery: (options: Readable<QueryOptions>) => {
      const result = writable<QueryResult>({
        data: undefined,
        error: null,
        isFetching: false,
        isPending: true,
        isPlaceholderData: false,
      });
      createdQueries.push({ options, result });
      return result;
    },
  };
});

// Only the instance id is read while the query options are built.
const client = { instanceId: "default" } as RuntimeClient;

function createTimeAndFilterStore(where?: V1Expression) {
  return writable<TimeAndFilterStore>({
    timeRange: { start: "2024-01-01T00:00:00Z", end: "2024-02-01T00:00:00Z" },
    comparisonTimeRange: undefined,
    where,
    timeGrain: undefined,
    showTimeComparison: false,
    timeRangeState: undefined,
    comparisonTimeRangeState: undefined,
    hasTimeSeries: true,
  });
}

function requestOf(query: CreatedQuery) {
  return get(query.options).queryKey[3] as V1MetricsViewAggregationRequest;
}

function resolve(query: CreatedQuery, rows: Record<string, unknown>[]) {
  query.result.set({
    data: { data: rows },
    error: null,
    isFetching: false,
    isPending: false,
    isPlaceholderData: false,
  });
}

// Whether every binary condition in `expr` has an operand on each side.
// The runtime rejects a condition that does not, an `IN` with no values for example.
function binaryConditionsAreComplete(expr: V1Expression | undefined) {
  let complete = true;
  if (!expr) return complete;
  forEachExpression(expr, (e) => {
    const op = e.cond?.op;
    if (
      !op ||
      op === V1Operation.OPERATION_AND ||
      op === V1Operation.OPERATION_OR
    ) {
      return;
    }
    if ((e.cond?.exprs?.length ?? 0) < 2) complete = false;
  });
  return complete;
}

const globalFilter = createAndExpression([
  createInExpression("region", ["EU"]),
]);

describe("CircularChartProvider data query", () => {
  const spec: CircularChartSpec = {
    metrics_view: "sales",
    measure: { field: "revenue", type: "quantitative" },
    color: { field: "country", type: "nominal", limit: 8 },
  };

  function createQueries(where?: V1Expression) {
    createdQueries.length = 0;
    const provider = new CircularChartProvider(readable(spec));
    const chartQuery = provider.createChartDataQuery(
      client,
      createTimeAndFilterStore(where),
    );
    const unsubscribe = chartQuery.subscribe(() => {});
    // Created in this order: top-N, total, "other", data.
    expect(createdQueries).toHaveLength(4);
    const [topN, , , data] = createdQueries;
    return { topN, data, unsubscribe };
  }

  it("narrows the data query to the values the top-N query found", () => {
    const { topN, data, unsubscribe } = createQueries();
    resolve(topN, [{ country: "US" }, { country: "CA" }]);

    expect(requestOf(data).where).toEqual(
      createInExpression("country", ["US", "CA"]),
    );
    expect(get(data.options).enabled).toBe(true);
    unsubscribe();
  });

  it.each([
    ["without a global filter", undefined],
    ["with a global filter", globalFilter],
  ])(
    "sends no IN condition without values when the top-N query finds none %s",
    (_, where) => {
      const { topN, data, unsubscribe } = createQueries(where);
      resolve(topN, []);

      const request = requestOf(data);
      expect(binaryConditionsAreComplete(request.where)).toBe(true);
      // The data query takes the top-N query's filter, so it finds nothing either.
      expect(request.where).toEqual(requestOf(topN).where);
      expect(get(data.options).enabled).toBe(true);
      unsubscribe();
    },
  );
});

describe("FunnelChartProvider data query", () => {
  // With `showNull` the stage filter adds nothing to the base filter, so the
  // top values are the only condition the data query has.
  const spec: FunnelChartSpec = {
    metrics_view: "sales",
    measure: { field: "revenue", type: "quantitative" },
    stage: { field: "step", type: "nominal", showNull: true },
  };

  function createQueries(where?: V1Expression) {
    createdQueries.length = 0;
    const provider = new FunnelChartProvider(readable(spec));
    const chartQuery = provider.createChartDataQuery(
      client,
      createTimeAndFilterStore(where),
    );
    const unsubscribe = chartQuery.subscribe(() => {});
    // Created in this order: top-N, data.
    expect(createdQueries).toHaveLength(2);
    const [topN, data] = createdQueries;
    return { topN, data, unsubscribe };
  }

  it("narrows the data query to the stages the top-N query found", () => {
    const { topN, data, unsubscribe } = createQueries();
    resolve(topN, [{ step: "visit" }, { step: "buy" }]);

    expect(requestOf(data).where).toEqual(
      createInExpression("step", ["visit", "buy"]),
    );
    unsubscribe();
  });

  it.each([
    ["without a global filter", undefined],
    ["with a global filter", globalFilter],
  ])(
    "sends no IN condition without values when the top-N query finds none %s",
    (_, where) => {
      const { topN, data, unsubscribe } = createQueries(where);
      resolve(topN, []);

      const request = requestOf(data);
      expect(binaryConditionsAreComplete(request.where)).toBe(true);
      expect(request.where).toEqual(requestOf(topN).where);
      expect(get(data.options).enabled).toBe(true);
      unsubscribe();
    },
  );
});
