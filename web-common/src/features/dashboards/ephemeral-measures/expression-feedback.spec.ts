import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import { describeExpressionError } from "./expression-feedback";
import { EPHEMERAL_MEASURE_FUNCTIONS } from "./expression-parser";

const MEASURES: MetricsViewSpecMeasure[] = [
  { name: "revenue", displayName: "Revenue" },
  { name: "orders", displayName: "# of Orders" },
  { name: "select", displayName: "Selected" },
];

describe("describeExpressionError", () => {
  it("returns nothing for an empty or valid expression", () => {
    expect(describeExpressionError("", MEASURES)).toBeUndefined();
    expect(describeExpressionError("   ", MEASURES)).toBeUndefined();
    expect(
      describeExpressionError("round(revenue / orders, 2)", MEASURES),
    ).toBeUndefined();
    expect(
      describeExpressionError('revenue + "select"', MEASURES),
    ).toBeUndefined();
  });

  it("asks for an operator before an operand that follows another one", () => {
    // Two chips with nothing in between, named by the display name.
    expect(describeExpressionError("revenue orders ", MEASURES)).toBe(
      m.dashboard_pivot_ephemeral_error_missing_operator({
        operand: "# of Orders",
      }),
    );
    // Chips serialize reserved names quoted.
    expect(describeExpressionError('revenue "select"', MEASURES)).toBe(
      m.dashboard_pivot_ephemeral_error_missing_operator({
        operand: "Selected",
      }),
    );
    expect(describeExpressionError("revenue 2", MEASURES)).toBe(
      m.dashboard_pivot_ephemeral_error_missing_operator({ operand: "2" }),
    );
    expect(describeExpressionError("revenue (orders)", MEASURES)).toBe(
      m.dashboard_pivot_ephemeral_error_missing_operator({ operand: "(" }),
    );
  });

  it("describes the other common mistakes", () => {
    expect(describeExpressionError("revenue +", MEASURES)).toBe(
      m.dashboard_pivot_ephemeral_error_incomplete(),
    );
    expect(describeExpressionError("(revenue + orders", MEASURES)).toBe(
      m.dashboard_pivot_ephemeral_error_unclosed_parenthesis(),
    );
    expect(describeExpressionError("2 * 3", MEASURES)).toBe(
      m.dashboard_pivot_ephemeral_error_no_measure(),
    );
    expect(describeExpressionError("revenue * / orders", MEASURES)).toBe(
      m.dashboard_pivot_ephemeral_error_invalid(),
    );
    expect(describeExpressionError("revenue + profit", MEASURES)).toBe(
      m.dashboard_pivot_ephemeral_error_unknown_measure({ name: "profit" }),
    );
    expect(describeExpressionError("avg(revenue)", MEASURES)).toBe(
      m.dashboard_pivot_ephemeral_error_unsupported_function({
        name: "avg",
        functions: Object.keys(EPHEMERAL_MEASURE_FUNCTIONS).join(", "),
      }),
    );
  });

  it("falls back to the parser's message for rare errors", () => {
    const error = describeExpressionError("revenue + 'x'", MEASURES);
    expect(error?.startsWith("String literals are not allowed")).toBe(true);
  });
});
