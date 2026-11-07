import { describe, expect, it } from "vitest";
import { getDroppedEphemeralMeasures } from "./form-utils";

const allMeasures = [
  { name: "revenue", displayName: "Revenue" },
  { name: "cost", displayName: "Cost" },
  { name: "users", displayName: "Users" },
];

describe("getDroppedEphemeralMeasures", () => {
  it("reports measures that reference hidden measures and keeps the rest", () => {
    const dropped = getDroppedEphemeralMeasures(
      {
        ephemeralMeasures: [
          {
            name: "profit",
            displayName: "Profit",
            expression: "revenue - cost",
          },
          { name: "arpu", displayName: "ARPU", expression: "revenue / users" },
        ],
      },
      ["revenue", "users"],
      allMeasures,
    );

    expect(dropped).toEqual([
      { name: "profit", displayName: "Profit", hiddenMeasures: ["Cost"] },
    ]);
  });

  it("falls back to the measure name when the hidden measure has no display name", () => {
    const dropped = getDroppedEphemeralMeasures(
      {
        ephemeralMeasures: [
          {
            name: "profit",
            displayName: "Profit",
            expression: "revenue - cost",
          },
        ],
      },
      ["revenue"],
      [{ name: "revenue", displayName: "Revenue" }],
    );

    expect(dropped[0].hiddenMeasures).toEqual(["cost"]);
  });

  it("drops nothing when every field is visible", () => {
    const dropped = getDroppedEphemeralMeasures(
      {
        ephemeralMeasures: [
          {
            name: "profit",
            displayName: "Profit",
            expression: "revenue - cost",
          },
        ],
      },
      undefined,
      allMeasures,
    );

    expect(dropped).toEqual([]);
  });
});
