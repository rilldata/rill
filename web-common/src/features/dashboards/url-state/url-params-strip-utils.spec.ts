import { stripDefaultOrEmptyUrlParams } from "@rilldata/web-common/features/dashboards/url-state/url-params-strip-utils";
import { describe, it, expect } from "vitest";

const TestCases: {
  title: string;
  defaultSearch: string;
  search: string;
  expectedStrippedSearch: string;
}[] = [
  {
    title: "should remove default params for explore views",
    defaultSearch:
      "view=explore&tr=PT6H&tz=UTC&compare_tr=&grain=hour&compare_dim=",
    search: "view=explore&tr=P7D&tz=UTC&compare_tr=&grain=day&compare_dim=",
    expectedStrippedSearch: "tr=P7D&grain=day",
  },
  {
    title: "should keep unknown params for explore views",
    defaultSearch:
      "view=explore&tr=PT6H&tz=UTC&compare_tr=&grain=hour&compare_dim=",
    search:
      "view=explore&tr=P7D&tz=UTC&compare_tr=&grain=day&compare_dim=&unknown=abc",
    expectedStrippedSearch: "tr=P7D&grain=day&unknown=abc",
  },

  {
    title: "should remove default params for explore and tdd",
    defaultSearch:
      "view=explore&tr=PT6H&tz=UTC&compare_tr=&grain=hour&compare_dim=",
    search:
      "view=tdd&tr=P7D&tz=UTC&compare_tr=&grain=hour&compare_dim=&measure=impressions",
    expectedStrippedSearch: "view=tdd&tr=P7D&measure=impressions",
  },
  {
    title: "should remove default params for explore and pivot",
    defaultSearch:
      "view=explore&tr=PT6H&tz=UTC&compare_tr=&grain=hour&compare_dim=&sort_by=impressions&sort_dir=DESC",
    search:
      "view=pivot&tr=P7D&tz=UTC&compare_tr=&rows=publisher&cols=impressions&sort_by=impressions&sort_dir=DESC",
    expectedStrippedSearch:
      "view=pivot&tr=P7D&rows=publisher&cols=impressions&sort_by=impressions&sort_dir=DESC",
  },

  {
    title: "should remove default params for explore and pivot in flat mode",
    defaultSearch:
      "view=explore&tr=PT6H&tz=UTC&compare_tr=&grain=hour&compare_dim=&sort_by=impressions&sort_dir=DESC",
    search:
      "view=pivot&tr=P7D&tz=UTC&compare_tr=&rows=&cols=publisher%2Cimpressions&sort_by=impressions&sort_dir=DESC&table_mode=flat",
    expectedStrippedSearch:
      "view=pivot&tr=P7D&cols=publisher%2Cimpressions&sort_by=impressions&sort_dir=DESC&table_mode=flat",
  },
];

describe("url-params-strip-utils", () => {
  for (const {
    title,
    defaultSearch,
    search,
    expectedStrippedSearch,
  } of TestCases) {
    it(title, () => {
      const defaultParams = new URLSearchParams(defaultSearch);
      const params = new URLSearchParams(search);

      const stripedParams = stripDefaultOrEmptyUrlParams(params, defaultParams);
      expect(stripedParams.toString()).toEqual(expectedStrippedSearch);
    });
  }
});
