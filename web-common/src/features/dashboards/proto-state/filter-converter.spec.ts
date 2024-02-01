import { convertFilterToExpression } from "@rilldata/web-common/features/dashboards/proto-state/filter-converter";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { describe, it, expect } from "vitest";

describe("filter-converter", () => {
  it("sanity test", () => {
    const expr = convertFilterToExpression({
      include: [
        {
          name: "dim0",
          in: [{ kind: { value: "da0" } }, { kind: { value: "da1" } }],
        },
        {
          name: "dim1",
          in: [{ kind: { value: true } }],
        },
      ],
      exclude: [
        {
          name: "dim2",
          in: [{ kind: { value: 2 } }, { kind: { value: 4 } }],
        },
      ],
    });
    expect(expr).toEqual(
      createAndExpression([
        createInExpression("dim0", ["da0", "da1"]),
        createInExpression("dim1", [true]),
        createInExpression("dim2", [2, 4], true),
      ]),
    );
  });
});
