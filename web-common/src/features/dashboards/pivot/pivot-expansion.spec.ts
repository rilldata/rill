import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { QueryClient } from "@tanstack/svelte-query";
import { get, readable } from "svelte/store";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  LOADING_CELL,
  SHOW_MORE_BUTTON,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { buildExpandKey } from "./pivot-expand-keys";
import {
  addExpandedDataToPivot,
  createSubTableCellQuery,
  getValuesForExpandedKey,
  queryExpandedRowMeasureValues,
} from "./pivot-expansion";
import { getAxisForDimensions } from "./pivot-queries";
import {
  type PivotDashboardContext,
  type PivotDataRow,
  type PivotDataStoreConfig,
} from "./types";

vi.mock("./pivot-queries", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./pivot-queries")>()),
  getAxisForDimensions: vi.fn(() => readable({ isFetching: true })),
}));

function getConfig(
  showTotalsRow: boolean,
  rowDimensionNames = ["publisher", "campaign"],
): PivotDataStoreConfig {
  return {
    measureNames: ["impressions"],
    rowDimensionNames,
    colDimensionNames: [],
    allMeasures: [],
    allDimensions: [],
    whereFilter: createAndExpression([]),
    pivot: {
      rows: [],
      columns: [],
      sorting: [],
      expanded: {},
      columnPage: 1,
      rowPage: 1,
      enableComparison: true,
      tableMode: "nest",
      activeCell: null,
      showTotalsColumn: true,
      showTotalsRow,
    },
    time: {
      timeStart: undefined,
      timeEnd: undefined,
      timeZone: "UTC",
      timeDimension: "timestamp",
    },
    enableComparison: false,
    comparisonTime: undefined,
    searchText: undefined,
    isFlat: false,
  } as unknown as PivotDataStoreConfig;
}

describe("pivot expansion (value-based keys)", () => {
  it("fills the node matched by dimension value, regardless of totals row", () => {
    for (const showTotals of [false, true]) {
      const tableData: PivotDataRow[] = [
        { publisher: "A", subRows: [{ publisher: LOADING_CELL }] },
        { publisher: "B", subRows: [{ publisher: LOADING_CELL }] },
      ];

      addExpandedDataToPivot(
        getConfig(showTotals),
        tableData,
        ["publisher", "campaign"],
        {},
        [
          {
            isFetching: false,
            expandIndex: buildExpandKey(["publisher", "campaign"], ["B"]),
            rowDimensionValues: ["B"],
            totals: [{ campaign: "campaign-1", impressions: 10 }],
            data: [],
          },
        ],
      );

      // Row "A" is untouched; row "B" (matched by value, not position) is filled.
      expect(tableData[0].subRows?.[0]?.publisher).toBe(LOADING_CELL);
      expect(tableData[1].subRows?.[0]).toMatchObject({
        publisher: "campaign-1",
        campaign: "campaign-1",
        impressions: 10,
      });
    }
  });

  it("resolves a nested value path to the correct deep node", () => {
    const tableData: PivotDataRow[] = [
      {
        publisher: "A",
        subRows: [
          { publisher: "camp1", subRows: [{ publisher: LOADING_CELL }] },
        ],
      },
    ];

    addExpandedDataToPivot(
      getConfig(true, ["publisher", "campaign", "adgroup"]),
      tableData,
      ["publisher", "campaign", "adgroup"],
      {},
      [
        {
          isFetching: false,
          expandIndex: buildExpandKey(
            ["publisher", "campaign", "adgroup"],
            ["A", "camp1"],
          ),
          rowDimensionValues: ["A", "camp1"],
          totals: [{ adgroup: "ad-1", impressions: 5 }],
          data: [],
        },
      ],
    );

    expect(tableData[0].subRows?.[0]?.subRows?.[0]).toMatchObject({
      publisher: "ad-1",
      adgroup: "ad-1",
      impressions: 5,
    });
  });

  it("does nothing when the value path matches no row", () => {
    const tableData: PivotDataRow[] = [
      { publisher: "A", subRows: [{ publisher: LOADING_CELL }] },
    ];
    addExpandedDataToPivot(
      getConfig(false),
      tableData,
      ["publisher", "campaign"],
      {},
      [
        {
          isFetching: false,
          expandIndex: buildExpandKey(
            ["publisher", "campaign"],
            ["does-not-exist"],
          ),
          rowDimensionValues: ["does-not-exist"],
          totals: [{ campaign: "c", impressions: 1 }],
          data: [],
        },
      ],
    );
    expect(tableData[0].subRows?.[0]?.publisher).toBe(LOADING_CELL);
  });

  it("does not overwrite a matched ancestor when a deeper segment is missing", () => {
    const tableData: PivotDataRow[] = [
      {
        publisher: "A",
        subRows: [
          { publisher: "camp1", subRows: [{ publisher: LOADING_CELL }] },
        ],
      },
    ];
    const before = structuredClone(tableData);
    addExpandedDataToPivot(
      getConfig(false, ["publisher", "campaign", "adgroup", "ad"]),
      tableData,
      ["publisher", "campaign", "adgroup", "ad"],
      {},
      [
        {
          isFetching: false,
          expandIndex: buildExpandKey(
            ["publisher", "campaign", "adgroup"],
            ["A", "camp1", "missing"],
          ),
          rowDimensionValues: ["ad-1"],
          totals: [{ ad: "ad-1", impressions: 1 }],
          data: [],
        },
      ],
    );
    expect(tableData).toEqual(before);
  });

  it("does not confuse a null row with a literal '<NULL>' row", () => {
    const tableData: PivotDataRow[] = [
      { publisher: null, subRows: [{ publisher: LOADING_CELL }] },
      { publisher: "<NULL>", subRows: [{ publisher: LOADING_CELL }] },
    ];
    addExpandedDataToPivot(
      getConfig(false),
      tableData,
      ["publisher", "campaign"],
      {},
      [
        {
          isFetching: false,
          expandIndex: buildExpandKey(["publisher", "campaign"], ["<NULL>"]),
          rowDimensionValues: ["<NULL>"],
          totals: [{ campaign: "c", impressions: 1 }],
          data: [],
        },
      ],
    );
    expect(tableData[0].subRows?.[0]?.publisher).toBe(LOADING_CELL);
    expect(tableData[1].subRows?.[0]).toMatchObject({ campaign: "c" });
  });

  it("fills a parent before its child, whatever the expanded order", () => {
    // A child key can come first in `expanded`, e.g. after its parent was
    // collapsed and expanded again.
    const dims = ["publisher", "campaign", "adgroup"];
    const tableData: PivotDataRow[] = [
      { publisher: "A", subRows: [{ publisher: LOADING_CELL }] },
    ];

    addExpandedDataToPivot(getConfig(false, dims), tableData, dims, {}, [
      {
        isFetching: false,
        expandIndex: buildExpandKey(dims, ["A", "camp1"]),
        rowDimensionValues: ["ad-1"],
        totals: [{ adgroup: "ad-1", impressions: 5 }],
        data: [],
      },
      {
        isFetching: false,
        expandIndex: buildExpandKey(dims, ["A"]),
        rowDimensionValues: ["camp1"],
        totals: [{ campaign: "camp1", impressions: 5 }],
        data: [],
      },
    ]);

    expect(tableData[0].subRows?.[0]?.subRows?.[0]).toMatchObject({
      publisher: "ad-1",
      adgroup: "ad-1",
    });
  });

  it("ends a group that has more rows past the last limit step", () => {
    const tableData: PivotDataRow[] = [
      {
        publisher: "A",
        subRows: [{ publisher: LOADING_CELL }],
      },
    ];

    addExpandedDataToPivot(
      getConfig(true),
      tableData,
      ["publisher", "campaign"],
      {},
      [
        {
          isFetching: false,
          expandIndex: buildExpandKey(["publisher", "campaign"], ["A"]),
          rowDimensionValues: ["A"],
          totals: [{ campaign: "campaign-1", impressions: 10 }],
          data: [],
          hasMore: true,
          effectiveLimit: 100,
        },
      ],
    );

    expect(tableData[0].subRows?.at(-1)).toEqual({
      publisher: SHOW_MORE_BUTTON,
      __currentLimit: 100,
    });
  });

  it("does not end a group that has exactly the limit of rows", () => {
    const tableData: PivotDataRow[] = [
      {
        publisher: "A",
        subRows: [{ publisher: LOADING_CELL }],
      },
    ];

    addExpandedDataToPivot(
      getConfig(true),
      tableData,
      ["publisher", "campaign"],
      {},
      [
        {
          isFetching: false,
          expandIndex: buildExpandKey(["publisher", "campaign"], ["A"]),
          rowDimensionValues: ["A"],
          totals: [{ campaign: "campaign-1", impressions: 10 }],
          data: [],
          hasMore: false,
          effectiveLimit: 100,
        },
      ],
    );

    expect(tableData[0].subRows?.at(-1)).toMatchObject({
      campaign: "campaign-1",
    });
  });
});

describe("getValuesForExpandedKey", () => {
  const tableData: PivotDataRow[] = [
    {
      publisher: "A",
      subRows: [{ publisher: "camp1" }, { publisher: "camp2" }],
    },
    { publisher: "B", subRows: [{ publisher: "camp3" }] },
  ];

  it("returns the actual values along the matched path", () => {
    expect(
      getValuesForExpandedKey(
        tableData,
        ["publisher", "campaign"],
        buildExpandKey(["publisher", "campaign"], ["B"]),
      ),
    ).toEqual(["B"]);
    expect(
      getValuesForExpandedKey(
        tableData,
        ["publisher", "campaign"],
        buildExpandKey(["publisher", "campaign"], ["A", "camp2"]),
      ),
    ).toEqual(["A", "camp2"]);
  });

  it("stops at the deepest resolvable segment", () => {
    expect(
      getValuesForExpandedKey(
        tableData,
        ["publisher", "campaign"],
        buildExpandKey(["publisher", "campaign"], ["A", "missing"]),
      ),
    ).toEqual(["A"]);
  });

  it("does not match a row of a different dimension with the same value", () => {
    // Expanded while "publisher" was the first row dimension, then the
    // dimension was swapped for "domain", which has a row with the same value.
    expect(
      getValuesForExpandedKey(
        tableData,
        ["domain", "campaign"],
        buildExpandKey(["publisher"], ["A"]),
      ),
    ).toEqual([]);
  });
});

describe("queryExpandedRowMeasureValues", () => {
  const rowDimensionNames = ["publisher", "domain", "campaign"];
  // Rows are [publisher, domain, campaign], so A's children are domains.
  const tableData: PivotDataRow[] = [
    { publisher: "A", subRows: [{ publisher: "example.com" }] },
    { publisher: "", subRows: [{ publisher: "example.com" }] },
  ];

  function query(expandedKey: string) {
    const config = getConfig(false, rowDimensionNames);
    config.pivot.expanded = { [expandedKey]: true };
    return queryExpandedRowMeasureValues(
      {} as PivotDashboardContext,
      config,
      tableData,
      {},
      {},
    );
  }

  beforeEach(() => {
    vi.mocked(getAxisForDimensions).mockClear();
  });

  it("queries the next dimension for a fully resolved key", () => {
    query(buildExpandKey(["publisher", "domain", "campaign"], ["A"]));
    expect(getAxisForDimensions).toHaveBeenCalledTimes(1);
    expect(vi.mocked(getAxisForDimensions).mock.calls[0][2]).toEqual([
      "domain",
    ]);
  });

  it("queries the next dimension for a top-level empty-string row", () => {
    query(buildExpandKey(["publisher", "domain", "campaign"], [""]));
    expect(getAxisForDimensions).toHaveBeenCalledTimes(1);
    expect(vi.mocked(getAxisForDimensions).mock.calls[0][2]).toEqual([
      "domain",
    ]);
  });

  it("does not query for a key that only partially resolves", () => {
    // Expanded as [publisher, campaign] before domain was inserted: "camp1"
    // no longer matches a row at depth 1, so only "A" resolves.
    query(buildExpandKey(["publisher", "campaign"], ["A", "camp1"]));
    expect(getAxisForDimensions).not.toHaveBeenCalled();
  });
});

describe("createSubTableCellQuery", () => {
  // An expanded row whose sub rows query came back empty, for example after
  // the filter or time range changed. There are no cells to fetch, and the row
  // filter would be an `IN` with no values, which the runtime rejects.
  it("skips the query for a row with no sub rows", () => {
    const requests: unknown[] = [];
    const ctx: PivotDashboardContext = {
      runtimeClient: {
        instanceId: "default",
        queryService: {
          metricsViewAggregation: (request: unknown) => {
            requests.push(request);
            return Promise.resolve({ toJson: () => ({ data: [] }) });
          },
        },
      } as unknown as RuntimeClient,
      metricsViewName: readable("metrics_view"),
      queryClient: new QueryClient({
        defaultOptions: { queries: { retry: false } },
      }),
      enabled: readable(true),
    };
    const config = { ...getConfig(false), colDimensionNames: ["country"] };

    const query = createSubTableCellQuery(
      ctx,
      config,
      "campaign",
      { country: ["US"] },
      { c0v0m0: 10, impressions: 10 },
      [],
      createAndExpression([createInExpression("publisher", ["A"])]),
      [],
    );

    expect(get(query)).toBeNull();
    expect(requests).toHaveLength(0);
  });
});
