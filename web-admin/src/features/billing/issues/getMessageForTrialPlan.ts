import {
  type V1BillingIssue,
  V1BillingIssueType,
} from "@rilldata/web-admin/client";
import type { BillingIssueMessage } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";
import * as m from "@rilldata/web-common/paraglide/messages.js";
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

export function getMessageForTrialPlan(
  trialIssue: V1BillingIssue,
): BillingIssueMessage {
  if (trialIssue.type === V1BillingIssueType.BILLING_ISSUE_TYPE_ON_CREDIT_TRIAL)
    return getMessageForCreditsTrial(trialIssue);
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
      message.description = m.billing_upgrade_within_to_maintain({ duration: humanizeDuration(gracePeriodDiff) });
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

function getMessageForCreditsTrial(trialIssue: V1BillingIssue) {
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

  if (onCreditTrial.lowCredit) {
    message.type = "warning";
    message.title = m.billing_trial_credit_running_low();
    message.description = "";
    message.dismissible = {
      key: trialIssue.org ?? "",
      id: `${trialIssue.type ?? ""}-low-credits`,
      ttl: 24 * 60 * 60, // 24 hrs
    };
  } else {
    message.type = "default";
    message.title = m.billing_welcome_to_rill();
    message.description = m.billing_free_trial_with_credits({ amount: String(onCreditTrial.creditAllocation ?? 0) });
    message.dismissible = {
      key: trialIssue.org ?? "",
      id: `${trialIssue.type ?? ""}`,
      ttl: 0, // Doesnt appear again once dismissed
    };
  }
  return message;
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
