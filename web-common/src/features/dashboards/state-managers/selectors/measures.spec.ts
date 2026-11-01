import {
  filterOutSomeAdvancedMeasures,
  measureSupportsTotalsQuery,
} from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  type V1MetricsViewSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { describe, it, expect } from "vitest";

describe("measures selectors", () => {
  describe("getFilteredMeasuresAndDimensions", () => {
    const TestCases = [
      {
        title: "with unspecified grains, selected DAY",
        measures: ["mes", "mes_time_no_grain"],
        timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
        expectedMeasures: ["mes", "mes_time_no_grain"],
      },
      {
        title: "with unspecified and specified grains, selected DAY",
        measures: [
          "mes",
          "mes_time_no_grain",
          "mes_time_day_grain",
          "mes_time_week_grain",
        ],
        timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
        expectedMeasures: ["mes", "mes_time_no_grain", "mes_time_day_grain"],
      },
      {
        title: "with unspecified and specified grains, selected WEEK",
        measures: [
          "mes",
          "mes_time_no_grain",
          "mes_time_day_grain",
          "mes_time_week_grain",
        ],
        timeGrain: V1TimeGrain.TIME_GRAIN_WEEK,
        expectedMeasures: ["mes", "mes_time_no_grain", "mes_time_week_grain"],
      },
      {
        title: "with unspecified and specified grains, selected MONTH",
        measures: [
          "mes",
          "mes_time_no_grain",
          "mes_time_day_grain",
          "mes_time_week_grain",
        ],
        timeGrain: V1TimeGrain.TIME_GRAIN_MONTH,
        expectedMeasures: ["mes", "mes_time_no_grain"],
      },
      {
        title: "with window measure and select it",
        measures: ["mes", "window_mes"],
        timeGrain: V1TimeGrain.TIME_GRAIN_MONTH,
        expectedMeasures: ["mes", "window_mes"],
      },
    ];
    const MetricsView: V1MetricsViewSpec = {
      measures: [
        {
          name: "mes",
          expression: "count(*)",
        },
        {
          name: "mes_time_no_grain",
          requiredDimensions: [
            {
              name: "time",
              timeGrain: V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
            },
          ],
        },
        {
          name: "mes_time_day_grain",
          requiredDimensions: [
            {
              name: "time",
              timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
            },
          ],
        },
        {
          name: "mes_time_week_grain",
          requiredDimensions: [
            {
              name: "time",
              timeGrain: V1TimeGrain.TIME_GRAIN_WEEK,
            },
          ],
        },
        {
          name: "window_mes",
          window: {
            partition: true,
          },
        },
      ],
    };
    for (const { title, measures, timeGrain, expectedMeasures } of TestCases) {
      it(title, () => {
        expect(
          filterOutSomeAdvancedMeasures(
            {
              selectedTimeRange: {
                interval: timeGrain,
              },
            } as ExploreState,
            MetricsView,
            measures,
            true,
          ),
        ).toEqual(expectedMeasures);
      });
    }

    it("with window measure and do not select it", () => {
      expect(
        filterOutSomeAdvancedMeasures(
          {
            selectedTimeRange: {
              interval: V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
            },
          } as ExploreState,
          MetricsView,
          ["mes", "window_mes"],
          false,
        ),
      ).toEqual(["mes"]);
    });
  });

  describe("measureSupportsTotalsQuery", () => {
    it("supports a measure without requiredDimensions", () => {
      expect(measureSupportsTotalsQuery({ name: "mes" })).toBe(true);
    });

    it("supports a measure with empty requiredDimensions", () => {
      expect(
        measureSupportsTotalsQuery({ name: "mes", requiredDimensions: [] }),
      ).toBe(true);
    });

    it("does not support a measure with a required time dimension", () => {
      expect(
        measureSupportsTotalsQuery({
          name: "mes_time_no_grain",
          requiredDimensions: [
            { name: "time", timeGrain: V1TimeGrain.TIME_GRAIN_UNSPECIFIED },
          ],
        }),
      ).toBe(false);
    });

    it("does not support a measure with a required categorical dimension", () => {
      expect(
        measureSupportsTotalsQuery({
          name: "mes_per_dim",
          requiredDimensions: [{ name: "domain" }],
        }),
      ).toBe(false);
    });
  });
});
