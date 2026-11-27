import { MAX_EPHEMERAL_MEASURES } from "./validation";
import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_METRICS_VIEW,
  AD_BIDS_NAME,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import {
  applyURLToExploreState,
  getCleanMetricsExploreForAssertion,
  useTestFilterManager,
} from "@rilldata/web-common/features/dashboards/url-state/test/url-state-test-utils";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { get } from "svelte/store";
import { beforeEach, describe, expect, it } from "vitest";
import {
  fromEphemeralMeasuresParam,
  toEphemeralMeasuresParam,
} from "./url-param";

describe("ephemeral url param", () => {
  describe("round trips", () => {
    const cases = [
      [{ name: "profit", displayName: "Profit", expression: "revenue - cost" }],
      [
        {
          name: "profit",
          displayName: "Profit",
          expression: "revenue - cost",
          formatPreset: "currency_usd",
        },
        {
          name: "arpu",
          displayName: "ARPU (µ)",
          expression: "round(revenue / users, 2)",
        },
      ],
      // A description with and without a format preset.
      [
        {
          name: "margin",
          displayName: "Margin",
          expression: "profit / revenue",
          description: "Profit per dollar: of revenue; net",
        },
        {
          name: "arpu",
          displayName: "ARPU",
          expression: "revenue / users",
          formatPreset: "currency_usd",
          description: "Average revenue per user",
        },
      ],
      // Hostile characters in display names and expressions.
      [
        {
          name: "weird",
          displayName: 'a:b;c,d%e+f"g',
          expression: '"total volume" % 100 + 0.5',
        },
      ],
    ];
    it.each(cases)("%j", (...defs) => {
      const param = toEphemeralMeasuresParam(defs);
      const { ephemeralMeasures, invalidEntries } =
        fromEphemeralMeasuresParam(param);
      expect(invalidEntries).toEqual([]);
      expect(ephemeralMeasures).toEqual(defs);
    });
  });

  it("returns an empty param for no definitions", () => {
    expect(toEphemeralMeasuresParam(undefined)).toBe("");
    expect(toEphemeralMeasuresParam([])).toBe("");
  });

  it("drops malformed entries and keeps the rest", () => {
    const { ephemeralMeasures, invalidEntries } = fromEphemeralMeasuresParam(
      "profit:Profit:revenue;bad-entry;other:Other:cost",
    );
    expect(ephemeralMeasures.map((d) => d.name)).toEqual(["profit", "other"]);
    expect(invalidEntries).toEqual(["bad-entry"]);
  });

  it("drops duplicate names", () => {
    const { ephemeralMeasures, invalidEntries } = fromEphemeralMeasuresParam(
      "profit:Profit:revenue;profit:Other:cost",
    );
    expect(ephemeralMeasures.map((d) => d.name)).toEqual(["profit"]);
    expect(invalidEntries).toEqual(["profit:Other:cost"]);
  });
});

// Filters live in the ExpressionFilterManager, which needs the specs of the
// metrics view backing the explore.
const getFilterManager = useTestFilterManager({
  [AD_BIDS_NAME]: AD_BIDS_METRICS_VIEW,
});

describe("ephemeral measures URL state integration", () => {
  beforeEach(() => {
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
    metricsExplorerStore.init(
      AD_BIDS_EXPLORE_NAME,
      getInitExploreStateForTest(
        AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
        AD_BIDS_EXPLORE_INIT,
        AD_BIDS_TIME_RANGE_SUMMARY,
      ),
    );
  });

  function applyUrl(url: string) {
    return applyURLToExploreState(
      new URL(url),
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
      AD_BIDS_EXPLORE_INIT,
      getFilterManager(),
    );
  }

  it("restores ephemeral measures and their pivot columns from the URL", () => {
    const errors = applyUrl(
      "http://localhost/explore/AdBids_explore?view=pivot&rows=publisher&cols=impressions,profit&adhoc_m=profit:Profit:impressions*2",
    );
    expect(errors).toEqual([]);

    const state = getCleanMetricsExploreForAssertion();
    expect(state.ephemeralMeasures).toEqual([
      { name: "profit", displayName: "Profit", expression: "impressions*2" },
    ]);
    expect(state.pivot?.columns).toEqual([
      { id: "impressions", title: "impressions", type: PivotChipType.Measure },
      { id: "profit", title: "Profit", type: PivotChipType.Measure },
    ]);
  });

  it("shows a new measure first in the explore view", () => {
    metricsExplorerStore.addEphemeralMeasure(AD_BIDS_EXPLORE_NAME, {
      name: "profit",
      displayName: "Profit",
      expression: "impressions * 2",
    });
    const state = getCleanMetricsExploreForAssertion();
    expect(state.visibleMeasures?.[0]).toBe("profit");
    expect(state.allMeasuresVisible).toBe(false);
  });

  it("adds a new measure as the first measure column of the pivot", () => {
    applyUrl(
      "http://localhost/explore/AdBids_explore?view=pivot&rows=publisher&cols=domain,impressions,bid_price",
    );
    metricsExplorerStore.addEphemeralMeasure(AD_BIDS_EXPLORE_NAME, {
      name: "profit",
      displayName: "Profit",
      expression: "impressions * 2",
    });
    const state = getCleanMetricsExploreForAssertion();
    expect(state.pivot?.columns.map((col) => col.id)).toEqual([
      "domain",
      "profit",
      "impressions",
      "bid_price",
    ]);
    // It is also the first shown measure on returning to the explore view.
    expect(state.visibleMeasures?.[0]).toBe("profit");
  });

  it("drops positional sorts that a new first measure column would shift", () => {
    const profit = {
      name: "profit",
      displayName: "Profit",
      expression: "impressions * 2",
    };
    const sortWith = (tableMode: "nest" | "flat", id: string) => {
      applyUrl(
        `http://localhost/explore/AdBids_explore?view=pivot&rows=publisher&cols=domain,impressions,bid_price&table_mode=${tableMode}`,
      );
      get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME].pivot.sorting = [
        { id, desc: true },
      ];
      metricsExplorerStore.addEphemeralMeasure(AD_BIDS_EXPLORE_NAME, profit);
      return getCleanMetricsExploreForAssertion().pivot?.sorting;
    };

    // In a nested table `c0v1m0` names the first measure under a column value.
    expect(sortWith("nest", "c0v1m0")).toEqual([]);
    expect(sortWith("nest", "publisher")).toEqual([
      { id: "publisher", desc: true },
    ]);
    // A flat table sorts by field name, which an insert does not shift.
    expect(sortWith("flat", "impressions")).toEqual([
      { id: "impressions", desc: true },
    ]);
  });

  it("drops positional sorts that removing a measure column shifts", () => {
    const sortAfterRemoving = (cols: string) => {
      applyUrl(
        `http://localhost/explore/AdBids_explore?view=pivot&rows=publisher&cols=${cols}&adhoc_m=profit:Profit:impressions*2`,
      );
      get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME].pivot.sorting = [
        { id: "c0v1m1", desc: true },
      ];
      metricsExplorerStore.removeEphemeralMeasure(
        AD_BIDS_EXPLORE_NAME,
        "profit",
        AD_BIDS_EXPLORE_INIT,
      );
      return getCleanMetricsExploreForAssertion().pivot?.sorting;
    };

    expect(sortAfterRemoving("domain,profit,impressions")).toEqual([]);
    // Without a profit column, no measure moves.
    expect(sortAfterRemoving("domain,impressions,bid_price")).toEqual([
      { id: "c0v1m1", desc: true },
    ]);
  });

  it("drops definitions referencing unknown measures, and their columns", () => {
    const errors = applyUrl(
      "http://localhost/explore/AdBids_explore?view=pivot&cols=impressions,profit&adhoc_m=profit:Profit:unknown*2",
    );
    expect(errors.map((e) => e.message)).toEqual([
      `Selected adhoc measure: "profit ("unknown" is not a measure in this dashboard)" is not valid.`,
      `Selected pivot column: "profit" is not valid.`,
    ]);

    const state = getCleanMetricsExploreForAssertion();
    expect(state.ephemeralMeasures).toBeUndefined();
    expect(state.pivot?.columns).toEqual([
      { id: "impressions", title: "impressions", type: PivotChipType.Measure },
    ]);
  });

  it("drops definitions named after a measure hidden from the explore", () => {
    // publisher_count exists in the metrics view but not in the explore, and
    // the server rejects computed fields that collide with any measure.
    const errors = applyUrl(
      "http://localhost/explore/AdBids_explore?view=pivot&cols=publisher_count&adhoc_m=publisher_count:PC:impressions*2",
    );
    expect(errors.length).toBe(2);
    expect(errors[0].message).toContain("already used by another field");
    expect(
      getCleanMetricsExploreForAssertion().ephemeralMeasures,
    ).toBeUndefined();
  });

  it("does not report all measures visible while a spec measure is hidden", () => {
    const errors = applyUrl(
      "http://localhost/explore/AdBids_explore?measures=impressions,profit&adhoc_m=profit:Profit:impressions*2",
    );
    expect(errors).toEqual([]);
    const state = getCleanMetricsExploreForAssertion();
    expect(state.visibleMeasures).toEqual(["impressions", "profit"]);
    expect(state.allMeasuresVisible).toBe(false);
  });

  it("drops definitions whose name collides with a metrics view field", () => {
    const errors = applyUrl(
      "http://localhost/explore/AdBids_explore?view=pivot&cols=bid_price&adhoc_m=bid_price:Custom:impressions*2",
    );
    expect(errors.length).toBe(1);
    expect(errors[0].message).toContain("already used by another field");

    const state = getCleanMetricsExploreForAssertion();
    expect(state.ephemeralMeasures).toBeUndefined();
  });

  it("accepts conditional formatting on an ephemeral measure", () => {
    const errors = applyUrl(
      "http://localhost/explore/AdBids_explore?view=pivot&cols=profit&adhoc_m=profit:Profit:impressions*2&format=profit:heatmap:greens",
    );
    expect(errors).toEqual([]);

    const state = getCleanMetricsExploreForAssertion();
    expect(state.pivot?.measureFormatting).toEqual({
      profit: { mode: "heatmap", scheme: "greens" },
    });
  });

  it("survives a state → URL → state round trip", async () => {
    const { convertPartialExploreStateToUrlParams } = await import(
      "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params"
    );
    const { convertURLSearchParamsToExploreState } = await import(
      "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState"
    );
    const { AD_BIDS_METRICS_VIEW } = await import(
      "@rilldata/web-common/features/dashboards/stores/test-data/data"
    );

    // Switch to the pivot view first so the add also places a pivot column.
    get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME].activePage =
      DashboardState_ActivePage.PIVOT;
    metricsExplorerStore.addEphemeralMeasure(AD_BIDS_EXPLORE_NAME, {
      name: "profit",
      displayName: "Profit (net)",
      expression: 'impressions * 2 + "bid_price"',
    });
    const exploreState =
      get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME];

    const urlParams = convertPartialExploreStateToUrlParams(
      AD_BIDS_EXPLORE_INIT,
      AD_BIDS_METRICS_VIEW,
      exploreState,
      undefined,
    );
    // Serialize + reparse, as a browser would.
    const reparsed = new URLSearchParams(urlParams.toString());
    // An empty `f=` param errors in the filter parser; in the app it is
    // stripped by cleanUrlParams, which is not part of this round trip.
    if (reparsed.get("f") === "") reparsed.delete("f");
    const { partialExploreState, errors } =
      convertURLSearchParamsToExploreState(
        reparsed,
        AD_BIDS_METRICS_VIEW,
        AD_BIDS_EXPLORE_INIT,
      );
    expect(errors).toEqual([]);
    expect(partialExploreState.ephemeralMeasures).toEqual([
      {
        name: "profit",
        displayName: "Profit (net)",
        expression: 'impressions * 2 + "bid_price"',
      },
    ]);
    expect(
      partialExploreState.pivot?.columns.find((c) => c.id === "profit"),
    ).toEqual({
      id: "profit",
      title: "Profit (net)",
      type: PivotChipType.Measure,
    });
  });
});

describe("ephemeral measures param limit", () => {
  it("writes at most the number of definitions the parser accepts", () => {
    const defs = Array.from({ length: MAX_EPHEMERAL_MEASURES + 2 }, (_, i) => ({
      name: `m${i}`,
      displayName: `M${i}`,
      expression: "impressions * 2",
    }));
    const { ephemeralMeasures, invalidEntries } = fromEphemeralMeasuresParam(
      toEphemeralMeasuresParam(defs),
    );
    expect(ephemeralMeasures).toEqual(defs.slice(0, MAX_EPHEMERAL_MEASURES));
    expect(invalidEntries).toEqual([]);
  });
});
