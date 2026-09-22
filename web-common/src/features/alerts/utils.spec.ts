import {
  type AlertFormValues,
  getAlertQueryArgsFromFormValues,
} from "@rilldata/web-common/features/alerts/form-utils";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
import { generateAlertName } from "@rilldata/web-common/features/alerts/utils";
import { ComparisonPercentOfTotal } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  MeasureFilterOperation,
  MeasureFilterType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types.ts";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

const MetricsView: V1MetricsViewSpec = {
  measures: [
    {
      name: "total_records",
      displayName: "Total records",
    },
  ],
};

describe("generateAlertName", () => {
  const TestCases: {
    title: string;
    formValues: Partial<AlertFormValues>;
    selectedComparisonTimeRange: DashboardTimeControls | undefined;
    expected: string;
  }[] = [
    {
      title: "value criteria without comparison",
      formValues: {
        measure: "total_records",
        criteria: [
          {
            measure: "total_records",
            type: MeasureFilterType.Value,
            operation: MeasureFilterOperation.LessThan,
            value1: "10",
            value2: "",
          },
        ],
      },
      selectedComparisonTimeRange: undefined,
      expected: "Total records value alert",
    },
    {
      title: "absolute change criteria with comparison",
      formValues: {
        measure: "total_records",
        criteria: [
          {
            measure: "total_records",
            type: MeasureFilterType.AbsoluteChange,
            operation: MeasureFilterOperation.GreaterThanOrEquals,
            value1: "10",
            value2: "",
          },
        ],
      },
      selectedComparisonTimeRange: {
        name: "rill-PW",
      } as DashboardTimeControls,
      expected: "Total records change vs previous week alert",
    },
    {
      title: "percent change criteria with comparison",
      formValues: {
        measure: "total_records",
        criteria: [
          {
            measure: "total_records",
            type: MeasureFilterType.PercentChange,
            operation: MeasureFilterOperation.GreaterThan,
            value1: "-10",
            value2: "",
          },
        ],
      },
      selectedComparisonTimeRange: {
        name: "rill-PM",
      } as DashboardTimeControls,
      expected: "Total records % change vs previous month alert",
    },
  ];

  for (const {
    title,
    formValues,
    selectedComparisonTimeRange,
    expected,
  } of TestCases) {
    it(title, () => {
      expect(
        generateAlertName(
          formValues as AlertFormValues,
          selectedComparisonTimeRange,
          MetricsView,
        ),
      ).toEqual(expected);
    });
  }
});

describe("getAlertQueryArgsFromFormValues", () => {
  const ephemeralMeasure = {
    name: "records_per_user",
    displayName: "Records per user",
    expression: "total_records / users",
  };
  const baseFormValues = {
    measure: ephemeralMeasure.name,
    metricsViewName: "ad_bids_metrics",
    ephemeralMeasures: [ephemeralMeasure],
    criteria: [
      {
        measure: ephemeralMeasure.name,
        type: MeasureFilterType.Value,
        operation: MeasureFilterOperation.LessThan,
        value1: "10",
        value2: "",
      },
    ],
  } as unknown as AlertFormValues;

  it("attaches the expression compute for an ephemeral measure", () => {
    const req = getAlertQueryArgsFromFormValues(
      baseFormValues,
      undefined,
      {} as TimeControlState,
      {},
    );

    expect(req.measures).toEqual([
      {
        name: ephemeralMeasure.name,
        expression: {
          expression: ephemeralMeasure.expression,
          displayName: ephemeralMeasure.displayName,
        },
      },
    ]);
  });

  it("builds percent-of-total on an ephemeral measure", () => {
    const req = getAlertQueryArgsFromFormValues(
      {
        ...baseFormValues,
        criteria: [
          {
            ...baseFormValues.criteria[0],
            type: MeasureFilterType.PercentOfTotal,
          },
        ],
      },
      undefined,
      {} as TimeControlState,
      {},
    );

    expect(req.measures).toEqual([
      {
        name: ephemeralMeasure.name,
        expression: {
          expression: ephemeralMeasure.expression,
          displayName: ephemeralMeasure.displayName,
        },
      },
      {
        name: ephemeralMeasure.name + ComparisonPercentOfTotal,
        percentOfTotal: { measure: ephemeralMeasure.name },
      },
    ]);
  });

  it("leaves a spec measure untouched", () => {
    const req = getAlertQueryArgsFromFormValues(
      { ...baseFormValues, measure: "total_records" },
      undefined,
      {} as TimeControlState,
      {},
    );

    expect(req.measures).toEqual([{ name: "total_records" }]);
  });
});
