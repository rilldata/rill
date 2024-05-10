import {
  mapExprToMeasureFilter,
  mapMeasureFilterToExpr,
  MeasureFilterComparisonType,
  MeasureFilterEntry,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { MeasureFilterOperation } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import {
  createAndExpression,
  createBinaryExpression,
  createOrExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { V1Expression, V1Operation } from "@rilldata/web-common/runtime-client";
import { describe, it, expect } from "vitest";

const TestCases: [
  title: string,
  measureFilter: MeasureFilterEntry,
  expr: V1Expression | undefined,
][] = [
  [
    "greater than",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.GreaterThan,
      comparison: MeasureFilterComparisonType.None,
    },
    createBinaryExpression("imp", V1Operation.OPERATION_GT, 10),
  ],
  [
    "invalid value",
    {
      value1: "abc",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.GreaterThan,
      comparison: MeasureFilterComparisonType.None,
    },
    undefined,
  ],
  [
    "greater than or equals",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.GreaterThanOrEquals,
      comparison: MeasureFilterComparisonType.None,
    },
    createBinaryExpression("imp", V1Operation.OPERATION_GTE, 10),
  ],
  [
    "less than",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.LessThan,
      comparison: MeasureFilterComparisonType.None,
    },
    createBinaryExpression("imp", V1Operation.OPERATION_LT, 10),
  ],
  [
    "less than or equals",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.LessThanOrEquals,
      comparison: MeasureFilterComparisonType.None,
    },
    createBinaryExpression("imp", V1Operation.OPERATION_LTE, 10),
  ],
  [
    "between",
    {
      value1: "10",
      value2: "20",
      measure: "imp",
      operation: MeasureFilterOperation.Between,
      comparison: MeasureFilterComparisonType.None,
    },
    createAndExpression([
      createBinaryExpression("imp", V1Operation.OPERATION_GT, 10),
      createBinaryExpression("imp", V1Operation.OPERATION_LT, 20),
    ]),
  ],
  [
    "not between",
    {
      value1: "10",
      value2: "20",
      measure: "imp",
      operation: MeasureFilterOperation.NotBetween,
      comparison: MeasureFilterComparisonType.None,
    },
    createOrExpression([
      createBinaryExpression("imp", V1Operation.OPERATION_LTE, 10),
      createBinaryExpression("imp", V1Operation.OPERATION_GTE, 20),
    ]),
  ],
  [
    "invalid greater than",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.GreaterThan,
      comparison: MeasureFilterComparisonType.PercentageComparison,
    },
    undefined,
  ],
  [
    "increases by value",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.IncreasesBy,
      comparison: MeasureFilterComparisonType.AbsoluteComparison,
    },
    createBinaryExpression("imp__delta_abs", V1Operation.OPERATION_GT, 10),
  ],
  [
    "decreases by percent",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.DecreasesBy,
      comparison: MeasureFilterComparisonType.PercentageComparison,
    },
    createBinaryExpression("imp__delta_rel", V1Operation.OPERATION_LT, -0.1),
  ],
  [
    "changes by percent",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.ChangesBy,
      comparison: MeasureFilterComparisonType.PercentageComparison,
    },
    createOrExpression([
      createBinaryExpression("imp__delta_rel", V1Operation.OPERATION_LT, -0.1),
      createBinaryExpression("imp__delta_rel", V1Operation.OPERATION_GT, 0.1),
    ]),
  ],
  [
    // TODO
    "share of totals",
    {
      value1: "10",
      value2: "",
      measure: "imp",
      operation: MeasureFilterOperation.ShareOfTotalsGreaterThan,
      comparison: MeasureFilterComparisonType.PercentageComparison,
    },
    undefined,
  ],
];

describe("mapMeasureFilterToExpr", () => {
  TestCases.forEach(([title, criteria, expr]) => {
    it(title, () => {
      expect(mapMeasureFilterToExpr(criteria)).toEqual(expr);
    });
  });
});

describe("mapExprToMeasureFilter", () => {
  TestCases.forEach(([title, criteria, expr]) => {
    if (!expr) return;
    it(title, () => {
      expect(mapExprToMeasureFilter(expr)).toEqual(criteria);
    });
  });
});
