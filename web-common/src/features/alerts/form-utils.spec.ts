import {
  alertFormValidationSchema,
  AlertNotifyRowsLimitDefault,
  checkIsTabValid,
  getNotifyRowsLimitFromFormValues,
  type AlertFormValues,
} from "@rilldata/web-common/features/alerts/form-utils";
import {
  MeasureFilterOperation,
  MeasureFilterType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import type { ValidationErrors } from "sveltekit-superforms";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

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

const percentOfTotalCriteria = {
  measure: "revenue",
  operation: MeasureFilterOperation.GreaterThan,
  type: MeasureFilterType.PercentOfTotal,
  value1: "10",
  value2: "",
};
const valueCriteria = {
  ...percentOfTotalCriteria,
  type: MeasureFilterType.Value,
  value1: "100",
};
const needsSplit = m.alert_form_criteria_percent_of_total_needs_split({
  field: "criteria[0].value1",
});

function getValidFormValues(
  overrides: Partial<AlertFormValues>,
): AlertFormValues {
  return {
    name: "Revenue share",
    measure: "revenue",
    splitByDimension: "country",
    notifyRowsLimit: "5",
    criteria: [percentOfTotalCriteria],
    criteriaOperation: V1Operation.OPERATION_AND,
    snooze: "0",
    slackUsers: [],
    emailRecipients: [],
    enableEmailNotification: false,
    exportAttach: false,
    ...overrides,
  } as unknown as AlertFormValues;
}

describe("alertFormValidationSchema", () => {
  it("rejects percent-of-total criteria when the alert isn't split by a dimension", async () => {
    const result = await alertFormValidationSchema.validate(
      getValidFormValues({ splitByDimension: "" }),
    );
    expect(result.success).toBe(false);
    expect(
      result.success ? [] : result.issues.map((issue) => issue.message),
    ).toEqual([needsSplit]);
  });

  it("accepts percent-of-total criteria when the alert is split by a dimension", async () => {
    const result = await alertFormValidationSchema.validate(
      getValidFormValues({}),
    );
    expect(result.success).toBe(true);
  });

  it("accepts alerts without a split or criteria", async () => {
    const result = await alertFormValidationSchema.validate(
      getValidFormValues({ splitByDimension: "", criteria: [] }),
    );
    expect(result.success).toBe(true);
  });
});

describe("alertFormValidationSchema notifications", () => {
  async function getMessages(overrides: Partial<AlertFormValues>) {
    const result = await alertFormValidationSchema.validate(
      getValidFormValues({
        emailRecipients: ["other@example.com"],
        enableEmailNotification: true,
        embedOwnerEmail: "",
        ...overrides,
      }),
    );
    return result.success ? [] : result.issues.map((issue) => issue.message);
  }
  const noNotification = m.alert_form_no_notification();

  describe("in an embed", () => {
    beforeEach(() => {
      vi.spyOn(EmbedStore, "isEmbedded").mockReturnValue(true);
    });
    afterEach(() => {
      vi.restoreAllMocks();
    });

    it.each<[string, Partial<AlertFormValues>, string[]]>([
      [
        "rejects alerts without any recipient",
        { emailRecipients: [""] },
        [noNotification],
      ],
      [
        "rejects email recipients when email notifications are turned off",
        { enableEmailNotification: false },
        [noNotification],
      ],
      ["accepts an email recipient", {}, []],
      [
        "accepts only the creator's email",
        { enableEmailNotification: false, embedOwnerEmail: "me@example.com" },
        [],
      ],
    ])("%s", async (_, overrides, expected) => {
      expect(await getMessages(overrides)).toEqual(expected);
    });
  });

  it("accepts alerts without any recipient outside of an embed", async () => {
    expect(await getMessages({ emailRecipients: [""] })).toEqual([]);
  });
});

describe("checkIsTabValid", () => {
  it("rejects the criteria tab when one criterion has an error and another was cleared", () => {
    expect(
      checkIsTabValid(
        1,
        getValidFormValues({
          criteria: [percentOfTotalCriteria, valueCriteria],
        }),
        {
          criteria: { 0: { value1: [needsSplit] }, 1: { value1: undefined } },
        } as unknown as ValidationErrors<AlertFormValues>,
      ),
    ).toBe(false);
  });
});
