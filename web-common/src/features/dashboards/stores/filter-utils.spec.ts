import { describe, expect, it } from "vitest";
import { V1Operation, type V1Expression } from "../../../runtime-client";
import {
  createAndExpression,
  createContainsAllExpression,
  createInExpression,
  getValuesInExpression,
  isContainsAllExpression,
  isExpressionIncomplete,
  isExpressionUnsupported,
  matchExpressionByName,
  filterIdentifiers,
  forEachIdentifier,
  getIdentFromContainsAllExpression,
} from "./filter-utils";

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

describe("contains-all expressions", () => {
  describe("createContainsAllExpression", () => {
    it("creates AND of single-value INs for include mode", () => {
      const expr = createContainsAllExpression("dim", ["a", "b", "c"]);
      expect(expr.cond?.op).toBe(V1Operation.OPERATION_AND);
      expect(expr.cond?.exprs).toHaveLength(3);
      expr.cond?.exprs?.forEach((child, i) => {
        expect(child.cond?.op).toBe(V1Operation.OPERATION_IN);
        expect(child.cond?.exprs?.[0]?.ident).toBe("dim");
        expect(child.cond?.exprs?.[1]?.val).toBe(["a", "b", "c"][i]);
      });
    });

    it("creates OR of single-value NINs for exclude mode", () => {
      const expr = createContainsAllExpression("dim", ["a", "b"], true);
      expect(expr.cond?.op).toBe(V1Operation.OPERATION_OR);
      expect(expr.cond?.exprs).toHaveLength(2);
      expr.cond?.exprs?.forEach((child) => {
        expect(child.cond?.op).toBe(V1Operation.OPERATION_NIN);
        expect(child.cond?.exprs?.[0]?.ident).toBe("dim");
      });
    });
  });

  describe("isContainsAllExpression", () => {
    it("detects AND of same-ident single-value INs", () => {
      const expr = createContainsAllExpression("dim", ["a", "b"]);
      expect(isContainsAllExpression(expr)).toBe(true);
    });

    it("detects OR of same-ident single-value NINs", () => {
      const expr = createContainsAllExpression("dim", ["a", "b"], true);
      expect(isContainsAllExpression(expr)).toBe(true);
    });

    it("returns false for regular IN expression", () => {
      const expr = createInExpression("dim", ["a", "b"]);
      expect(isContainsAllExpression(expr)).toBe(false);
    });

    it("returns false for AND with different idents", () => {
      const expr: V1Expression = {
        cond: {
          op: V1Operation.OPERATION_AND,
          exprs: [
            createInExpression("dim1", ["a"]),
            createInExpression("dim2", ["b"]),
          ],
        },
      };
      expect(isContainsAllExpression(expr)).toBe(false);
    });

    it("returns false for AND with multi-value INs", () => {
      const expr: V1Expression = {
        cond: {
          op: V1Operation.OPERATION_AND,
          exprs: [createInExpression("dim", ["a", "b"])],
        },
      };
      expect(isContainsAllExpression(expr)).toBe(false);
    });

    it("returns false for empty expression", () => {
      expect(isContainsAllExpression({})).toBe(false);
    });
  });

  describe("getValuesInExpression with contains-all", () => {
    it("extracts values from contains-all include expression", () => {
      const expr = createContainsAllExpression("dim", ["x", "y", "z"]);
      expect(getValuesInExpression(expr)).toEqual(["x", "y", "z"]);
    });

    it("extracts values from contains-all exclude expression", () => {
      const expr = createContainsAllExpression("dim", ["a", "b"], true);
      expect(getValuesInExpression(expr)).toEqual(["a", "b"]);
    });

    it("still extracts values from regular IN expression", () => {
      const expr = createInExpression("dim", ["a", "b", "c"]);
      expect(getValuesInExpression(expr)).toEqual(["a", "b", "c"]);
    });
  });

  describe("getIdentFromContainsAllExpression", () => {
    it("returns the ident from a contains-all expression", () => {
      const expr = createContainsAllExpression("myDim", ["a"]);
      expect(getIdentFromContainsAllExpression(expr)).toBe("myDim");
    });
  });

  describe("matchExpressionByName with contains-all", () => {
    it("matches contains-all expression by dimension name", () => {
      const expr = createContainsAllExpression("city", ["NYC", "LA"]);
      expect(matchExpressionByName(expr, "city")).toBe(true);
      expect(matchExpressionByName(expr, "country")).toBe(false);
    });

    it("still matches regular IN expression by name", () => {
      const expr = createInExpression("city", ["NYC"]);
      expect(matchExpressionByName(expr, "city")).toBe(true);
    });
  });

  describe("isExpressionUnsupported with contains-all", () => {
    it("accepts top-level filter with contains-all child", () => {
      const whereFilter = createAndExpression([
        createContainsAllExpression("tags", ["a", "b"]),
      ]);
      expect(isExpressionUnsupported(whereFilter)).toBe(false);
    });

    it("accepts mix of regular IN and contains-all", () => {
      const whereFilter = createAndExpression([
        createInExpression("country", ["US", "UK"]),
        createContainsAllExpression("tags", ["a", "b"]),
      ]);
      expect(isExpressionUnsupported(whereFilter)).toBe(false);
    });
  });

  describe("filterIdentifiers with contains-all", () => {
    it("finds and filters contains-all expressions by ident", () => {
      const whereFilter = createAndExpression([
        createInExpression("country", ["US"]),
        createContainsAllExpression("tags", ["a", "b"]),
      ]);

      const result = filterIdentifiers(whereFilter, (_e, ident) => {
        return ident !== "tags";
      });

      // Should have removed the tags filter
      expect(result?.cond?.exprs).toHaveLength(1);
      expect(result?.cond?.exprs?.[0]?.cond?.exprs?.[0]?.ident).toBe(
        "country",
      );
    });
  });

  describe("forEachIdentifier with contains-all", () => {
    it("visits contains-all as a single expression with its ident", () => {
      const whereFilter = createAndExpression([
        createInExpression("country", ["US"]),
        createContainsAllExpression("tags", ["a", "b"]),
      ]);

      const visited: string[] = [];
      forEachIdentifier(whereFilter, (_e, ident) => {
        visited.push(ident);
      });

      expect(visited).toContain("country");
      expect(visited).toContain("tags");
    });
  });
});
