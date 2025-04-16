import {
  mergeDefaultUrlParams,
  stripDefaultUrlParams,
} from "@rilldata/web-common/features/dashboards/url-state/url-params-strip-utils";
import { describe, it, expect } from "vitest";

const TestCases: {
  title: string;
  search: string;
  expectedStrippedSearch: string;
}[] = [
  {
    title: "should remove and add back default params",
    search: "view=explore&tr=P7D&tz=UTC&compare_tr=&grain=day&compare_dim=",
    expectedStrippedSearch: "tr=P7D&grain=day",
  },
  {
    title: "should keep unknown params",
    search:
      "view=explore&tr=P7D&tz=UTC&compare_tr=&grain=day&compare_dim=&unknown=abc",
    expectedStrippedSearch: "tr=P7D&grain=day&unknown=abc",
  },
];

describe("url-params-strip-utils", () => {
  const defaultParams = new URLSearchParams(
    "view=explore&tr=PT6H&tz=UTC&compare_tr=&grain=hour&compare_dim=",
  );

  for (const { title, search, expectedStrippedSearch } of TestCases) {
    it(title, () => {
      const params = new URLSearchParams(search);

      const stripedParams = stripDefaultUrlParams(params, defaultParams);
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
