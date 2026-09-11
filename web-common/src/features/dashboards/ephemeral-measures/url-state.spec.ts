import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { describe, expect, it } from "vitest";
import type { EphemeralMeasureDef } from "./types";
import { referencedEphemeralMeasures } from "./url-state";

const profit: EphemeralMeasureDef = {
  name: "profit",
  displayName: "Profit",
  expression: "revenue - cost",
};
const arpu: EphemeralMeasureDef = {
  name: "arpu",
  displayName: "ARPU",
  expression: "revenue / users",
};
const margin: EphemeralMeasureDef = {
  name: "margin",
  displayName: "Margin",
  expression: "profit / revenue",
};

describe("referencedEphemeralMeasures", () => {
  it("returns nothing without definitions", () => {
    expect(referencedEphemeralMeasures({})).toEqual([]);
    expect(
      referencedEphemeralMeasures({
        ephemeralMeasures: [],
        visibleMeasures: [],
      }),
    ).toEqual([]);
  });

  it("keeps every definition when all measures are visible", () => {
    expect(
      referencedEphemeralMeasures({
        ephemeralMeasures: [profit, arpu],
        allMeasuresVisible: true,
        visibleMeasures: ["revenue"],
      }),
    ).toEqual([profit, arpu]);
  });

  it("drops definitions no view references", () => {
    expect(
      referencedEphemeralMeasures({
        ephemeralMeasures: [profit, arpu, margin],
        allMeasuresVisible: false,
        visibleMeasures: ["revenue", "profit"],
        leaderboardMeasureNames: ["revenue"],
        leaderboardSortByMeasureName: "revenue",
      }),
    ).toEqual([profit]);
  });

  it("treats leaderboard, sort, TDD and pivot usages as references", () => {
    const state: Partial<ExploreState> = {
      ephemeralMeasures: [profit, arpu, margin],
      allMeasuresVisible: false,
      visibleMeasures: [],
      leaderboardMeasureNames: ["arpu"],
      leaderboardSortByMeasureName: "arpu",
    };
    expect(referencedEphemeralMeasures(state)).toEqual([arpu]);

    expect(
      referencedEphemeralMeasures({
        ...state,
        leaderboardMeasureNames: [],
        leaderboardSortByMeasureName: "margin",
      }),
    ).toEqual([margin]);

    expect(
      referencedEphemeralMeasures({
        ...state,
        leaderboardMeasureNames: [],
        leaderboardSortByMeasureName: "revenue",
        tdd: {
          expandedMeasureName: "profit",
          chartType: "" as never,
          pinIndex: -1,
        },
      }),
    ).toEqual([profit]);

    expect(
      referencedEphemeralMeasures({
        ...state,
        leaderboardMeasureNames: [],
        leaderboardSortByMeasureName: "revenue",
        pivot: {
          columns: [
            { id: "margin", title: "Margin", type: PivotChipType.Measure },
          ],
          rows: [],
          sorting: [{ id: "profit", desc: true }],
        } as never,
      }),
    ).toEqual([profit, margin]);
  });
});
