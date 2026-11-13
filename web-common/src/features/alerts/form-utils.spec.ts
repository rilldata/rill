import {
  AlertNotifyRowsLimitDefault,
  getNotifyRowsLimitFromFormValues,
  type AlertFormValues,
} from "@rilldata/web-common/features/alerts/form-utils";
import { describe, expect, it } from "vitest";

function getFormValues(overrides: Partial<AlertFormValues>): AlertFormValues {
  return {
    splitByDimension: "country",
    notifyRowsLimit: "10",
    ...overrides,
  } as AlertFormValues;
}

describe("getNotifyRowsLimitFromFormValues", () => {
  it("returns the limit for alerts split by a dimension", () => {
    expect(getNotifyRowsLimitFromFormValues(getFormValues({}))).toBe(10);
  });

  it("leaves the default limit unset", () => {
    expect(
      getNotifyRowsLimitFromFormValues(
        getFormValues({ notifyRowsLimit: String(AlertNotifyRowsLimitDefault) }),
      ),
    ).toBe(0);
  });

  it("leaves the limit unset for alerts that aren't split by a dimension", () => {
    expect(
      getNotifyRowsLimitFromFormValues(getFormValues({ splitByDimension: "" })),
    ).toBe(0);
  });
});
