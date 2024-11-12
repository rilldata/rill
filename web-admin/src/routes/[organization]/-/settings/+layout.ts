import { getNeverSubscribedIssue } from "@rilldata/web-admin/features/billing/issues/getMessageForCancelledIssue";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ parent }) => {
  const { issues } = await parent();
  const neverSubscribed = !!getNeverSubscribedIssue(issues);
  return {
    neverSubscribed,
  };
};
