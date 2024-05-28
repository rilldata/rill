import { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import { getTitleAndDescription } from "@rilldata/web-common/features/alerts/utils";
import {
  MeasureFilterOperation,
  MeasureFilterType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

const MetricsView: V1MetricsViewSpec = {
  measures: [
    {
      name: "total_records",
      label: "Total records",
    },
  ],
};

describe("getTitleAndDescription", () => {
  const TestCases: {
    title: string;
    formValues: Partial<AlertFormValues>;
    expected: ReturnType<typeof getTitleAndDescription>;
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
      expected: {
        title: "Total records value alert",
        description: "Triggers when Total records value is less than 10",
      },
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
        comparisonTimeRange: {
          isoOffset: "rill-PW",
        },
      },
      expected: {
        title: "Total records change vs previous week alert",
        description:
          "Triggers when Total records change compared to the previous week is greater than or equal to 10",
      },
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
        comparisonTimeRange: {
          isoOffset: "rill-PM",
        },
      },
      expected: {
        title: "Total records % change vs previous month alert",
        description:
          "Triggers when Total records % change compared to the previous month is greater than -10%",
      },
    },
  ];

  for (const { title, formValues, expected } of TestCases) {
    it(title, () => {
      expect(
        getTitleAndDescription(formValues as AlertFormValues, MetricsView),
      ).toEqual(expected);
    });
  }
});
