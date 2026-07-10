import type { RpcStatus } from "@rilldata/web-admin/client";
import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { error, redirect } from "@sveltejs/kit";
import { isAxiosError } from "axios";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params: { organization }, url }) => {
  let redirectUrl = `/${organization}/-/settings/billing`;
  try {
    redirectUrl = await fetchPaymentsPortalURL(
      organization,
      `${url.protocol}//${url.host}/${organization}`,
    );
  } catch (e) {
    if (!isAxiosError<RpcStatus>(e) || !e.response) {
      throw error(500, m.billing_error_redirect_payment());
    }

    throw error(e.response.status, e.response.data.message);
  }
  throw redirect(307, redirectUrl);
};
