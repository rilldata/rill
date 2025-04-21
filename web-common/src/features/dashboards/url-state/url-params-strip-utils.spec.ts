import {
  AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getRillDefaultExploreUrlParamsByView } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params";
import { describe, it, expect } from "vitest";
import {
  mergeDefaultUrlParams,
  stripDefaultOrEmptyUrlParams,
} from "./url-params-strip-utils";

const TestCases: {
  title: string;
  search: string;
  expectedStrippedSearch: string;
}[] = [
  {
    title: "should remove default params for explore and tdd",
    search:
      "view=tdd&tr=P7D&tz=UTC&compare_tr=&grain=hour&compare_dim=&f=&measure=impressions&chart_type=timeseries",
    expectedStrippedSearch:
      "view=tdd&tr=P7D&measure=impressions&chart_type=timeseries",
  },
  {
    title: "should remove default params for explore and pivot",
    search:
      "view=pivot&tr=P7D&tz=UTC&compare_tr=&f=&rows=publisher&cols=impressions&sort_by=impressions&sort_dir=DESC&table_mode=nest",
    expectedStrippedSearch:
      "view=pivot&tr=P7D&rows=publisher&cols=impressions&sort_by=impressions&sort_dir=DESC&table_mode=nest",
  },

  {
    title: "should remove default params for explore and pivot in flat mode",
    search:
      "view=pivot&tr=P7D&tz=UTC&compare_tr=&f=&rows=&cols=publisher%2Cimpressions&sort_by=impressions&sort_dir=DESC&table_mode=flat",
    expectedStrippedSearch:
      "view=pivot&tr=P7D&cols=publisher%2Cimpressions&sort_by=impressions&sort_dir=DESC&table_mode=flat",
  },
];

describe("url-params-strip-utils", () => {
  const rillDefaultExploreURLParamsByView =
    getRillDefaultExploreUrlParamsByView(
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
      AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
      AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
    );

  for (const { title, search, expectedStrippedSearch } of TestCases) {
    it(title, () => {
      const params = new URLSearchParams(search);

      const stripedParams = stripDefaultOrEmptyUrlParams(
        params,
        rillDefaultExploreURLParamsByView.explore,
      );
      expect(stripedParams.toString()).toEqual(expectedStrippedSearch);

      const mergedParams = mergeDefaultUrlParams(
        stripedParams,
        rillDefaultExploreURLParamsByView,
      );
      // Order is not maintained but they are equal ignoring order
      expect(mergedParams.size).toEqual(params.size);
      expect(
        [...mergedParams.entries()].every(
          ([value, key]) => value === params.get(key),
        ),
      );
    });
  }
});
