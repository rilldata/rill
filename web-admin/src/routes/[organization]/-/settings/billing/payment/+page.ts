import { redirectToLoginIfNotLoggedIn } from "@rilldata/web-admin/features/authentication/checkUserAccess";
import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
import { error, redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params: { organization }, url }) => {
  let redirectUrl = `/${organization}/-/settings/billing`;
  try {
    redirectUrl = await fetchPaymentsPortalURL(
      organization,
      `${url.protocol}//${url.host}/${organization}`,
    );
  } catch (e) {
    if (e.response?.status !== 403 || (await redirectToLoginIfNotLoggedIn())) {
      throw error(e.response.status, "Error redirecting to payment portal");
    }
  }
  throw redirect(307, redirectUrl);
};
