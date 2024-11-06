import {
  type V1BillingIssue,
  V1BillingIssueType,
} from "@rilldata/web-admin/client";
import type { BillingIssueMessage } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";
import { shiftToLargest } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import { DateTime, type Duration } from "luxon";

const WarningPeriodInDays = 7;

export function getTrialIssue(issues: V1BillingIssue[]) {
  return issues.find(
    (i) =>
      i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_ON_TRIAL ||
      i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_ENDED,
  );
}

export function getMessageForTrialPlan(
  trialIssue: V1BillingIssue,
): BillingIssueMessage {
  const endDateStr =
    trialIssue.metadata?.onTrial?.endDate ??
    trialIssue.metadata?.trialEnded?.gracePeriodEndDate ??
    "";

  const message: BillingIssueMessage = {
    type: "default",
    title: "Your trial has expired.",
    description: "Upgrade to maintain access.",
    iconType: "alert",
    cta: {
      text: "Upgrade",
      type: "upgrade",
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
    message.title = getTrialMessageForDays(diff, endDate);
    message.type = daysDiff.days < WarningPeriodInDays ? "warning" : "default";
  } else {
    const gracePeriodDate = DateTime.fromJSDate(
      new Date(trialIssue.metadata?.trialEnded?.gracePeriodEndDate ?? ""),
    );
    const gracePeriodDiff = gracePeriodDate.isValid
      ? gracePeriodDate.diff(today)
      : null;
    if (gracePeriodDiff && gracePeriodDiff.milliseconds > 0) {
      message.description = `Upgrade within ${humanizeDuration(gracePeriodDiff)} to maintain access.`;
      message.type = "warning";
    } else {
      message.title =
        "Your trial has expired and this org’s projects are now hibernating.";
      message.description = "Upgrade to wake projects and regain full access.";
      message.type = "error";
      message.cta.teamPlanDialogType = "trial-expired";
    }
  }

  return message;
}

export function getTrialMessageForDays(diff: Duration, endDate: DateTime) {
  if (diff.milliseconds < 0) return "Your trial has expired.";
  const UTCFormattedDate = endDate
    .setLocale("UTC")
    .toLocaleString(DateTime.DATE_MED);
  return `Your trial expires in ${humanizeDuration(diff)} (${UTCFormattedDate} UTC).`;
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
