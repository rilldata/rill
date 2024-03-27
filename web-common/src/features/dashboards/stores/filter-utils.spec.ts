import { describe, expect, it } from "vitest";
import { V1Operation } from "../../../runtime-client";
import { isExpressionIncomplete } from "./filter-utils";

// Mock data for test cases
const testCases = [
  {
    description: "Simple complete expression with a single condition",
    criteria: {
      cond: {
        op: V1Operation.OPERATION_AND,
        exprs: [{ val: "1", ident: "measure_0" }],
      },
    },
    incomplete: false,
  },
  {
    description: "Simple incomplete expression with a missing val",
    criteria: {
      cond: { op: V1Operation.OPERATION_AND, exprs: [{ ident: "measure_0" }] },
    },
    incomplete: true,
  },
  {
    description: "Incomplete expression with an unspecified operation",
    criteria: { cond: { exprs: [{ val: "1", ident: "measure_0" }] } },
    incomplete: true,
  },
  {
    description: "Incomplete expression with an empty string val",
    criteria: {
      cond: {
        op: V1Operation.OPERATION_AND,
        exprs: [{ val: "", ident: "measure_0" }],
      },
    },
    incomplete: true,
  },
  {
    description:
      "Nested complete expression with one incomplete nested condition",
    criteria: {
      cond: {
        op: V1Operation.OPERATION_AND,
        exprs: [
          {
            cond: {
              op: V1Operation.OPERATION_OR,
              exprs: [{ val: "1", ident: "measure_0" }, { ident: "measure_1" }],
            },
          },
        ],
      },
    },
    incomplete: false,
  },
  {
    description:
      "Nested incomplete expression with unspecified operation in nested condition",
    criteria: {
      cond: {
        op: V1Operation.OPERATION_AND,
        exprs: [
          {
            cond: {
              exprs: [{ val: "1", ident: "measure_0" }, { ident: "measure_1" }],
            },
          },
        ],
      },
    },
    incomplete: true,
  },
  // Add more test cases as needed
];

// Test suite
describe("isExpressionIncomplete", () => {
  testCases.forEach((testCase, index) => {
    it(`Test case ${index + 1}: ${testCase.description}`, () => {
      const result = isExpressionIncomplete(testCase.criteria);
      expect(result).toBe(testCase.incomplete);
    });
  });
});
