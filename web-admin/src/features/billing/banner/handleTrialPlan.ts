import {
  type V1BillingIssue,
  V1BillingIssueType,
  type V1Subscription,
} from "@rilldata/web-admin/client";
import { showUpgradeDialog } from "@rilldata/web-admin/features/billing/banner/bannerCTADialogs";
import type { BannerMessage } from "@rilldata/web-common/lib/event-bus/events";
import { shiftToLargest } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import { DateTime, type Duration } from "luxon";

const WarningPeriodInDays = 7;

const cta: BannerMessage["cta"] = {
  text: "Upgrade ->",
  type: "button",
  onClick: () => showUpgradeDialog.set(true),
};

export function handleTrialPlan(
  subscription: V1Subscription,
  issues: V1BillingIssue[],
): BannerMessage {
  const trialIssue = issues[issues.length - 1];

  const today = DateTime.now();
  const endDate = DateTime.fromJSDate(new Date(subscription.trialEndDate));
  if (!endDate.isValid || !trialIssue) {
    return {
      type: "warning",
      message: "Your trial has expired. Upgrade to maintain access.",
      iconType: "alert",
      cta,
    };
  }

  const bannerMessage: BannerMessage = {
    type: "info",
    message: "",
    iconType: "alert",
    cta,
  };
  const diff = endDate.diff(today);
  if (
    diff.milliseconds > 0 &&
    trialIssue.type !== V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_ENDED
  ) {
    const daysDiff = diff.shiftTo("days");
    bannerMessage.message += `${getTrialMessageForDays(diff)} Upgrade to maintain access.`;
    bannerMessage.type =
      daysDiff.days < WarningPeriodInDays ? "warning" : "info";
  } else {
    const gracePeriodDate = DateTime.fromJSDate(
      new Date(trialIssue.metadata?.trialEnded?.gracePeriodEndDate),
    );
    const gracePeriodDiff = gracePeriodDate.isValid
      ? gracePeriodDate.diff(today)
      : null;
    if (gracePeriodDiff && gracePeriodDiff.milliseconds > 0) {
      bannerMessage.message = `Your trial has expired. Upgrade within ${getTrialMessageForDays(gracePeriodDiff)} to maintain access.`;
      bannerMessage.type = "warning";
    } else {
      bannerMessage.message = `Your trial has expired and this org’s projects are now hibernating. Upgrade to wake projects and regain full access.`;
      bannerMessage.type = "error";
    }
  }

  return bannerMessage;
}

export function getTrialMessageForDays(diff: Duration<true>) {
  diff = shiftToLargest(diff, ["seconds", "minutes", "hours", "days"]);
  const formattedDiff = diff.toHuman({ unitDisplay: "short" });
  return `Your trial expires in ${formattedDiff}.`;
}
