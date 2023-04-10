import { redirect } from "@sveltejs/kit";
import {
  adminServiceGetCurrentUser,
  adminServiceListOrganizations,
} from "../client";
import { ADMIN_URL } from "../client/http-client";

export async function load() {
  const user = await adminServiceGetCurrentUser();
  if (!user.user) {
    throw redirect(307, `${ADMIN_URL}/auth/login?redirect=${window.origin}`);
  }

  const orgs = await adminServiceListOrganizations();
  if (orgs.organizations.length > 0) {
    throw redirect(307, `/${orgs.organizations[0].name}`);
  }

  // No organizations. Go to "You're lonely" page.
  return;
}
