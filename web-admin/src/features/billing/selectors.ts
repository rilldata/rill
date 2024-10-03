import { createAdminServiceGetBillingSubscription } from "@rilldata/web-admin/client";

export function getPlanForOrg(org: string, enabled = true) {
  return createAdminServiceGetBillingSubscription(org, {
    query: {
      enabled: enabled && !!org,
      select: (data) => data.subscription.plan,
    },
  });
}
