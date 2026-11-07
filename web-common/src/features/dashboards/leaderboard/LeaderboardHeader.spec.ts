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

  it("renders a precision chip only for measures with an exact variant", async () => {
    const toggled: string[] = [];
    const { container } = render(LeaderboardHeader, {
      props: {
        ...baseProps,
        measuresWithContext: new Set<string>(),
        precisionFor: (measureName: string) =>
          measureName === "impressions"
            ? {
                state: "approx" as const,
                measureLabel: "impressions",
                onToggle: () => toggled.push(measureName),
              }
            : undefined,
      },
    });

    const chips = container.querySelectorAll("button.precision-chip");
    expect(chips.length).toBe(1);
    expect(chips[0].textContent?.trim()).toBe("≈");
    expect(chips[0].getAttribute("aria-pressed")).toBe("false");

    (chips[0] as HTMLButtonElement).click();
    expect(toggled).toEqual(["impressions"]);
  });

  it("renders the chip beside the # header of a single-measure leaderboard", () => {
    const { container } = render(LeaderboardHeader, {
      props: {
        ...baseProps,
        leaderboardMeasureNames: ["impressions"],
        measuresWithContext: new Set<string>(),
        precisionFor: () => ({
          state: "exact" as const,
          measureLabel: "impressions",
          onToggle: () => {},
        }),
      },
    });

    const header = container.querySelector("th[data-measure-header]");
    expect(header?.textContent).toContain("#");
    const chip = header?.querySelector("button.precision-chip");
    expect(chip?.textContent?.trim()).toBe("=");
    expect(chip?.getAttribute("aria-pressed")).toBe("true");
  });
});
