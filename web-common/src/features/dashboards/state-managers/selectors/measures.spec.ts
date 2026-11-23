import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import {
  filterOutSomeAdvancedAggregationMeasures,
  filterOutSomeAdvancedMeasures,
  getDimensionQueryMeasuresAndSort,
  measureSupportsTotalsQuery,
} from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  type MetricsViewSpecMeasure,
  type V1MetricsViewAggregationMeasure,
  type V1MetricsViewSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { describe, it, expect } from "vitest";

describe("measures selectors", () => {
  const SimpleMeasure: MetricsViewSpecMeasure = {
    name: "mes",
    expression: "count(*)",
  };
  const TimeNoGrainMeasure: MetricsViewSpecMeasure = {
    name: "mes_time_no_grain",
    requiredDimensions: [
      {
        name: "time",
        timeGrain: V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      },
    ],
  };
  const TimeDayGrainMeasure: MetricsViewSpecMeasure = {
    name: "mes_time_day_grain",
    requiredDimensions: [
      {
        name: "time",
        timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
      },
    ],
  };
  const TimeWeekGrainMeasure: MetricsViewSpecMeasure = {
    name: "mes_time_week_grain",
    requiredDimensions: [
      {
        name: "time",
        timeGrain: V1TimeGrain.TIME_GRAIN_WEEK,
      },
    ],
  };
  // Requires a categorical dimension, e.g. `requires: [domain]`.
  const DomainMeasure: MetricsViewSpecMeasure = {
    name: "mes_domain",
    requiredDimensions: [{ name: "domain" }],
  };
  const WindowMeasure: MetricsViewSpecMeasure = {
    name: "window_mes",
    window: {
      partition: true,
    },
  };
  // A rolling window ordered by the time dimension, which the parser also adds to the required dimensions.
  const WindowTimeMeasure: MetricsViewSpecMeasure = {
    name: "window_time_mes",
    window: {
      orderBy: [{ name: "time" }],
      frameExpression: "RANGE BETWEEN INTERVAL 6 DAY PRECEDING AND CURRENT ROW",
    },
    requiredDimensions: [
      {
        name: "time",
        timeGrain: V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      },
    ],
  };
  const MetricsView: V1MetricsViewSpec = {
    timeDimension: "time",
    measures: [
      SimpleMeasure,
      TimeNoGrainMeasure,
      TimeDayGrainMeasure,
      TimeWeekGrainMeasure,
      DomainMeasure,
      WindowMeasure,
      WindowTimeMeasure,
    ],
  };
  const AllMeasureNames = MetricsView.measures!.map((m) => m.name!);

  describe("filterOutSomeAdvancedMeasures for queries grouped by the time dimension", () => {
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
      {
        title:
          "with a rolling window on the time dimension and a measure requiring another dimension",
        measures: ["mes", "window_time_mes", "mes_domain"],
        timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
        expectedMeasures: ["mes", "window_time_mes"],
      },
    ];
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
            { time: true },
          ),
        ).toEqual(expectedMeasures);
      });
    }

    it("excludes measures requiring a time dimension other than the selected one", () => {
      expect(
        filterOutSomeAdvancedMeasures(
          {
            selectedTimeRange: {
              interval: V1TimeGrain.TIME_GRAIN_DAY,
            },
            selectedTimeDimension: "other_time",
          } as ExploreState,
          MetricsView,
          ["mes", "mes_time_no_grain", "window_time_mes"],
          { time: true },
        ),
      ).toEqual(["mes"]);
    });
  });

  describe("filterOutSomeAdvancedMeasures for queries without the time dimension", () => {
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
          {},
        ),
      ).toEqual(["mes"]);
    });

    it("excludes every measure with required dimensions, whatever their grain", () => {
      expect(
        filterOutSomeAdvancedMeasures(
          {
            selectedTimeRange: {
              interval: V1TimeGrain.TIME_GRAIN_DAY,
            },
          } as ExploreState,
          MetricsView,
          AllMeasureNames,
          {},
        ),
      ).toEqual(["mes"]);
    });
  });

  describe("filterOutSomeAdvancedAggregationMeasures for queries without the time dimension", () => {
    it("excludes measures with required dimensions and the comparison measures derived from them", () => {
      const measures: V1MetricsViewAggregationMeasure[] = [
        { name: "mes" },
        { name: "mes_time_no_grain" },
        { name: "mes_domain" },
        { name: "window_time_mes" },
        { name: "mes__previous", comparisonValue: { measure: "mes" } },
        {
          name: "mes_time_no_grain__previous",
          comparisonValue: { measure: "mes_time_no_grain" },
        },
        {
          name: "mes_domain__delta_abs",
          comparisonDelta: { measure: "mes_domain" },
        },
      ];
      expect(
        filterOutSomeAdvancedAggregationMeasures(
          {
            selectedTimeRange: {
              interval: V1TimeGrain.TIME_GRAIN_DAY,
            },
          } as ExploreState,
          MetricsView,
          measures,
          {},
        ).map((m) => m.name),
      ).toEqual(["mes", "mes__previous"]);
    });
  });

  describe("filterOutSomeAdvancedAggregationMeasures for queries grouped by a categorical dimension", () => {
    it("keeps measures that require only that dimension", () => {
      expect(
        filterOutSomeAdvancedAggregationMeasures(
          {
            selectedTimeRange: {
              interval: V1TimeGrain.TIME_GRAIN_DAY,
            },
          } as ExploreState,
          MetricsView,
          MetricsView.measures!,
          { dimensions: ["domain"] },
        ).map((m) => m.name),
      ).toEqual(["mes", "mes_domain"]);
    });
  });

  describe("getDimensionQueryMeasuresAndSort", () => {
    const exploreState = {
      selectedTimeRange: { interval: V1TimeGrain.TIME_GRAIN_DAY },
      sortDirection: SortDirection.DESCENDING,
    } as ExploreState;

    it("keeps a simple measure and sorts by it", () => {
      expect(
        getDimensionQueryMeasuresAndSort(
          exploreState,
          MetricsView,
          [{ name: "mes" }],
          "publisher",
          "mes",
        ),
      ).toEqual({
        measures: [{ name: "mes" }],
        sort: [{ desc: true, name: "mes" }],
      });
    });

    it("keeps a measure that requires only the queried dimension and sorts by it", () => {
      expect(
        getDimensionQueryMeasuresAndSort(
          exploreState,
          MetricsView,
          [{ name: "mes" }, { name: "mes_domain" }],
          "domain",
          "mes_domain",
        ),
      ).toEqual({
        measures: [{ name: "mes" }, { name: "mes_domain" }],
        sort: [{ desc: true, name: "mes_domain" }],
      });
    });

    it("omits a measure that needs another dimension and sorts by the dimension instead", () => {
      expect(
        getDimensionQueryMeasuresAndSort(
          exploreState,
          MetricsView,
          [{ name: "mes" }, { name: "mes_time_no_grain" }],
          "publisher",
          "mes_time_no_grain",
        ),
      ).toEqual({
        measures: [{ name: "mes" }],
        sort: [{ desc: true, name: "publisher" }],
      });
    });

    it("keeps sorting by the dimension when that is selected", () => {
      expect(
        getDimensionQueryMeasuresAndSort(
          { ...exploreState, sortDirection: SortDirection.ASCENDING },
          MetricsView,
          [{ name: "mes" }],
          "publisher",
          "publisher",
        ),
      ).toEqual({
        measures: [{ name: "mes" }],
        sort: [{ desc: false, name: "publisher" }],
      });
    });

    it("keeps an ephemeral measure, which has no spec entry", () => {
      expect(
        getDimensionQueryMeasuresAndSort(
          {
            ...exploreState,
            ephemeralMeasures: [{ name: "mes_doubled" }],
          } as ExploreState,
          MetricsView,
          [{ name: "mes_doubled" }],
          "publisher",
          "mes_doubled",
        ),
      ).toEqual({
        measures: [{ name: "mes_doubled" }],
        sort: [{ desc: true, name: "mes_doubled" }],
      });
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
      expect(measureSupportsTotalsQuery(TimeNoGrainMeasure)).toBe(false);
    });

    it("does not support a measure with a required categorical dimension", () => {
      expect(measureSupportsTotalsQuery(DomainMeasure)).toBe(false);
    });
  });
});
