import { redirectToLogin } from "@rilldata/web-admin/client/redirect-utils";

export async function load({ parent }) {
  const { user } = await parent();

  if (!user) redirectToLogin();

  return { user };
}
