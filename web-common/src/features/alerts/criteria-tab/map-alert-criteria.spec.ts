import {
  mapAlertCriteriaToExpression,
  mapExpressionToAlertCriteria,
} from "@rilldata/web-common/features/alerts/criteria-tab/map-alert-criteria";
import {
  CompareWith,
  CriteriaOperations,
} from "@rilldata/web-common/features/alerts/criteria-tab/operations";
import type { AlertCriteria } from "@rilldata/web-common/features/alerts/form-utils";
import {
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
  criteria: AlertCriteria,
  expr: V1Expression | undefined,
][] = [
  [
    "simple greater than",
    {
      value: "10",
      field: "imp",
      operation: CriteriaOperations.GreaterThan,
      compareWith: CompareWith.Value,
    },
    createBinaryExpression("imp", V1Operation.OPERATION_GT, 10),
  ],
  [
    "backwards compatibility for greater than or equals",
    {
      value: "10",
      field: "imp",
      operation: CriteriaOperations.GreaterThanOrEquals,
      compareWith: CompareWith.Value,
    },
    createBinaryExpression("imp", V1Operation.OPERATION_GTE, 10),
  ],
  [
    "simple less than",
    {
      value: "10",
      field: "imp",
      operation: CriteriaOperations.LessThan,
      compareWith: CompareWith.Value,
    },
    createBinaryExpression("imp", V1Operation.OPERATION_LT, 10),
  ],
  [
    "backwards compatibility for less than or equals",
    {
      value: "10",
      field: "imp",
      operation: CriteriaOperations.LessThanOrEquals,
      compareWith: CompareWith.Value,
    },
    createBinaryExpression("imp", V1Operation.OPERATION_LTE, 10),
  ],
  [
    "invalid greater than",
    {
      value: "10",
      field: "imp",
      operation: CriteriaOperations.GreaterThan,
      compareWith: CompareWith.Percent,
    },
    undefined,
  ],
  [
    "increases by value",
    {
      value: "10",
      field: "imp",
      operation: CriteriaOperations.IncreasesBy,
      compareWith: CompareWith.Value,
    },
    createBinaryExpression("imp__delta_abs", V1Operation.OPERATION_GT, 10),
  ],
  [
    "decreases by percent",
    {
      value: "10",
      field: "imp",
      operation: CriteriaOperations.DecreasesBy,
      compareWith: CompareWith.Percent,
    },
    createBinaryExpression("imp__delta_rel", V1Operation.OPERATION_LT, -0.1),
  ],
  [
    "changes by percent",
    {
      value: "10",
      field: "imp",
      operation: CriteriaOperations.ChangesBy,
      compareWith: CompareWith.Percent,
    },
    createOrExpression([
      createBinaryExpression("imp__delta_rel", V1Operation.OPERATION_LT, -0.1),
      createBinaryExpression("imp__delta_rel", V1Operation.OPERATION_GT, 0.1),
    ]),
  ],
  [
    "share of totals",
    {
      value: "10",
      field: "imp",
      operation: CriteriaOperations.ShareOfTotalsGreaterThan,
      compareWith: CompareWith.Percent,
    },
    undefined,
  ],
];

describe("mapAlertCriteriaToExpression", () => {
  TestCases.forEach(([title, criteria, expr]) => {
    it(title, () => {
      expect(mapAlertCriteriaToExpression(criteria)).toEqual(expr);
    });
  });
});

describe("mapExpressionToAlertCriteria", () => {
  TestCases.forEach(([title, criteria, expr]) => {
    if (!expr) return;
    it(title, () => {
      expect(mapExpressionToAlertCriteria(expr)).toEqual(criteria);
    });
  });
});
