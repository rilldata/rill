import { mergeFilterParams } from "@rilldata/web-common/features/dashboards/filters/expr-utils.ts";
import {
  createAndExpression,
  createInExpression,
  createOrExpression,
  createSubQueryExpression,
  createBinaryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

const countryUS = createInExpression("country", ["US"]);
const countryUSCA = createInExpression("country", ["US", "CA"]);
const countryCAUS = createInExpression("country", ["CA", "US"]);
const publisherYahoo = createInExpression("publisher", ["Yahoo"]);
const impressionsGt10 = createSubQueryExpression(
  "country",
  ["impressions"],
  createBinaryExpression("impressions", V1Operation.OPERATION_GT, 10),
);

function toParam(expr: V1Expression) {
  return convertExpressionToFilterParam(expr);
}

describe("mergeFilterParams", () => {
  it("returns an empty expression when there is nothing to merge", () => {
    expect(mergeFilterParams(new URLSearchParams())).toEqual({
      expr: createAndExpression([]),
      inList: [],
      advanced: false,
    });
    expect(mergeFilterParams(new URLSearchParams("mv1=&mv2="))).toEqual({
      expr: createAndExpression([]),
      inList: [],
      advanced: false,
    });
  });

  it("drops a malformed param", () => {
    const merged = mergeFilterParams(
      new URLSearchParams(`mv1=country IN (&mv2=${toParam(countryUS)}`),
    );
    expect(merged.expr).toEqual(createAndExpression([countryUS]));
    expect(merged.advanced).toBe(false);
  });

  it("unions conditions across metrics views", () => {
    const merged = mergeFilterParams(
      new URLSearchParams([
        ["mv1", toParam(createAndExpression([countryUS]))],
        ["mv2", toParam(createAndExpression([publisherYahoo]))],
      ]),
    );
    expect(merged.expr).toEqual(
      createAndExpression([countryUS, publisherYahoo]),
    );
    expect(merged.advanced).toBe(false);
  });

  it("keeps a repeated condition once, regardless of value order", () => {
    const merged = mergeFilterParams(
      new URLSearchParams([
        ["mv1", toParam(createAndExpression([countryUSCA, publisherYahoo]))],
        ["mv2", toParam(createAndExpression([countryCAUS]))],
      ]),
    );
    expect(merged.expr).toEqual(
      createAndExpression([countryUSCA, publisherYahoo]),
    );
    expect(merged.advanced).toBe(false);
  });

  it("unwraps a param that is a single condition", () => {
    const merged = mergeFilterParams(
      new URLSearchParams([
        ["mv1", toParam(countryUS)],
        ["mv2", toParam(createAndExpression([countryUS, publisherYahoo]))],
      ]),
    );
    expect(merged.expr).toEqual(
      createAndExpression([countryUS, publisherYahoo]),
    );
    expect(merged.advanced).toBe(false);
  });

  it("merges measure filters", () => {
    const merged = mergeFilterParams(
      new URLSearchParams([
        ["mv1", toParam(createAndExpression([impressionsGt10]))],
        ["mv2", toParam(createAndExpression([impressionsGt10]))],
      ]),
    );
    expect(merged.expr).toEqual(createAndExpression([impressionsGt10]));
    expect(merged.advanced).toBe(false);
  });

  it("is advanced when an identifier is filtered two different ways", () => {
    const merged = mergeFilterParams(
      new URLSearchParams([
        ["mv1", toParam(createAndExpression([countryUS]))],
        ["mv2", toParam(createAndExpression([countryUSCA]))],
      ]),
    );
    expect(merged.expr).toEqual(createAndExpression([countryUS, countryUSCA]));
    expect(merged.advanced).toBe(true);
  });

  it("is advanced when a measure is filtered two different ways", () => {
    const impressionsGt20 = createSubQueryExpression(
      "country",
      ["impressions"],
      createBinaryExpression("impressions", V1Operation.OPERATION_GT, 20),
    );
    const merged = mergeFilterParams(
      new URLSearchParams([
        ["mv1", toParam(createAndExpression([impressionsGt10]))],
        ["mv2", toParam(createAndExpression([impressionsGt20]))],
      ]),
    );
    expect(merged.advanced).toBe(true);
  });

  it("is advanced when a condition is a nested AND/OR", () => {
    const nested = createOrExpression([countryUS, publisherYahoo]);
    const merged = mergeFilterParams(
      new URLSearchParams([["mv1", toParam(createAndExpression([nested]))]]),
    );
    expect(merged.expr).toEqual(createAndExpression([nested]));
    expect(merged.advanced).toBe(true);
  });

  it("keeps a top level OR whole", () => {
    const topLevelOr = createOrExpression([countryUS, publisherYahoo]);
    const merged = mergeFilterParams(
      new URLSearchParams([
        ["mv1", toParam(topLevelOr)],
        ["mv2", toParam(createAndExpression([countryUS]))],
      ]),
    );
    expect(merged.expr).toEqual(createAndExpression([topLevelOr, countryUS]));
    expect(merged.advanced).toBe(true);
  });

  it("unions the in list dimensions", () => {
    const merged = mergeFilterParams(
      new URLSearchParams([
        [
          "mv1",
          convertExpressionToFilterParam(createAndExpression([countryUSCA]), [
            "country",
          ]),
        ],
        [
          "mv2",
          convertExpressionToFilterParam(
            createAndExpression([countryUSCA, publisherYahoo]),
            ["country", "publisher"],
          ),
        ],
      ]),
    );
    expect(merged.inList).toEqual(["country", "publisher"]);
    expect(merged.advanced).toBe(false);
  });
});
