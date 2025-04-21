import { describe, it, expect } from "vitest";
import {
  mergeDefaultUrlParams,
  stripDefaultOrEmptyUrlParams,
} from "./url-params-strip-utils";

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

      const mergedParams = mergeDefaultUrlParams(stripedParams, defaultParams);
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
