import {
  type V1BillingIssue,
  V1BillingIssueType,
  type V1Subscription,
} from "@rilldata/web-admin/client";
import { showUpgradeDialog } from "@rilldata/web-admin/features/billing/banner/bannerCTADialogs";
import type { BannerMessage } from "@rilldata/web-common/lib/event-bus/events";
import { DateTime } from "luxon";

const WarningPeriodInDays = 5;
const TrialGracePeriodInDays = 9;

const cta: BannerMessage["cta"] = {
  text: "Upgrade ->",
  type: "button",
  onClick: () => showUpgradeDialog.set(true),
};

export function handleTrialPlan(
  subscription: V1Subscription,
  issues: V1BillingIssue[],
): BannerMessage {
  const trialIssue = issues.find(
    (i) =>
      i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_ON_TRIAL ||
      i.type === V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_ENDED,
  );

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
  const diff = today.diff(endDate);
  if (trialIssue.type !== V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_ENDED) {
    bannerMessage.message += `${getTrialMessageForDays(diff.days)} Upgrade to maintain access.`;
    bannerMessage.type = diff.days > WarningPeriodInDays ? "info" : "warning";
  } else if (-diff.days < TrialGracePeriodInDays) {
    bannerMessage.message = `Your trial has expired. Upgrade within ${TrialGracePeriodInDays} days to maintain access.`;
    bannerMessage.type = "warning";
  } else {
    bannerMessage.message = `Your trial has expired and this org’s projects are now hibernating. Upgrade to wake projects and regain full access.`;
    bannerMessage.type = "error";
  }

  return bannerMessage;
}

export function getTrialMessageForDays(days: number) {
  switch (days) {
    case 0:
      return "Your trial expires today.";
    case 1:
      return "Your trial expires tomorrow.";
    default:
      return `Your trial expires in ${days} days.`;
  }
}
