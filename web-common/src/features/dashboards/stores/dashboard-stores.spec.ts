import {
  type PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";
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
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_INIT,
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
