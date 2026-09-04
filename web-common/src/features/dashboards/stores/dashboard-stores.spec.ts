import { getUpdatedPinIndex } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  createAndExpression,
  createInExpression,
  createLikeExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { describe, expect, it } from "vitest";

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
