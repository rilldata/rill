import { mergeFilters } from "@rilldata/web-common/features/dashboards/pivot/pivot-merge-filters";
import {
  createInExpression,
  createAndExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
import { describe, it, expect } from "vitest";

describe("mergeFilters", () => {
  it("merge in filters", () => {
    expect(
      convertExpressionToFilterParam(
        mergeFilters(
          createAndExpression([
            createInExpression("publisher", ["Facebook", "Yahoo"]),
            createInExpression("publisher", ["Google"], true),
          ]),
          createAndExpression([
            createInExpression("publisher", ["Facebook", "Microsoft"]),
            createInExpression("publisher", [null], true),
          ]),
        )!,
      ),
    ).toEqual(`publisher IN ('Facebook') AND publisher NIN ('Google',null)`);
  });

  it("merge filters without wrappers", () => {
    expect(
      convertExpressionToFilterParam(
        mergeFilters(
          createAndExpression([
            createInExpression("publisher", ["Facebook", "Microsoft"]),
          ]),
          createInExpression("publisher", ["Facebook", "Yahoo"]),
        )!,
      ),
    ).toEqual(`publisher IN ('Facebook')`);

    expect(
      convertExpressionToFilterParam(
        mergeFilters(
          createInExpression("publisher", ["Facebook", "Microsoft"]),
          createAndExpression([
            createInExpression("publisher", ["Facebook", "Yahoo"]),
          ]),
        )!,
      ),
    ).toEqual(`publisher IN ('Facebook')`);

    expect(
      convertExpressionToFilterParam(
        mergeFilters(
          createInExpression("publisher", ["Facebook", "Microsoft"]),
          createInExpression("publisher", ["Facebook", "Yahoo"]),
        )!,
      ),
    ).toEqual(`publisher IN ('Facebook')`);
  });
});
