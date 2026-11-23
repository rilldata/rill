import {
  Runtimev1Operation,
  type V1Expression,
} from "@rilldata/web-admin/client";
import { render } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import FiltersCell from "./FiltersCell.svelte";

function inExpression(
  dimension: string,
  values: string[],
  op: Runtimev1Operation = Runtimev1Operation.OPERATION_IN,
): V1Expression {
  return {
    cond: {
      op,
      exprs: [{ ident: dimension }, ...values.map((val) => ({ val }))],
    },
  };
}

describe("FiltersCell", () => {
  it("renders a chip per condition when a dimension is filtered more than once", () => {
    const { container } = render(FiltersCell, {
      props: {
        metricsViewFilters: {
          // An include and an exclude on the same dimension.
          sales: {
            cond: {
              op: Runtimev1Operation.OPERATION_AND,
              exprs: [
                inExpression("country", ["USA"]),
                inExpression(
                  "country",
                  ["CAN"],
                  Runtimev1Operation.OPERATION_NIN,
                ),
              ],
            },
          },
          // The same dimension filtered in a second metrics view.
          costs: inExpression("country", ["MEX"]),
        },
      },
    });

    const chips = container.querySelectorAll(".font-bold");
    expect(Array.from(chips).map((chip) => chip.textContent?.trim())).toEqual([
      "country",
      "Exclude country",
      "country",
    ]);
  });
});
