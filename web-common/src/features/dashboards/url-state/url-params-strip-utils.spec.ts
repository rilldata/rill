import { describe, it, expect } from "vitest";
import { stripDefaultOrEmptyUrlParams } from "./url-params-strip-utils";

const TestCases: {
  title: string;
  defaultSearch: string;
  search: string;
  expectedStrippedSearch: string;
}[] = [
  {
    title: "should remove default params for explore and tdd",
    defaultSearch:
      "view=explore&tr=PT6H&tz=UTC&compare_tr=&grain=hour&compare_dim=",
    search:
      "view=tdd&tr=P7D&tz=UTC&compare_tr=&grain=hour&compare_dim=&f=&measure=impressions&chart_type=timeseries",
    expectedStrippedSearch:
      "view=tdd&tr=P7D&measure=impressions&chart_type=timeseries",
  },
  {
    title: "should remove default params for explore and pivot",
    defaultSearch:
      "view=explore&tr=PT6H&tz=UTC&compare_tr=&grain=hour&compare_dim=",
    search:
      "view=pivot&tr=P7D&tz=UTC&compare_tr=&f=&rows=publisher&cols=impressions&sort_by=impressions&sort_dir=DESC&table_mode=nest",
    expectedStrippedSearch:
      "view=pivot&tr=P7D&rows=publisher&cols=impressions&sort_by=impressions&sort_dir=DESC&table_mode=nest",
  },

  {
    title: "should remove default params for explore and pivot in flat mode",
    defaultSearch:
      "view=explore&tr=PT6H&tz=UTC&compare_tr=&grain=hour&compare_dim=",
    search:
      "view=pivot&tr=P7D&tz=UTC&compare_tr=&f=&rows=&cols=publisher%2Cimpressions&sort_by=impressions&sort_dir=DESC&table_mode=flat",
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
