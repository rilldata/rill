import {
  type PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { buildExpandKey } from "@rilldata/web-common/features/dashboards/pivot/pivot-expand-keys";
import {
  getUpdatedPinIndex,
  metricsExplorerStore,
} from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  createAndExpression,
  createInExpression,
  createLikeExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import { get } from "svelte/store";
import { beforeEach, describe, expect, it } from "vitest";

const DIMENSION = "publisher";

function whereFilterWith(values: string[]) {
  return createAndExpression([createInExpression(DIMENSION, values)]);
}

describe("getUpdatedPinIndex", () => {
  it("decrements the pin when a value before it is removed", () => {
    expect(
      getUpdatedPinIndex(
        2,
        DIMENSION,
        whereFilterWith(["Facebook", "Google", "Yahoo", "Microsoft"]),
        whereFilterWith(["Google", "Yahoo", "Microsoft"]),
      ),
    ).toEqual(1);
  });

  it("keeps the pin when a value after it is removed", () => {
    expect(
      getUpdatedPinIndex(
        1,
        DIMENSION,
        whereFilterWith(["Facebook", "Google", "Yahoo", "Microsoft"]),
        whereFilterWith(["Facebook", "Google", "Microsoft"]),
      ),
    ).toEqual(1);
  });

  it("keeps the pin when a value is added", () => {
    expect(
      getUpdatedPinIndex(
        1,
        DIMENSION,
        whereFilterWith(["Facebook", "Google"]),
        whereFilterWith(["Facebook", "Google", "Yahoo"]),
      ),
    ).toEqual(1);
  });

  it("handles multiple pinned values being removed at once", () => {
    expect(
      getUpdatedPinIndex(
        2,
        DIMENSION,
        whereFilterWith(["Facebook", "Google", "Yahoo", "Microsoft"]),
        whereFilterWith(["Google", "Microsoft"]),
      ),
    ).toEqual(0);
  });

  it("drops the pin when every pinned value is removed", () => {
    expect(
      getUpdatedPinIndex(
        1,
        DIMENSION,
        whereFilterWith(["Facebook", "Google", "Yahoo"]),
        whereFilterWith(["Yahoo"]),
      ),
    ).toEqual(-1);
  });

  it("drops the pin when the filter on the dimension is removed", () => {
    expect(
      getUpdatedPinIndex(
        1,
        DIMENSION,
        whereFilterWith(["Facebook", "Google", "Yahoo"]),
        createAndExpression([createInExpression("domain", ["facebook.com"])]),
      ),
    ).toEqual(-1);
  });

  it("drops the pin when all filters are cleared", () => {
    expect(
      getUpdatedPinIndex(
        1,
        DIMENSION,
        whereFilterWith(["Facebook", "Google", "Yahoo"]),
        createAndExpression([]),
      ),
    ).toEqual(-1);
  });

  it("ignores filters on other dimensions", () => {
    expect(
      getUpdatedPinIndex(
        1,
        DIMENSION,
        whereFilterWith(["Facebook", "Google", "Yahoo"]),
        createAndExpression([
          createInExpression(DIMENSION, ["Facebook", "Google", "Yahoo"]),
          createInExpression("domain", ["facebook.com"]),
        ]),
      ),
    ).toEqual(1);
  });

  it("leaves an unpinned dashboard alone", () => {
    expect(
      getUpdatedPinIndex(
        -1,
        DIMENSION,
        whereFilterWith(["Facebook", "Google"]),
        whereFilterWith(["Google"]),
      ),
    ).toEqual(-1);
  });

  it("leaves the pin alone when there is no comparison dimension", () => {
    expect(
      getUpdatedPinIndex(
        1,
        undefined,
        whereFilterWith(["Facebook", "Google"]),
        whereFilterWith(["Google"]),
      ),
    ).toEqual(1);
  });

  it("leaves the pin alone for a contains filter, since its values come from a query", () => {
    expect(
      getUpdatedPinIndex(
        1,
        DIMENSION,
        whereFilterWith(["Facebook", "Google", "Yahoo"]),
        createAndExpression([createLikeExpression(DIMENSION, "%oo%")]),
      ),
    ).toEqual(1);
  });
});

describe("pivot field reducers", () => {
  const dim = (id: string): PivotChipData => ({
    id,
    title: id,
    type: PivotChipType.Dimension,
  });
  const measure = (id: string): PivotChipData => ({
    id,
    title: id,
    type: PivotChipType.Measure,
  });
  const pivot = () =>
    get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME].pivot;

  beforeEach(() => {
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
    metricsExplorerStore.init(
      AD_BIDS_EXPLORE_NAME,
      getInitExploreStateForTest(AD_BIDS_METRICS_INIT, AD_BIDS_EXPLORE_INIT),
    );
  });

  it("setPivotRows keeps a repeated field once", () => {
    metricsExplorerStore.setPivotRows(AD_BIDS_EXPLORE_NAME, [
      dim("publisher"),
      dim("domain"),
      dim("publisher"),
    ]);
    expect(pivot().rows).toEqual([dim("publisher"), dim("domain")]);
  });

  it.each([
    { next: ["publisher", "domain", "campaign", "adgroup"], depths: [1, 2] },
    { next: ["publisher", "domain"], depths: [1] },
    { next: ["publisher", "campaign", "domain"], depths: [1] },
    { next: ["publisher", "device", "campaign"], depths: [1] },
    { next: ["device", "domain", "campaign"], depths: [] },
    { next: ["domain", "publisher", "campaign"], depths: [] },
    { next: ["publisher"], depths: [] },
    { next: [], depths: [] },
  ])("setPivotRows retains valid expansions for $next", ({ next, depths }) => {
    const previous = ["publisher", "domain", "campaign"];
    const keys = [1, 2].map((depth) =>
      buildExpandKey(previous, ["Facebook", "facebook.com"].slice(0, depth)),
    );
    metricsExplorerStore.setPivotRows(AD_BIDS_EXPLORE_NAME, previous.map(dim));
    metricsExplorerStore.setPivotExpanded(
      AD_BIDS_EXPLORE_NAME,
      Object.fromEntries(keys.map((key) => [key, true])),
    );

    metricsExplorerStore.setPivotRows(AD_BIDS_EXPLORE_NAME, next.map(dim));

    expect(pivot().expanded).toEqual(
      Object.fromEntries(depths.map((depth) => [keys[depth - 1], true])),
    );
  });

  it("setPivotRows distinguishes time grains from ordinary dimensions", () => {
    const month = { ...dim("month"), type: PivotChipType.Time };
    metricsExplorerStore.setPivotRows(AD_BIDS_EXPLORE_NAME, [
      month,
      dim("publisher"),
    ]);
    metricsExplorerStore.setPivotExpanded(AD_BIDS_EXPLORE_NAME, {
      [buildExpandKey(["timestamp_rill_month"], ["2022-01-01T00:00:00Z"])]:
        true,
    });

    // Appending a dimension preserves the month despite its encoded time alias.
    metricsExplorerStore.setPivotRows(AD_BIDS_EXPLORE_NAME, [
      month,
      dim("publisher"),
      dim("domain"),
    ]);
    expect(Object.keys(pivot().expanded)).toHaveLength(1);

    metricsExplorerStore.setPivotRows(AD_BIDS_EXPLORE_NAME, [
      dim("month"),
      dim("publisher"),
    ]);
    expect(pivot().expanded).toEqual({});
  });

  it("setPivotColumns keeps a repeated field once", () => {
    metricsExplorerStore.setPivotColumns(AD_BIDS_EXPLORE_NAME, [
      dim("domain"),
      measure("impressions"),
      dim("domain"),
      measure("impressions"),
    ]);
    expect(pivot().columns).toEqual([dim("domain"), measure("impressions")]);
  });

  it("addPivotField ignores a field already placed in either zone", () => {
    metricsExplorerStore.createPivot(
      AD_BIDS_EXPLORE_NAME,
      [dim("publisher")],
      [measure("impressions")],
    );
    metricsExplorerStore.addPivotField(
      AD_BIDS_EXPLORE_NAME,
      dim("publisher"),
      true,
    );
    metricsExplorerStore.addPivotField(
      AD_BIDS_EXPLORE_NAME,
      dim("publisher"),
      false,
    );
    metricsExplorerStore.addPivotField(
      AD_BIDS_EXPLORE_NAME,
      measure("impressions"),
      false,
    );
    expect(pivot().rows).toEqual([dim("publisher")]);
    expect(pivot().columns).toEqual([measure("impressions")]);

    metricsExplorerStore.addPivotField(
      AD_BIDS_EXPLORE_NAME,
      dim("domain"),
      true,
    );
    expect(pivot().rows).toEqual([dim("publisher"), dim("domain")]);
  });

  it("switching to a flat table keeps a field in both zones once", () => {
    // The columns a nest to flat switch builds: column dimensions, then row
    // dimensions, then measures. A dimension in both zones appears twice.
    metricsExplorerStore.setPivotTableMode(
      AD_BIDS_EXPLORE_NAME,
      "flat",
      [],
      [dim("publisher"), dim("publisher"), measure("impressions")],
    );
    expect(pivot().tableMode).toEqual("flat");
    expect(pivot().columns).toEqual([dim("publisher"), measure("impressions")]);
  });

  it("createPivot keeps a repeated field once", () => {
    metricsExplorerStore.createPivot(
      AD_BIDS_EXPLORE_NAME,
      [dim("publisher"), dim("publisher")],
      [measure("impressions"), measure("impressions")],
    );
    expect(pivot().rows).toEqual([dim("publisher")]);
    expect(pivot().columns).toEqual([measure("impressions")]);
  });
});

describe("sync", () => {
  const measureChip = (id: string) => ({
    id,
    title: id,
    type: PivotChipType.Measure,
  });

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
    metricsExplorerStore.createPivot(
      AD_BIDS_EXPLORE_NAME,
      [
        {
          id: AD_BIDS_PUBLISHER_DIMENSION,
          title: AD_BIDS_PUBLISHER_DIMENSION,
          type: PivotChipType.Dimension,
        },
      ],
      [
        measureChip(AD_BIDS_IMPRESSIONS_MEASURE),
        measureChip(AD_BIDS_BID_PRICE_MEASURE),
      ],
    );
  });

  function syncWithMeasures(measures: string[]) {
    metricsExplorerStore.sync(AD_BIDS_EXPLORE_NAME, {
      ...AD_BIDS_EXPLORE_INIT,
      measures,
    });
    return get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME].pivot;
  }

  function syncWithoutBidPrice() {
    return syncWithMeasures([AD_BIDS_IMPRESSIONS_MEASURE]);
  }

  // The editor preview syncs the state with every spec edit; the pivot's
  // queries fail when the sort names a measure the spec no longer has.
  it("drops the pivot sort together with its measure", () => {
    metricsExplorerStore.setPivotSort(AD_BIDS_EXPLORE_NAME, [
      { id: AD_BIDS_BID_PRICE_MEASURE, desc: true },
    ]);

    const pivot = syncWithoutBidPrice();
    expect(pivot.columns.map((c) => c.id)).toEqual([
      AD_BIDS_IMPRESSIONS_MEASURE,
    ]);
    expect(pivot.sorting).toEqual([]);
  });

  it("keeps a pivot sort on a measure the spec still has", () => {
    metricsExplorerStore.setPivotSort(AD_BIDS_EXPLORE_NAME, [
      { id: AD_BIDS_IMPRESSIONS_MEASURE, desc: true },
    ]);

    const pivot = syncWithoutBidPrice();
    expect(pivot.sorting).toEqual([
      { id: AD_BIDS_IMPRESSIONS_MEASURE, desc: true },
    ]);
  });

  // A sort on a measure under one column value points at the measure by
  // position, so it would silently move to another measure once one is
  // removed: `m0` meant impressions and would mean bid price.
  it("drops a positional sort once a column chip is removed", () => {
    metricsExplorerStore.setPivotSort(AD_BIDS_EXPLORE_NAME, [
      { id: "m0", desc: true },
    ]);

    const pivot = syncWithMeasures([AD_BIDS_BID_PRICE_MEASURE]);
    expect(pivot.columns.map((c) => c.id)).toEqual([AD_BIDS_BID_PRICE_MEASURE]);
    expect(pivot.sorting).toEqual([]);
  });

  it("keeps a positional sort while the columns are intact", () => {
    metricsExplorerStore.setPivotSort(AD_BIDS_EXPLORE_NAME, [
      { id: "m0", desc: true },
    ]);

    const pivot = syncWithMeasures([
      AD_BIDS_IMPRESSIONS_MEASURE,
      AD_BIDS_BID_PRICE_MEASURE,
    ]);
    expect(pivot.sorting).toEqual([{ id: "m0", desc: true }]);
  });
});
