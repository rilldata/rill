import { redirect } from "@sveltejs/kit";
import { ACCOUNT_TOKENS_PATH } from "@rilldata/web-admin/features/personal-access-tokens/personal-access-tokens";

export const load = () => {
  // Access tokens are the only account page so far.
  throw redirect(307, ACCOUNT_TOKENS_PATH);
};
