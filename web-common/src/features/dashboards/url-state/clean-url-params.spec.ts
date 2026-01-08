import { describe, it, expect } from "vitest";
import { cleanUrlParams } from "web-common/src/features/dashboards/url-state/clean-url-params";

const TestCases: {
  title: string;
  defaultSearch: string;
  search: string;
  expectedCleanedSearch: string;
}[] = [
  {
    title: "should remove default params for explore and tdd",
    defaultSearch:
      "view=explore&tr=PT6H&tz=UTC&compare_tr=&grain=hour&compare_dim=",
    search:
      "view=tdd&tr=P7D&tz=UTC&compare_tr=&grain=hour&compare_dim=&f=&measure=impressions&chart_type=line",
    expectedCleanedSearch:
      "view=tdd&tr=P7D&grain=hour&measure=impressions&chart_type=line",
  },
  {
    title: "should remove default params for explore and pivot",
    defaultSearch:
      "view=explore&tr=PT6H&tz=UTC&compare_tr=&grain=hour&compare_dim=",
    search:
      "view=pivot&tr=P7D&tz=UTC&compare_tr=&f=&rows=publisher&cols=impressions&sort_by=impressions&sort_dir=DESC&table_mode=nest",
    expectedCleanedSearch:
      "view=pivot&tr=P7D&rows=publisher&cols=impressions&sort_by=impressions&sort_dir=DESC&table_mode=nest",
  },

  {
    title: "should remove default params for explore and pivot in flat mode",
    defaultSearch:
      "view=explore&tr=PT6H&tz=UTC&compare_tr=&grain=hour&compare_dim=",
    search:
      "view=pivot&tr=P7D&tz=UTC&compare_tr=&f=&rows=&cols=publisher%2Cimpressions&sort_by=impressions&sort_dir=DESC&table_mode=flat",
    expectedCleanedSearch:
      "view=pivot&tr=P7D&cols=publisher%2Cimpressions&sort_by=impressions&sort_dir=DESC&table_mode=flat",
  },
];

describe("clean-url-params", () => {
  for (const {
    title,
    defaultSearch,
    search,
    expectedCleanedSearch,
  } of TestCases) {
    it(title, () => {
      const defaultParams = new URLSearchParams(defaultSearch);
      const params = new URLSearchParams(search);

      const cleanedParams = cleanUrlParams(params, defaultParams);
      expect(cleanedParams.toString()).toEqual(expectedCleanedSearch);
    });
  }
});
