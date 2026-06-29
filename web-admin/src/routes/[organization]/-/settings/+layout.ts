import {
  adminServiceGetBillingSubscription,
  getAdminServiceGetBillingSubscriptionQueryKey,
  type RpcStatus,
} from "@rilldata/web-admin/client";
import { getNeverSubscribedIssue } from "@rilldata/web-admin/features/billing/issues/getMessageForCancelledIssue";
import * as m from "@rilldata/web-common/paraglide/messages.js";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import { error, redirect } from "@sveltejs/kit";
import { isAxiosError } from "axios";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ parent, params }) => {
  const { issues, organizationPermissions } = await parent();

  if (!organizationPermissions?.manageOrg) {
    throw redirect(307, `/${params.organization}`);
  }

  const neverSubscribed = !!getNeverSubscribedIssue(issues);

  const queryKey = getAdminServiceGetBillingSubscriptionQueryKey(
    params.organization,
  );
  const queryFn = () => adminServiceGetBillingSubscription(params.organization);

  try {
    const billingSubscription = await queryClient.fetchQuery({
      queryKey,
      queryFn,
    });
    return {
      subscription: billingSubscription.subscription,
      billingPortalUrl: billingSubscription.billingPortalUrl,
      neverSubscribed,
    };
  } catch (e) {
    if (!isAxiosError<RpcStatus>(e) || !e.response) {
      throw error(500, m.route_error_fetching_billing_subscription());
    }

    throw error(e.response.status, e.response.data.message);
  }
};
