import { getExploreStateFromHomeBookmarkUrl } from "@rilldata/web-admin/features/bookmarks/selectors.ts";
import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data.ts";
import {
  V1Operation,
  type V1ExploreSpec,
} from "@rilldata/web-common/runtime-client";
import { describe, expect, it, vi } from "vitest";

vi.mock("$app/stores", () => ({ page: {} }));

describe("getExploreStateFromHomeBookmarkUrl", () => {
  // What the runtime parses from `defaults.filters: publisher = 'Contoso'`.
  const exploreWithDefaultFilter: V1ExploreSpec = {
    ...AD_BIDS_EXPLORE_INIT,
    defaultPreset: {
      where: createBinaryExpression(
        AD_BIDS_PUBLISHER_DIMENSION,
        V1Operation.OPERATION_EQ,
        "Contoso",
      ),
    },
  };
  const whereFilterOf = (urlSearch: string) =>
    getExploreStateFromHomeBookmarkUrl(
      urlSearch,
      AD_BIDS_METRICS_INIT,
      exploreWithDefaultFilter,
      AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
    ).whereFilter;

  it("opens a bookmark saved without a filter unfiltered, whatever the explore's default filter", () => {
    // Bookmarks saved today hold an empty filter param. Older ones hold none.
    expect(whereFilterOf("tr=P7D&f=")).toEqual(createAndExpression([]));
    expect(whereFilterOf("tr=P7D")).toEqual(createAndExpression([]));
  });

  it("keeps the filter the bookmark was saved with", () => {
    expect(whereFilterOf("tr=P7D&f=domain IN ('fabrikam.com')")).toEqual(
      createAndExpression([
        createInExpression(AD_BIDS_DOMAIN_DIMENSION, ["fabrikam.com"]),
      ]),
    );
  });
});
