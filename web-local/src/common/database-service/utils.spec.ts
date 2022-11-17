import { describe, expect } from "@jest/globals";
import {
  getFilterFromMetricsViewFilters,
  getWhereClauseFromFilters,
} from "@rilldata/web-local/common/database-service/utils";
import type { MetricsViewRequestFilter } from "@rilldata/web-local/common/rill-developer-service/MetricsViewActions";

const NullOnlyFilter: MetricsViewRequestFilter = {
  include: [
    {
      name: "col0",
      like: [],
      in: [null],
    },
  ],
  exclude: [],
};
const EmptyFilter: MetricsViewRequestFilter = {
  include: [
    {
      name: "col0",
      like: [],
      in: [],
    },
  ],
  exclude: [],
};

// TODO: add more exhaustive tests when this is moved to go
describe("Database Utils", () => {
  describe("getFilterFromMetricsViewFilters", () => {
    it("null only filter", () => {
      expect(getFilterFromMetricsViewFilters(NullOnlyFilter)).toBe(
        `("col0" IS  NULL)`
      );
    });

    it("empty filters", () => {
      expect(getFilterFromMetricsViewFilters(EmptyFilter)).toBe("");
    });
  });

  describe("getWhereClauseFromFilters", () => {
    it("empty filters with timestamp range", () => {
      expect(
        getWhereClauseFromFilters(
          EmptyFilter,
          "ts",
          {
            start: "2022-01-01",
            end: "2022-03-31",
          },
          "WHERE"
        )
      ).toBe(
        `WHERE "ts" >= TIMESTAMP '2022-01-01T00:00:00.000Z' AND "ts" <= TIMESTAMP '2022-03-31T00:00:00.000Z'`
      );
    });

    it("empty filter without timestamp range", () => {
      expect(getWhereClauseFromFilters(EmptyFilter, "ts", {}, "WHERE")).toBe(
        ""
      );
    });
  });
});
