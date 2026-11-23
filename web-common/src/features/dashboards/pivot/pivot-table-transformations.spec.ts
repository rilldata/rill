import { describe, expect, it } from "vitest";
import { mergeRowTotalsInOrder } from "./pivot-table-transformations";

describe("mergeRowTotalsInOrder", () => {
  // Rows sorted by a measure come from a sort-only query that orders them but
  // cannot carry URI measures. The rows displayed are the measure totals
  // query's rows, in the sort-only query's order.
  const sortedRowTotals = [
    { country: "Spain", measure: 30 },
    { country: "France", measure: 20 },
    { country: "Italy", measure: 10 },
  ];
  const [france, italy, spain] = [
    { country: "France", measure: 20, country__rill_uri: "https://x.io/fr" },
    { country: "Italy", measure: 10, country__rill_uri: "https://x.io/it" },
    { country: "Spain", measure: 30, country__rill_uri: "https://x.io/es" },
  ];

  it("returns the totals rows in the sorted order, URI measures included", () => {
    expect(
      mergeRowTotalsInOrder(
        ["Spain", "France", "Italy"],
        sortedRowTotals,
        ["France", "Italy", "Spain"],
        [france, italy, spain],
      ),
    ).toEqual([spain, france, italy]);
  });

  it("drops sorted rows that the totals query did not return", () => {
    expect(
      mergeRowTotalsInOrder(
        ["Spain", "France", "Italy"],
        sortedRowTotals,
        ["France", "Spain"],
        [france, spain],
      ),
    ).toEqual([spain, france]);
  });

  it("falls back to the sorted totals without a totals query", () => {
    expect(
      mergeRowTotalsInOrder(
        ["Spain", "France", "Italy"],
        sortedRowTotals,
        [],
        [],
      ),
    ).toBe(sortedRowTotals);
  });
});
