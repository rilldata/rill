import {
  type V1BillingIssue,
  V1BillingIssueType,
} from "@rilldata/web-admin/client";
import type { BillingIssueMessage } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { shiftToLargest } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import { DateTime, type Duration } from "luxon";

const WarningPeriodInDays = 7;

export function getTrialIssue(issues: V1BillingIssue[]) {
  return issues.find(
    (i) =>
      i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_ON_TRIAL ||
      i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_ENDED ||
      i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_ON_CREDIT_TRIAL ||
      i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_CREDITS_DEPLETED,
  );
}

/**
 * @param creditBalance Remaining credit balance for orgs on the credit trial;
 * undefined if the balance could not be fetched.
 */
export function getMessageForTrialPlan(
  trialIssue: V1BillingIssue,
  creditBalance?: number,
): BillingIssueMessage {
  if (trialIssue.type === V1BillingIssueType.BILLING_ISSUE_TYPE_ON_CREDIT_TRIAL)
    return getMessageForCreditsTrial(trialIssue, creditBalance);
  else if (
    trialIssue.type ===
    V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_CREDITS_DEPLETED
  )
    return getMessageForCreditsDepletedIssue();

  // Legacy time-based trial handling

  const endDateStr =
    trialIssue.metadata?.onTrial?.endDate ??
    trialIssue.metadata?.trialEnded?.gracePeriodEndDate ??
    "";

  const message: BillingIssueMessage = {
    type: "default",
    title: m.billing_trial_expired(),
    description: m.billing_choose_plan_to_maintain(),
    iconType: "alert",
    cta: {
      text: m.billing_choose_a_plan(),
      type: "show-upgrade",
      teamPlanDialogType: "base",
    },
  };

  const today = DateTime.now();
  const endDate = DateTime.fromJSDate(new Date(endDateStr));
  if (!endDate.isValid) {
    message.type = "warning";
    return message;
  }

  const diff = endDate.diff(today);
  if (
    diff.milliseconds > 0 &&
    trialIssue.type !== V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_ENDED
  ) {
    const daysDiff = diff.shiftTo("days");
    message.title = getTrialMessageForDays(diff);
    message.type = daysDiff.days < WarningPeriodInDays ? "warning" : "default";
  } else {
    const gracePeriodDate = DateTime.fromJSDate(
      new Date(trialIssue.metadata?.trialEnded?.gracePeriodEndDate ?? ""),
    );
    const gracePeriodDiff = gracePeriodDate.isValid
      ? gracePeriodDate.diff(today)
      : null;
    if (gracePeriodDiff && gracePeriodDiff.milliseconds > 0) {
      message.description = m.billing_upgrade_within_to_maintain({
        duration: humanizeDuration(gracePeriodDiff),
      });
      message.type = "warning";
    } else {
      message.title = m.billing_trial_expired_hibernating();
      message.description = m.billing_upgrade_to_wake();
      message.type = "error";
      if (message.cta) message.cta.teamPlanDialogType = "trial-expired";
    }
  }

  return message;
}

const CreditTrialBannerDismissTTL = 24 * 60 * 60 * 1000; // 24 hrs

function getMessageForCreditsTrial(
  trialIssue: V1BillingIssue,
  creditBalance: number | undefined,
) {
  const message: BillingIssueMessage = {
    type: "default",
    title: m.billing_trial_expired(),
    description: m.billing_choose_plan_to_maintain(),
    iconType: "alert",
    cta: {
      text: m.billing_choose_a_plan(),
      type: "show-upgrade",
    },
  };
  const onCreditTrial = trialIssue.metadata?.onCreditTrial;
  if (!onCreditTrial?.creditAllocation) return message;

  const org = trialIssue.org ?? "";
  const billingUrl = `/${encodeURIComponent(org)}/-/settings/billing`;
  const amount =
    creditBalance === undefined ? undefined : formatTrialCredit(creditBalance);
  message.includesHtml = true;
  message.description = "";

  if (onCreditTrial.lowCredit) {
    message.type = "warning";
    message.title = amount
      ? m.billing_trial_credit_running_low_remaining({ billingUrl, amount })
      : m.billing_trial_credit_running_low();
    message.dismissible = {
      key: org,
      id: `${trialIssue.type ?? ""}-low-credits`,
      ttl: CreditTrialBannerDismissTTL,
    };
  } else {
    message.type = "default";
    message.title = amount
      ? m.billing_free_trial_remaining_credits({ billingUrl, amount })
      : m.billing_free_trial_see_remaining_credits({ billingUrl });
    message.cta = {
      text: m.billing_upgrade_today(),
      type: "show-upgrade",
    };
    message.dismissible = {
      key: org,
      id: `${trialIssue.type ?? ""}-remaining-credits`,
      ttl: CreditTrialBannerDismissTTL,
    };
  }
  return message;
}

// Rounds down to whole dollars so the banner never overstates the remaining credit.
function formatTrialCredit(credit: number) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(Math.max(0, Math.floor(credit)));
}

function getMessageForCreditsDepletedIssue() {
  return {
    type: "error",
    title: m.billing_credits_depleted(),
    description: "",
    iconType: "alert",
    cta: {
      text: m.billing_choose_a_plan(),
      type: "show-upgrade",
    },
  } satisfies BillingIssueMessage;
}

export function getTrialMessageForDays(diff: Duration) {
  if (diff.milliseconds < 0) return m.billing_trial_expired();
  return m.billing_trial_expires_in({ duration: humanizeDuration(diff) });
}

export function trialHasPastGracePeriod(trialEndedIssue: V1BillingIssue) {
  const gracePeriodDate = new Date(
    trialEndedIssue.metadata?.trialEnded?.gracePeriodEndDate ?? "",
  );
  const gracePeriodTime = gracePeriodDate.getTime();
  return Number.isNaN(gracePeriodTime) || gracePeriodTime < Date.now();
}

function humanizeDuration(dur: Duration) {
  dur = shiftToLargest(dur, ["seconds", "minutes", "hours", "days"]);
  return dur.toHuman({ unitDisplay: "short" });
}
