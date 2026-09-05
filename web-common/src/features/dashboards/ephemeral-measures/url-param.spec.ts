import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import {
  applyURLToExploreState,
  getCleanMetricsExploreForAssertion,
} from "@rilldata/web-common/features/dashboards/url-state/url-state-variations.spec";
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
    const defaultExplorePreset = getDefaultExplorePreset(
      AD_BIDS_EXPLORE_INIT,
      AD_BIDS_METRICS_INIT,
      AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
    );
    return applyURLToExploreState(
      new URL(url),
      AD_BIDS_EXPLORE_INIT,
      defaultExplorePreset,
    );
  }

  it("restores ephemeral measures and their pivot columns from the URL", () => {
    const errors = applyUrl(
      "http://localhost/explore/AdBids_explore?view=pivot&rows=publisher&cols=impressions,profit&ephemeral=profit:Profit:impressions*2",
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

  it("drops definitions referencing unknown measures, and their columns", () => {
    const errors = applyUrl(
      "http://localhost/explore/AdBids_explore?view=pivot&cols=impressions,profit&ephemeral=profit:Profit:unknown*2",
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

  it("drops definitions whose name collides with a metrics view field", () => {
    const errors = applyUrl(
      "http://localhost/explore/AdBids_explore?view=pivot&cols=bid_price&ephemeral=bid_price:Custom:impressions*2",
    );
    expect(errors.length).toBe(1);
    expect(errors[0].message).toContain("already used by another field");

    const state = getCleanMetricsExploreForAssertion();
    expect(state.ephemeralMeasures).toBeUndefined();
  });

  it("accepts conditional formatting on an ephemeral measure", () => {
    const errors = applyUrl(
      "http://localhost/explore/AdBids_explore?view=pivot&cols=profit&ephemeral=profit:Profit:impressions*2&format=profit:heatmap:greens",
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
        {},
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
