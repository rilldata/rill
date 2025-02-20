import {
  type BillingIssue,
  BillingIssueType,
} from "@rilldata/web-common/proto/gen/rill/admin/v1/api_pb";

export function getNeverSubscribedIssue(issues: BillingIssue[]) {
  return issues.find((i) => i.type === BillingIssueType.NEVER_SUBSCRIBED);
}

export function getTrialIssue(issues: BillingIssue[]) {
  return issues.find(
    (i) =>
      i.type === BillingIssueType.ON_TRIAL ||
      i.type === BillingIssueType.TRIAL_ENDED,
  );
}
