import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
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

  function syncWithoutBidPrice() {
    metricsExplorerStore.sync(AD_BIDS_EXPLORE_NAME, {
      ...AD_BIDS_EXPLORE_INIT,
      measures: [AD_BIDS_IMPRESSIONS_MEASURE],
    });
    return get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME].pivot;
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
});
