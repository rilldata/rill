import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params: { organization }, parent }) => {
  const { billingPortalUrl } = await parent();

  if (!billingPortalUrl) {
    throw redirect(307, `/${organization}/-/settings`);
  }
};
