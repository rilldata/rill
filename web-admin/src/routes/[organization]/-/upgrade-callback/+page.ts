import { getCancelledIssue } from "@rilldata/web-admin/features/billing/issues/getMessageForCancelledIssue";
import { getPaymentIssues } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ parent }) => {
  const { issues } = await parent();

  return {
    paymentIssues: getPaymentIssues(issues),
    cancelled: !!getCancelledIssue(issues),
  };
};
