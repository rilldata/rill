import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
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

  describe("explore page", () => {
    it("keeps every definition when all measures are visible", () => {
      expect(
        referencedEphemeralMeasures({
          ephemeralMeasures: [profit, arpu],
          allMeasuresVisible: true,
          visibleMeasures: ["revenue"],
        }),
      ).toEqual([profit, arpu]);
    });

    it("keeps visible, leaderboard and sort measures only", () => {
      const state: Partial<ExploreState> = {
        ephemeralMeasures: [profit, arpu, margin],
        allMeasuresVisible: false,
        visibleMeasures: ["revenue", "profit"],
        leaderboardMeasureNames: ["revenue"],
        leaderboardSortByMeasureName: "revenue",
      };
      expect(referencedEphemeralMeasures(state)).toEqual([profit]);
      expect(
        referencedEphemeralMeasures({
          ...state,
          visibleMeasures: [],
          leaderboardMeasureNames: ["arpu"],
          leaderboardSortByMeasureName: "margin",
        }),
      ).toEqual([arpu, margin]);
    });

    it("ignores pivot chips and the TDD measure", () => {
      expect(
        referencedEphemeralMeasures({
          activePage: DashboardState_ActivePage.DEFAULT,
          ephemeralMeasures: [profit, arpu, margin],
          allMeasuresVisible: false,
          visibleMeasures: ["profit"],
          tdd: {
            expandedMeasureName: "arpu",
            chartType: "" as never,
            pinIndex: -1,
          },
          pivot: {
            columns: [
              { id: "margin", title: "Margin", type: PivotChipType.Measure },
            ],
            rows: [],
            sorting: [],
          } as never,
        }),
      ).toEqual([profit]);
    });
  });

  it("keeps only the expanded measure on the TDD page", () => {
    expect(
      referencedEphemeralMeasures({
        activePage: DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
        ephemeralMeasures: [profit, arpu, margin],
        allMeasuresVisible: true,
        visibleMeasures: ["profit", "margin"],
        tdd: {
          expandedMeasureName: "arpu",
          chartType: "" as never,
          pinIndex: -1,
        },
      }),
    ).toEqual([arpu]);
  });

  it("keeps only row, column and sort chips on the pivot page", () => {
    const state: Partial<ExploreState> = {
      activePage: DashboardState_ActivePage.PIVOT,
      ephemeralMeasures: [profit, arpu, margin],
      allMeasuresVisible: true,
      visibleMeasures: ["profit", "arpu", "margin"],
      leaderboardMeasureNames: ["arpu"],
      pivot: {
        columns: [
          { id: "margin", title: "Margin", type: PivotChipType.Measure },
        ],
        rows: [],
        sorting: [{ id: "profit", desc: true }],
      } as never,
    };
    expect(referencedEphemeralMeasures(state)).toEqual([profit, margin]);
    expect(
      referencedEphemeralMeasures({
        ...state,
        pivot: { columns: [], rows: [], sorting: [] } as never,
      }),
    ).toEqual([]);
  });
});
