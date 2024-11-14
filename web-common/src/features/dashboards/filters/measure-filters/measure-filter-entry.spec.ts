import {
  mapExprToMeasureFilter,
  mapMeasureFilterToExpr,
  type MeasureFilterEntry,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  MeasureFilterOperation,
  MeasureFilterType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import {
  createAndExpression,
  createBinaryExpression,
  createOrExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

const TestCases: [
  title: string,
  measureFilter: MeasureFilterEntry,
  expr: V1Expression | undefined,
][] = [
  [
    "greater than",
    {
      measure: "imp",
      operation: MeasureFilterOperation.GreaterThan,
      type: MeasureFilterType.Value,
      value1: "10",
      value2: "",
    },
    createBinaryExpression("imp", V1Operation.OPERATION_GT, 10),
  ],
  [
    "invalid value",
    {
      measure: "imp",
      operation: MeasureFilterOperation.GreaterThan,
      type: MeasureFilterType.Value,
      value1: "abc",
      value2: "",
    },
    undefined,
  ],
  [
    "greater than or equals",
    {
      measure: "imp",
      operation: MeasureFilterOperation.GreaterThanOrEquals,
      type: MeasureFilterType.Value,
      value1: "10",
      value2: "",
    },
    createBinaryExpression("imp", V1Operation.OPERATION_GTE, 10),
  ],
  [
    "less than",
    {
      measure: "imp",
      operation: MeasureFilterOperation.LessThan,
      type: MeasureFilterType.Value,
      value1: "10",
      value2: "",
    },
    createBinaryExpression("imp", V1Operation.OPERATION_LT, 10),
  ],
  [
    "less than or equals",
    {
      measure: "imp",
      operation: MeasureFilterOperation.LessThanOrEquals,
      type: MeasureFilterType.Value,
      value1: "10",
      value2: "",
    },
    createBinaryExpression("imp", V1Operation.OPERATION_LTE, 10),
  ],
  [
    "equals",
    {
      measure: "imp",
      operation: MeasureFilterOperation.Equals,
      type: MeasureFilterType.Value,
      value1: "10",
      value2: "",
    },
    createBinaryExpression("imp", V1Operation.OPERATION_EQ, 10),
  ],
  [
    "not equals",
    {
      measure: "imp",
      operation: MeasureFilterOperation.NotEquals,
      type: MeasureFilterType.Value,
      value1: "10",
      value2: "",
    },
    createBinaryExpression("imp", V1Operation.OPERATION_NEQ, 10),
  ],
  [
    "between",
    {
      measure: "imp",
      operation: MeasureFilterOperation.Between,
      type: MeasureFilterType.Value,
      value1: "10",
      value2: "20",
    },
    createAndExpression([
      createBinaryExpression("imp", V1Operation.OPERATION_GT, 10),
      createBinaryExpression("imp", V1Operation.OPERATION_LT, 20),
    ]),
  ],
  [
    "not between",
    {
      measure: "imp",
      operation: MeasureFilterOperation.NotBetween,
      type: MeasureFilterType.Value,
      value1: "10",
      value2: "20",
    },
    createOrExpression([
      createBinaryExpression("imp", V1Operation.OPERATION_LTE, 10),
      createBinaryExpression("imp", V1Operation.OPERATION_GTE, 20),
    ]),
  ],
  [
    "increases by value",
    {
      measure: "imp",
      operation: MeasureFilterOperation.GreaterThan,
      type: MeasureFilterType.AbsoluteChange,
      value1: "10",
      value2: "",
    },
    createBinaryExpression("imp_delta", V1Operation.OPERATION_GT, 10),
  ],
  [
    "decreases by percent",
    {
      measure: "imp",
      operation: MeasureFilterOperation.LessThan,
      type: MeasureFilterType.PercentChange,
      value1: "-10",
      value2: "",
    },
    createBinaryExpression("imp_delta_perc", V1Operation.OPERATION_LT, -0.1),
  ],
  [
    "changes by absolute",
    {
      measure: "imp",
      operation: MeasureFilterOperation.Between,
      type: MeasureFilterType.AbsoluteChange,
      value1: "-10",
      value2: "10",
    },
    undefined,
  ],
  [
    "changes by percent",
    {
      measure: "imp",
      operation: MeasureFilterOperation.Between,
      type: MeasureFilterType.PercentChange,
      value1: "-10",
      value2: "10",
    },
    undefined,
  ],
  [
    "percent of totals",
    {
      measure: "imp",
      operation: MeasureFilterOperation.LessThan,
      type: MeasureFilterType.PercentOfTotal,
      value1: "10",
      value2: "",
    },
    createBinaryExpression(
      "imp_percent_of_total",
      V1Operation.OPERATION_LT,
      0.1,
    ),
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
