import { describe, expect, it } from "vitest";
import { compileNotices, NO_NOTICES, truncationBanner } from "./notices";

const chart = { truncationNote: undefined, truncationWarning: undefined };

describe("compileNotices", () => {
  it("keeps only the warnings that drop values", () => {
    const notices = compileNotices({
      error: null,
      warnings: [
        {
          code: "invalid-option-value",
          message:
            "chartProperties.interpolate: 'wavy' is not a valid option. Falling back to the default.",
        },
        { code: "negative-values-pyramid", message: "Negative values." },
        {
          code: "overflow",
          message: "12 of 40 values in 'device' were omitted.",
        },
        {
          code: "too-many-groups-pyramid",
          message: "Only the first two are shown.",
        },
      ],
    });
    expect(notices).toEqual({
      error: undefined,
      dropped: [
        "12 of 40 values in 'device' were omitted.",
        "Only the first two are shown.",
      ],
    });
  });
});

describe("truncationBanner", () => {
  it("says nothing was left out for a warning that drops no data", () => {
    const notices = compileNotices({
      error: null,
      warnings: [{ code: "invalid-option-value", message: "Fell back." }],
    });
    expect(truncationBanner(chart, notices)).toBeUndefined();
  });

  it("explains Flint's cut, and prefers the server's note", () => {
    const notices = compileNotices({
      error: null,
      warnings: [{ code: "overflow", message: "3 values were omitted." }],
    });
    expect(truncationBanner(chart, notices)).toEqual({
      note: "Some values were left out to fit the chart.",
      detail: "3 values were omitted.",
    });
    expect(
      truncationBanner(
        {
          truncationNote: "Showing the top 10 values.",
          truncationWarning: "Capped at 10.",
        },
        NO_NOTICES,
      ),
    ).toEqual({ note: "Showing the top 10 values.", detail: "Capped at 10." });
  });
});
