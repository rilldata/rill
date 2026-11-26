import { describe, expect, it } from "vitest";
import {
  controlLabel,
  controlFieldLabels,
  controlOptionLabel,
  controlSummary,
  timeRangeLabel,
} from "./control-labels";

const fieldLabel = (name: string) =>
  ({ total_bids: "Total bids", publisher: "Publisher" })[name] ?? name;

describe("control labels", () => {
  it("names time ranges and grains the way the time picker does", () => {
    expect(timeRangeLabel("P7D")).toBe("Last 7 Days");
    expect(timeRangeLabel("rill-MTD")).toBe("Month to date");
    expect(controlOptionLabel({ kind: "time_grain" }, "week", fieldLabel)).toBe(
      "Week",
    );
  });

  it("uses display names for fields and raw values for filter values", () => {
    expect(
      controlOptionLabel({ kind: "measure" }, "total_bids", fieldLabel),
    ).toBe("Total bids");
    expect(
      controlOptionLabel({ kind: "filter" }, "total_bids", fieldLabel),
    ).toBe("total_bids");
  });

  it("labels a filter by its dimension and leaves an unlabeled swap bare", () => {
    expect(
      controlLabel({ kind: "filter", target: "publisher" }, fieldLabel),
    ).toBe("Publisher");
    expect(
      controlLabel({ kind: "measure", target: "total_bids" }, fieldLabel),
    ).toBe("");
    expect(
      controlLabel(
        { kind: "measure", target: "total_bids", label: "Metric" },
        fieldLabel,
      ),
    ).toBe("Metric");
  });

  it("summarizes values", () => {
    const filter = { kind: "filter", target: "publisher" };
    expect(controlSummary(filter, [], fieldLabel)).toBe("All");
    expect(controlSummary(filter, ["Acme"], fieldLabel)).toBe("Acme");
    expect(controlSummary(filter, ["Acme", "Globex"], fieldLabel)).toBe(
      "Acme +1 other",
    );
    expect(controlSummary(filter, ["a", "b", "c"], fieldLabel)).toBe(
      "a +2 others",
    );
    expect(controlSummary({ kind: "time_range" }, ["P7D"], fieldLabel)).toBe(
      "Last 7 Days",
    );
    expect(controlSummary({ kind: "time_range" }, [], fieldLabel)).toBe(
      "Inherit from canvas",
    );
  });
});

it("uses the same first field label for the viewer and inspector", () => {
  const labels = controlFieldLabels({
    a: {
      state: {
        validSpec: {
          measures: [{ name: "spend", displayName: "Spend" }],
          dimensions: [{ name: "country" }],
        },
      },
    },
    b: {
      state: {
        validSpec: {
          measures: [{ name: "spend", displayName: "Other spend" }],
        },
      },
    },
  });
  expect([...labels]).toEqual([
    ["spend", "Spend"],
    ["country", "country"],
  ]);
});
