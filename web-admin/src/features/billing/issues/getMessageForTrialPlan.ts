import {
  type V1BillingIssue,
  V1BillingIssueType,
} from "@rilldata/web-admin/client";
import type { BillingIssueMessage } from "@rilldata/web-admin/features/billing/issues/useBillingIssueMessage";
import { shiftToLargest } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import { DateTime, type Duration } from "luxon";
import type { BannerMessage } from "@rilldata/web-common/lib/event-bus/events.ts";

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
      message.description = `Upgrade within ${humanizeDuration(gracePeriodDiff)} to maintain access.`;
      message.type = "warning";
    } else {
      message.title =
        "Your trial has expired and this org’s projects are now hibernating.";
      message.description = "Upgrade to wake projects and regain full access.";
      message.type = "error";
      if (message.cta) message.cta.teamPlanDialogType = "trial-expired";
    }
  }

  return message;
}

function getMessageForCreditsTrial(trialIssue: V1BillingIssue) {
  const message: BillingIssueMessage = {
    type: "default",
    title: `Your trial has expired.`,
    description: "Subscribe to Pro to maintain access.",
    iconType: "alert",
    cta: {
      text: "Subscribe to Pro",
      type: "upgrade",
    },
  };
  const onCreditTrial = trialIssue.metadata?.onCreditTrial;
  if (!onCreditTrial?.creditAllocation) return message;

  if (onCreditTrial.lowCredit) {
    message.type = "warning";
    message.title = "Your trial credit is running low.";
    message.description = "";
    message.dismissible = {
      key: trialIssue.org ?? "",
      id: `${trialIssue.type ?? ""}-low-credits`,
      ttl: 24 * 60 * 60, // 24 hrs
    };
  } else {
    message.type = "default";
    message.title = `Welcome to rill.`;
    message.description = `You've on a free trial with ${onCreditTrial.creditAllocation ?? 0}$ in credits.`;
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
    title:
      "Trial credit is used up. Projects are hibernated and dashboards are offline.",
    description: "",
    iconType: "alert",
    cta: {
      text: "Subscribe to Pro",
      type: "upgrade",
    },
  } satisfies BillingIssueMessage;
}

export function getTrialMessageForDays(diff: Duration) {
  if (diff.milliseconds < 0) return "Your trial has expired.";
  return `Your trial expires in ${humanizeDuration(diff)}.`;
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

function buildDismissableForIssue(issue: V1BillingIssue, ttl: number) {
  return {
    key: issue.org ?? "",
    id: issue.type ?? "",
    ttl,
  } satisfies BannerMessage["dismissible"];
}
