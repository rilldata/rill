import { render } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import LeaderboardHeader from "./LeaderboardHeader.svelte";
import { SortType } from "../proto-state/derived-types";

const baseProps = {
  dimensionName: "country",
  displayName: "Country",
  dimensionDescription: "",
  isFetching: false,
  isTimeComparisonActive: true,
  isBeingCompared: false,
  sortedAscending: false,
  hovered: false,
  // Any sort type other than the ones rendering a sort arrow, which needs the
  // Web Animations API that jsdom does not implement.
  sortType: SortType.DIMENSION,
  allowDimensionComparison: true,
  allowExpandTable: true,
  leaderboardMeasureNames: ["impressions", "bids"],
  leaderboardSortByMeasureName: "impressions",
  isValidPercentOfTotal: () => false,
  measureLabel: (measureName: string) => measureName,
  toggleSort: () => {},
  setPrimaryDimension: () => {},
  toggleComparisonDimension: () => {},
  dimensionColumnWidth: 164,
};

describe("LeaderboardHeader", () => {
  it("adds context columns for every measure when the set of measures with context changes", async () => {
    const { container, rerender } = render(LeaderboardHeader, {
      props: { ...baseProps, measuresWithContext: new Set(["impressions"]) },
    });

    expect(
      container.querySelectorAll("th[data-absolute-change-header]").length,
    ).toBe(1);
    expect(
      container.querySelectorAll("th[data-percent-change-header]").length,
    ).toBe(1);

    await rerender({
      ...baseProps,
      measuresWithContext: new Set(["impressions", "bids"]),
    });

    expect(
      container.querySelectorAll("th[data-absolute-change-header]").length,
    ).toBe(2);
    expect(
      container.querySelectorAll("th[data-percent-change-header]").length,
    ).toBe(2);
  });
});
