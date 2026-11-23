import {
  type V1BillingIssue,
  V1BillingIssueType,
} from "@rilldata/web-admin/client";
import { getMessageForTrialPlan } from "@rilldata/web-admin/features/billing/issues/getMessageForTrialPlan";
import { describe, expect, it } from "vitest";

function creditTrialIssue(org: string, lowCredit = false): V1BillingIssue {
  return {
    org,
    type: V1BillingIssueType.BILLING_ISSUE_TYPE_ON_CREDIT_TRIAL,
    metadata: { onCreditTrial: { creditAllocation: 250, lowCredit } },
  };
}

describe("getMessageForTrialPlan for the credit trial", () => {
  it("shows the remaining credits rounded down, linked to the billing page", () => {
    const message = getMessageForTrialPlan(creditTrialIssue("acme"), 221.87);
    expect(message.title).toBe(
      'You are on a free trial with <a href="/acme/-/settings/billing">$221 of remaining credits</a>.',
    );
    expect(message.includesHtml).toBe(true);
    expect(message.cta?.text).toBe("Upgrade today");
    expect(message.type).toBe("default");
  });

  it("falls back to a link to the billing page when the balance is unavailable", () => {
    const message = getMessageForTrialPlan(creditTrialIssue("acme"), undefined);
    expect(message.title).toBe(
      'You are on a free trial, <a href="/acme/-/settings/billing">see remaining credits</a>.',
    );
  });

  it("shows a zero balance instead of falling back", () => {
    const message = getMessageForTrialPlan(creditTrialIssue("acme"), 0);
    expect(message.title).toContain("$0 of remaining credits");
  });

  it("shows the remaining credits in the low-credit warning", () => {
    const message = getMessageForTrialPlan(
      creditTrialIssue("acme", true),
      12.5,
    );
    expect(message.title).toBe(
      'Your trial credit is running low: <a href="/acme/-/settings/billing">$12 remaining</a>.',
    );
    expect(message.type).toBe("warning");
  });

  it("keeps the low-credit warning when the balance is unavailable", () => {
    const message = getMessageForTrialPlan(
      creditTrialIssue("acme", true),
      undefined,
    );
    expect(message.title).toBe("Your trial credit is running low.");
  });

  it("encodes the org name in the billing link", () => {
    const message = getMessageForTrialPlan(creditTrialIssue("a&b"), 100);
    expect(message.title).toContain('href="/a%26b/-/settings/billing"');
  });

  it("dismisses for 24 hours under a new ID so older dismissals don't hide it", () => {
    const message = getMessageForTrialPlan(creditTrialIssue("acme"), 100);
    expect(message.dismissible).toEqual({
      key: "acme",
      id: "BILLING_ISSUE_TYPE_ON_CREDIT_TRIAL-remaining-credits",
      ttl: 24 * 60 * 60 * 1000,
    });
  });
});
